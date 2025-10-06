package db

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/google/uuid"
)

type LegacyValidatorBalanceTable struct {
	ValidatorIndex   []uint32     `db:"validator_index"`
	EpochTimestamp   []*time.Time `db:"epoch_timestamp"`
	Balance          []uint64     `db:"balance"`
	EffectiveBalance []uint64     `db:"effective_balance"`
}

func (c LegacyValidatorBalanceTable) Get(str string) any {
	switch str {
	case "validator_index":
		return c.ValidatorIndex
	case "epoch_timestamp":
		return c.EpochTimestamp
	case "balance":
		return c.Balance
	case "effective_balance":
		return c.EffectiveBalance
	default:
		return nil
	}
}
func (c LegacyValidatorBalanceTable) Extend(cOther UltraFastClickhouseStruct) error {
	return nil
}

func SaveLegacyValidatorBalancesToClickhouse(epoch uint64, state []*types.Validator) error {
	epochTs := utils.EpochToTime(epoch)

	epochData := LegacyValidatorBalanceTable{}
	for _, validator := range state {
		epochData.ValidatorIndex = append(epochData.ValidatorIndex, uint32(validator.Index))
		epochData.EpochTimestamp = append(epochData.EpochTimestamp, &epochTs)
		epochData.Balance = append(epochData.Balance, validator.Balance)
		epochData.EffectiveBalance = append(epochData.EffectiveBalance, validator.EffectiveBalance)
	}

	err := UltraFastDumpToClickhouse(epochData, "legacy_validator_epoch_balances", uuid.New().String())
	if err != nil {
		return fmt.Errorf("error writing legacy validator epoch balances to clickhouse: %w", err)
	}

	return nil
}
