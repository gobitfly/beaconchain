package dataaccess

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/ext"
	"github.com/doug-martin/goqu/v9"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
)

type ValidatorRepository interface {
	GetValidatorsEffectiveBalances(ctx context.Context, validators []t.VDBValidator, onlyActive bool) (map[t.VDBValidator]uint64, error)

	GetValidatorsByDepositAddress(ctx context.Context, depositAddress string) ([]t.VDBValidator, error)
	GetValidatorsByWithdrawalCredentials(ctx context.Context, withdrawalCredentials string) ([]t.VDBValidator, error)
	GetValidatorsByGraffiti(ctx context.Context, graffiti string) ([]t.VDBValidator, error)
}

// returns the effective balances of the provided validators
// if onlyActive = true: executed from the vdb premium limits pov, i.e. exited validators account for the EB at exit time
func (d *DataAccessService) GetValidatorsEffectiveBalances(ctx context.Context, validators []t.VDBValidator, onlyActive bool) (map[t.VDBValidator]uint64, error) {
	validatorMapping, err := d.services.GetCurrentValidatorMapping()
	if err != nil {
		return nil, err
	}
	var validatorTable *ext.Table

	// active
	effectiveBalances := make(map[t.VDBValidator]uint64)
	for _, validator := range validators {
		if len(validatorMapping.ValidatorMetadata) <= int(validator) {
			return nil, fmt.Errorf("validator index %d not found in validator mapping", validator)
		}
		status := constypes.ValidatorDbStatus(validatorMapping.ValidatorMetadata[validator].Status)
		balanceWithdrawn := !onlyActive &&
			(status == constypes.DbSlashed || status == constypes.DbExited) &&
			validatorMapping.ValidatorMetadata[validator].EffectiveBalance == 0
		if !balanceWithdrawn {
			effectiveBalances[validator] = validatorMapping.ValidatorMetadata[validator].EffectiveBalance
			continue
		}
		// exited & balance withdrawn, need to query latest EB before exit
		if !validatorMapping.ValidatorMetadata[validator].ExitEpoch.Valid {
			log.Warnf("validator %d has no exit epoch", validator)
		}
		if validatorTable == nil {
			validatorTable, err = ext.NewTable("exited_validators",
				ext.Column("validator_index", "UInt64"),
			)
			if err != nil {
				return nil, err
			}
		}
		if err = validatorTable.Append(validator); err != nil {
			return nil, err
		}
	}
	if validatorTable == nil {
		return effectiveBalances, nil
	}

	// exited
	ds := goqu.Dialect("postgres").
		Select(
			goqu.I("validator_index"),
			goqu.I("balance_effective_end"),
		).
		From("validator_dashboard_effective_balance_lookup").
		Where(
			goqu.L("validator_index").In(goqu.L("SELECT * FROM exited_validators")),
		)

	type ValidatorEB struct {
		ValidatorIndex   uint64 `db:"validator_index"`
		EffectiveBalance uint64 `db:"balance_effective_end"`
	}
	ebsBeforeExit, err := runQueryRows[[]ValidatorEB](clickhouse.Context(ctx, clickhouse.WithExternalTable(validatorTable)), d.clickhouseReader, ds)
	if err != nil {
		return nil, err
	}
	for _, eb := range ebsBeforeExit {
		effectiveBalances[eb.ValidatorIndex] = eb.EffectiveBalance
	}
	return effectiveBalances, nil
}

func (d *DataAccessService) GetValidatorsByDepositAddress(ctx context.Context, depositAddress string) ([]t.VDBValidator, error) {
	addressParsed, err := hex.DecodeString(strings.TrimPrefix(depositAddress, "0x"))
	if err != nil {
		return nil, err
	}

	validatorsDs := goqu.Dialect("postgres").
		From(goqu.T("validators").As("v")).
		SelectDistinct(goqu.I("v.validatorindex")).
		InnerJoin(
			goqu.T("eth1_deposits").As("d"),
			goqu.On(goqu.I("v.pubkey").Eq(goqu.I("d.publickey"))),
		).
		Where(
			goqu.I("d.from_address").Eq(addressParsed),
		)

	return runQueryRows[[]t.VDBValidator](ctx, d.readerDb, validatorsDs)
}

func (d *DataAccessService) GetValidatorsByWithdrawalCredentials(ctx context.Context, withdrawalCredentials string) ([]t.VDBValidator, error) {
	addressParsed, err := hex.DecodeString(strings.TrimPrefix(withdrawalCredentials, "0x"))
	if err != nil {
		return nil, err
	}

	validatorsDs := goqu.Dialect("postgres").
		From(goqu.T("validators").As("v")).
		SelectDistinct(goqu.I("v.validatorindex")).
		Where(
			goqu.I("v.withdrawalcredentials").Eq(addressParsed),
		)

	return runQueryRows[[]t.VDBValidator](ctx, d.readerDb, validatorsDs)
}

func (d *DataAccessService) GetValidatorsByGraffiti(ctx context.Context, graffiti string) ([]t.VDBValidator, error) {
	// includes orphaned blocks
	validatorsDs := goqu.Dialect("postgres").
		From(goqu.T("blocks").As("b")).
		SelectDistinct(goqu.I("b.proposer")).
		Where(
			goqu.I("b.graffiti_text").Eq(graffiti),
		)

	return runQueryRows[[]t.VDBValidator](ctx, d.readerDb, validatorsDs)
}

// estimate activation for pending and deposited validators
// TODO support estimates for validators without index
func (d *DataAccessService) getValidatorActivation(ctx context.Context, validator uint64) (uint64, error) {
	validatorMapping, err := d.services.GetCurrentValidatorMapping()
	if err != nil {
		return 0, err
	}
	if validator >= uint64(len(validatorMapping.ValidatorMetadata)) {
		return 0, fmt.Errorf("validator index %d not found in validator mapping", validator)
	}
	metadata := validatorMapping.ValidatorMetadata[validator]
	if metadata.ActivationEpoch.Valid {
		return uint64(metadata.ActivationEpoch.Int64), nil
	}

	// estimate
	latestEpoch := cache.LatestFinalizedEpoch.Get()
	if d.config.ClConfig.ElectraForkEpoch > latestEpoch {
		if constypes.ValidatorDbStatus(metadata.Status) != constypes.DbPending {
			// probably not enough deposits yet
			return 0, fmt.Errorf("validator %d is not pending activation yet", validator)
		}
		if !metadata.Queues.ActivationIndex.Valid {
			return 0, fmt.Errorf("validator %d has no activation index", validator)
		}
		queuePosition := uint64(metadata.Queues.ActivationIndex.Int64)

		latestStats := cache.LatestStats.Get()
		activationChurnRate := uint64(4)
		if latestStats.ValidatorActivationChurnLimit == nil {
			log.Warnf("Activation Churn rate not set in config, using 4 as default")
		} else {
			activationChurnRate = *latestStats.ValidatorActivationChurnLimit
		}

		epochsToWait := (queuePosition - 1) / activationChurnRate
		// calculate dequeue epoch
		estimatedActivationEpoch := latestEpoch + epochsToWait + 1
		// add activation offset
		estimatedActivationEpoch += utils.Config.Chain.ClConfig.MaxSeedLookahead + 1

		return estimatedActivationEpoch, nil
	}

	// post pectra
	if constypes.ValidatorDbStatus(metadata.Status) != constypes.DbDeposited {
		// should not happen since there's no more activation queue after deposits have been processed (see process_registry_updates)
		return 0, fmt.Errorf("validator %d has no activation epoch", validator)
	}
	// determine sum of previous deposit
	// check if there's a pending deposit which pushes above min activation
	// return estimate of that from db

	// could also simulate in db, but wouldn't be pretty
	ds := goqu.Dialect("postgres").From("pending_deposits_queue").
		Select(
			goqu.I("est_clear_epoch"),
			goqu.I("amount"),
		).
		Where(
			goqu.I("validator_index").Eq(validator),
		).
		Order(goqu.I("id").Asc())

	type dbResult struct {
		ClearEpoch uint64 `db:"est_clear_epoch"`
		Amount     uint64 `db:"amount"`
	}
	results, err := runQueryRows[[]dbResult](ctx, d.alloyReader, ds)
	if err != nil {
		return 0, err
	}

	effectiveBalance := metadata.EffectiveBalance
	balance := metadata.Balance
	upwardThreshold := utils.Config.Chain.ClConfig.EffectiveBalanceIncrement / utils.Config.Chain.ClConfig.HysteresisQuotient * utils.Config.Chain.ClConfig.HysteresisUpwardMultiplier
	for _, deposit := range results {
		balance += deposit.Amount
		if effectiveBalance+upwardThreshold < balance {
			effectiveBalance = balance - balance%utils.Config.Chain.ClConfig.EffectiveBalanceIncrement
			if effectiveBalance >= utils.Config.Chain.ClConfig.MinActivationBalance {
				return deposit.ClearEpoch, nil
			}
		}
	}

	return 0, fmt.Errorf("validator %d has not enough pending ETH deposits", validator)
}
