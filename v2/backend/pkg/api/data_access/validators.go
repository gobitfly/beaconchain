package dataaccess

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
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

	// active
	effectiveBalances := make(map[t.VDBValidator]uint64)
	var validatorExitEpochs []exp.Expression
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
			return nil, fmt.Errorf("validator %d has no exit epoch", validator)
		}
		// goqu can't pass tuples directly
		epochTs := utils.EpochToTime(uint64(validatorMapping.ValidatorMetadata[validator].ExitEpoch.Int64)).Unix()
		validatorExitEpochs = append(validatorExitEpochs, goqu.L("(fromUnixTimestamp(?), ?)", epochTs, validator))
	}

	// exited
	if len(validatorExitEpochs) == 0 {
		return effectiveBalances, nil
	}
	ds := goqu.Dialect("postgres").
		Select(
			// goqu.SUM(goqu.I("balance_effective_end")).As("balance_effective_end"),
			goqu.I("validator_index"),
			goqu.I("balance_effective_end"),
		).
		From("validator_dashboard_data_epoch").
		Where(
			goqu.L("(epoch_timestamp, validator_index)").In(validatorExitEpochs),
		)

	type ValidatorEB struct {
		ValidatorIndex   uint64 `db:"validator_index"`
		EffectiveBalance uint64 `db:"balance_effective_end"`
	}
	ebsBeforeExit, err := runQueryRows[[]ValidatorEB](ctx, d.clickhouseReader, ds)
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

	return runQueryRows[[]t.VDBValidator](ctx, d.alloyReader, validatorsDs)
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

	return runQueryRows[[]t.VDBValidator](ctx, d.alloyReader, validatorsDs)
}

func (d *DataAccessService) GetValidatorsByGraffiti(ctx context.Context, graffiti string) ([]t.VDBValidator, error) {
	// includes orphaned blocks
	validatorsDs := goqu.Dialect("postgres").
		From(goqu.T("blocks").As("b")).
		SelectDistinct(goqu.I("b.proposer")).
		Where(
			goqu.I("b.graffiti_text").Eq(graffiti),
		)

	return runQueryRows[[]t.VDBValidator](ctx, d.alloyReader, validatorsDs)
}
