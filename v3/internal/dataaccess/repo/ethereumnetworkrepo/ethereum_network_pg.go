package ethereumnetworkrepo

import (
	"context"
	"errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

type DBRepository struct {
	roChainDb *sqlx.DB
}

func (r *DBRepository) Initialize(roChainDb *sqlx.DB) {
	r.roChainDb = roChainDb
}

func (r *DBRepository) selectDatabase(chain domain.Chain) *sqlx.DB {
	if chain == domain.ChainMainnet {
		return r.roChainDb
	}
	// TODO: add chains once supported
	return r.roChainDb
}

type getSlotResult struct {
	AttestationCount         int    `db:"attestationscount"`
	AttestationSlashingCount int    `db:"attesterslashingscount"`
	ProposerSlashingCount    int    `db:"proposerslashingscount"`
	BlockRoot                []byte `db:"blockroot"`
	Graffiti                 []byte `db:"graffiti"`
	Proposer                 int    `db:"proposer"`
	ProposerPubkey           []byte `db:"proposer_pubkey"`
	Status                   string `db:"status"`
	Finalized                bool   `db:"finalized"`

	// Events (aggregated)
	ConsolidationAggrQueuedCount     int `db:"ca_queued_consolidation_count"`
	ConsolidationAggrQueuedAmount    int `db:"ca_queued_consolidation_amount"`
	ConsolidationAggrProcessedCount  int `db:"ca_processed_consolidation_count"`
	ConsolidationAggrProcessedAmount int `db:"ca_processed_consolidation_amount"`

	DepositAggrQueuedCount     int `db:"da_queued_deposit_count"`
	DepositAggrQueuedAmount    int `db:"da_queued_deposit_amount"`
	DepositAggrProcessedCount  int `db:"da_processed_deposit_count"`
	DepositAggrProcessedAmount int `db:"da_processed_deposit_amount"`

	ManualWithdrawalsAggrProcessedCount  int `db:"mwa_processed_manual_withdrawn_count"`
	ManualWithdrawalsAggrProcessedAmount int `db:"mwa_processed_manual_withdrawn_amount"`
	ManualWithdrawalsAggrQueuedCount     int `db:"mwa_queued_manual_withdrawn_count"`
	ManualWithdrawalsAggrQueuedAmount    int `db:"mwa_queued_manual_withdrawn_amount"`

	AutoWithdrawalsAggrProcessedCount  int `db:"awa_processed_auto_withdrawn_count"`
	AutoWithdrawalsAggrProcessedAmount int `db:"awa_processed_auto_withdrawn_amount"`
}

func (r *DBRepository) GetSlot(ctx context.Context, chain domain.Chain, slot int) (domain.Slot, error) {
	query := goqu.Dialect("postgres").
		From(goqu.T("blocks").As("b")).
		With("ca_queued", consolidationAggregationCTE(slot, SlotEventQueued)).
		With("ca_processed", consolidationAggregationCTE(slot, SlotEventProcessed)).
		With("da_queued", depositAggregationCTE(slot, SlotEventQueued)).
		With("da_processed", depositAggregationCTE(slot, SlotEventProcessed)).
		With("mwa_queued", manualWithdrawalsAggregationCTE(slot, SlotEventQueued)).
		With("mwa_processed", manualWithdrawalsAggregationCTE(slot, SlotEventProcessed)).
		With("cwa_processed", combinedWithdrawalsAggregationCTE(slot)).
		Select(
			// basic
			goqu.I("b.slot"),
			goqu.I("b.attestationscount"),
			goqu.I("b.attesterslashingscount"),
			goqu.I("b.proposerslashingscount"),
			goqu.I("b.blockroot"),
			goqu.I("b.graffiti"),
			goqu.I("b.proposer"),
			goqu.I("validators.pubkey").As("proposer_pubkey"),
			goqu.I("b.status"),
			goqu.I("b.finalized"),

			// consolidation aggregation
			goqu.I("ca_queued.consolidation_count").As("ca_queued_consolidation_count"),
			goqu.I("ca_queued.consolidation_amount").As("ca_queued_consolidation_amount"),
			goqu.I("ca_processed.consolidation_count").As("ca_processed_consolidation_count"),
			goqu.I("ca_processed.consolidation_amount").As("ca_processed_consolidation_amount"),

			// deposit aggregation
			goqu.I("da_queued.deposit_count").As("da_queued_deposit_count"),
			goqu.I("da_queued.deposit_amount").As("da_queued_deposit_amount"),
			goqu.I("da_processed.deposit_count").As("da_processed_deposit_count"),
			goqu.I("da_processed.deposit_amount").As("da_processed_deposit_amount"),

			// Manual Withdrawals aggregation
			goqu.I("mwa_processed.manual_withdrawn_count").As("mwa_processed_manual_withdrawn_count"),
			goqu.I("mwa_processed.manual_withdrawn_amount").As("mwa_processed_manual_withdrawn_amount"),
			goqu.I("mwa_queued.manual_withdrawn_count").As("mwa_queued_manual_withdrawn_count"),
			goqu.I("mwa_queued.manual_withdrawn_amount").As("mwa_queued_manual_withdrawn_amount"),

			// Auto Withdrawals aggregation [combined - manual] (only processed)
			goqu.L("cwa_processed.combined_withdrawn_count - mwa_processed.manual_withdrawn_count").As("awa_processed_auto_withdrawn_count"),
			goqu.L("cwa_processed.combined_withdrawn_amount - mwa_processed.manual_withdrawn_amount").As("awa_processed_auto_withdrawn_amount"),
		).
		Where(goqu.I("slot").Eq(slot)).

		// Consolidation
		Join(goqu.T("ca_queued"), goqu.On(goqu.L("true"))).
		Join(goqu.T("ca_processed"), goqu.On(goqu.L("true"))).

		// Deposit
		Join(goqu.T("da_processed"), goqu.On(goqu.L("true"))).
		Join(goqu.T("da_queued"), goqu.On(goqu.L("true"))).

		// Manual Withdrawals
		Join(goqu.T("mwa_processed"), goqu.On(goqu.L("true"))).
		Join(goqu.T("mwa_queued"), goqu.On(goqu.L("true"))).

		// Combined Withdrawals used to get auto withdrawals (only processed)
		Join(goqu.T("cwa_processed"), goqu.On(goqu.L("true"))).

		// Validator
		Join(goqu.T("validators"), goqu.On(goqu.L("validators.validatorindex = b.proposer")))

	result, err := repo.RunQuery[getSlotResult](ctx, r.selectDatabase(chain), query)
	if err != nil {
		return domain.Slot{}, err
	}

	return transformSlotToDomain(slot, result), nil
}

func transformSlotToDomain(slot int, result getSlotResult) domain.Slot {
	return domain.Slot{
		Slot:                     slot,
		AttestationSlashingCount: result.AttestationSlashingCount,
		ProposerSlashingCount:    result.ProposerSlashingCount,
		AttestationCount:         result.AttestationCount,
		BlockRoot:                result.BlockRoot,
		Graffiti:                 result.Graffiti,
		Finalized:                result.Finalized,
		Proposer: domain.Validator{
			Index:  result.Proposer,
			Pubkey: result.ProposerPubkey,
		},
		Status: domain.DutyStatus(result.Status),
		ProcessedAutoWithdrawals: domain.ClEventDetails{
			Count:  result.AutoWithdrawalsAggrProcessedCount,
			Amount: result.AutoWithdrawalsAggrProcessedAmount,
		},
		ProcessedConsolidations: domain.ClEventDetails{
			Count:  result.ConsolidationAggrProcessedCount,
			Amount: result.ConsolidationAggrProcessedAmount,
		},
		ProcessedDeposits: domain.ClEventDetails{
			Count:  result.DepositAggrProcessedCount,
			Amount: result.DepositAggrProcessedAmount,
		},
		ProcessedManualWithdrawals: domain.ClEventDetails{
			Count:  result.ManualWithdrawalsAggrProcessedCount,
			Amount: result.ManualWithdrawalsAggrProcessedAmount,
		},
		QueuedConsolidations: domain.ClEventDetails{
			Count:  result.ConsolidationAggrQueuedCount,
			Amount: result.ConsolidationAggrQueuedAmount,
		},
		QueuedDeposits: domain.ClEventDetails{
			Count:  result.DepositAggrQueuedCount,
			Amount: result.DepositAggrQueuedAmount,
		},
		QueuedWithdrawals: domain.ClEventDetails{
			Count:  result.ManualWithdrawalsAggrQueuedCount,
			Amount: result.ManualWithdrawalsAggrQueuedAmount,
		},
	}
}

type SlotEventStatus string

const (
	SlotEventQueued    SlotEventStatus = "slot_queued"
	SlotEventProcessed SlotEventStatus = "slot_processed"
)

func consolidationAggregationCTE(slot int, status SlotEventStatus) *goqu.SelectDataset {
	query := goqu.Dialect("postgres").
		From(goqu.T("blocks_consolidation_requests_v2")).
		Select(
			goqu.COALESCE(goqu.SUM(goqu.I("amount_consolidated")), 0).As("consolidation_amount"),
			goqu.COALESCE(goqu.COUNT(goqu.I("id")), 0).As("consolidation_count"),
		)
	query = query.Where(goqu.I(string(status)).Eq(slot))
	return query
}

func depositAggregationCTE(slot int, status SlotEventStatus) *goqu.SelectDataset {
	query := goqu.Dialect("postgres").
		From(goqu.T("blocks_deposit_requests_v2")).
		Select(
			goqu.COALESCE(goqu.SUM(goqu.I("amount")), 0).As("deposit_amount"),
			goqu.COALESCE(goqu.COUNT(goqu.I("amount")), 0).As("deposit_count"),
		)
	query = query.Where(goqu.I(string(status)).Eq(slot))
	return query
}

func manualWithdrawalsAggregationCTE(slot int, status SlotEventStatus) *goqu.SelectDataset {
	query := goqu.Dialect("postgres").
		From(goqu.T("blocks_withdrawal_requests_v2")).
		Select(
			goqu.COALESCE(goqu.SUM(goqu.I("amount")), 0).As("manual_withdrawn_amount"),
			goqu.COALESCE(goqu.COUNT(goqu.I("amount")), 0).As("manual_withdrawn_count"),
		)
	query = query.Where(goqu.I(string(status)).Eq(slot))
	return query
}

func combinedWithdrawalsAggregationCTE(slot int) *goqu.SelectDataset { // only processed
	query := goqu.Dialect("postgres").
		From(goqu.T("blocks_withdrawals")).
		Select(
			goqu.COALESCE(goqu.SUM(goqu.I("amount")), 0).As("combined_withdrawn_amount"),
			goqu.COALESCE(goqu.COUNT(goqu.I("amount")), 0).As("combined_withdrawn_count"),
		)
	query = query.Where(goqu.I("block_slot").Eq(slot))
	return query
}

func (r *DBRepository) GetLatestState(ctx context.Context, chain domain.Chain, view domain.ConsensusView) (domain.LatestState, error) {
	maxSlotDs := goqu.Dialect("postgres").From("blocks").Select(goqu.MAX(goqu.I("slot")))
	maxEpochDs := goqu.Dialect("postgres").From("epochs").Select(goqu.MAX(goqu.I("epoch")))

	switch view {
	case domain.ConsensusViewFinalized:
		maxSlotDs = maxSlotDs.Where(goqu.I("finalized").Eq(true))
		maxEpochDs = maxEpochDs.Where(goqu.I("finalized").Eq(true))
	case domain.ConsensusViewJustified:
		return domain.LatestState{}, errors.New("not yet supported")
	}

	ds := goqu.Dialect("postgres").
		Select(
			maxSlotDs.As("slot"),
			maxEpochDs.As("epoch"),
		)
	result, err := repo.RunQuery[domain.LatestState](ctx, r.selectDatabase(chain), ds)
	if err != nil {
		return domain.LatestState{}, err
	}

	result.ConsensusView = view
	return result, nil
}
