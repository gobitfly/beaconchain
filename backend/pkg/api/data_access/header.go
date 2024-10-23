package dataaccess

import (
	"context"
	"database/sql"
	"fmt"

	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/price"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

func (d *DataAccessService) GetLatestSlot() (uint64, error) {
	latestSlot := cache.LatestSlot.Get()
	return latestSlot, nil
}

func (d *DataAccessService) GetLatestFinalizedEpoch() (uint64, error) {
	finalizedEpoch := cache.LatestFinalizedEpoch.Get()
	return finalizedEpoch, nil
}

func (d *DataAccessService) GetLatestBlock() (uint64, error) {
	// @DATA-ACCESS implement
	return d.dummy.GetLatestBlock()
}

func (d *DataAccessService) GetBlockHeightAt(slot uint64) (uint64, error) {
	// @DATA-ACCESS implement; return error if no block at slot
	return d.dummy.GetBlockHeightAt(slot)
}

// returns the block number of the latest existing block at or before the given slot
func (d *DataAccessService) GetLatestBlockHeightForSlot(ctx context.Context, slot uint64) (uint64, error) {
	query := `SELECT MAX(exec_block_number) FROM blocks WHERE slot <= $1`
	res := uint64(0)
	err := d.alloyReader.GetContext(ctx, &res, query, slot)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnf("no EL block found at or before slot %d", slot)
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get latest existing block height at or before slot %d: %w", slot, err)
	}
	return res, nil
}

func (d *DataAccessService) GetLatestBlockHeightsForEpoch(ctx context.Context, epoch uint64) ([]uint64, error) {
	// use 2 epochs as safety margin
	query := `
	WITH recent_blocks AS (
		SELECT slot, exec_block_number
		FROM blocks
		WHERE slot < $1
		ORDER BY slot DESC
		LIMIT $2 * 2
	)
	SELECT MAX(exec_block_number) OVER (ORDER BY slot) AS block
	FROM recent_blocks
	ORDER BY slot DESC
	LIMIT $2`
	res := []uint64{}
	err := d.alloyReader.SelectContext(ctx, &res, query, (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest existing block heights for slots in epoch %d: %w", epoch, err)
	}
	return res, nil
}

func (d *DataAccessService) GetLatestExchangeRates() ([]t.EthConversionRate, error) {
	result := []t.EthConversionRate{}

	availableCurrencies := price.GetAvailableCurrencies()
	for _, code := range availableCurrencies {
		if code == "ETH" {
			// Don't return ETH/ETH info
			continue
		}
		rate := price.GetPrice("ETH", code)
		result = append(result, t.EthConversionRate{
			Currency: price.GetCurrencyLabel(code),
			Code:     code,
			Symbol:   price.GetCurrencySymbol(code),
			Rate:     rate,
		})
	}

	return result, nil
}
