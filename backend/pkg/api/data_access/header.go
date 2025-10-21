package dataaccess

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/ethereum/go-ethereum/common/hexutil"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/price"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

func (d *DataAccessService) GetLatestSlot(ctx context.Context) (uint64, error) {
	latestSlot := cache.LatestSlot.Get()
	return latestSlot, nil
}

func (d *DataAccessService) GetLatestFinalizedEpoch(ctx context.Context) (uint64, error) {
	finalizedEpoch := cache.LatestFinalizedEpoch.Get()
	return finalizedEpoch, nil
}

func (d *DataAccessService) GetLatestBlock(ctx context.Context) (uint64, error) {
	ds := goqu.Dialect("postgres").
		From(goqu.T("blocks")).
		Select(goqu.MAX(goqu.C("exec_block_number")))

	res, err := runQuery[uint64](ctx, d.readerDb, ds)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn("no EL block found")
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get latest existing block height: %w", err)
	}
	return res, nil
}

func (d *DataAccessService) GetLatestTransaction(ctx context.Context) (t.Hash, error) {
	indexedBlock, err := d.bigtable.GetMostRecentBlockFromDataTable()
	if err != nil {
		return "", fmt.Errorf("failed to get latest block: %w", err)
	}
	block, err := d.bigtable.GetBlockFromBlocksTable(indexedBlock.GetNumber())
	if err != nil {
		if err == db.ErrBlockNotFound {
			err = db.ErrNotFound
		}
		return "", fmt.Errorf("failed to get latest block from bigtable: %w", err)
	}
	transactions := block.GetTransactions()
	hash := transactions[len(transactions)-1].GetHash()
	return t.Hash(hexutil.Encode(hash)), nil
}

func (d *DataAccessService) GetBlockHeightAt(ctx context.Context, slot uint64) (uint64, error) {
	// @DATA-ACCESS implement; return error if no block at slot
	return getDummyData[uint64](ctx)
}

// returns the block number of the latest existing block at or before the given slot
func (d *DataAccessService) GetLatestBlockHeightForSlot(ctx context.Context, slot uint64) (uint64, error) {
	query := `SELECT MAX(exec_block_number) FROM blocks WHERE slot <= $1`
	res := uint64(0)
	err := d.readerDb.GetContext(ctx, &res, query, slot)
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
	err := d.readerDb.SelectContext(ctx, &res, query, (epoch+1)*utils.Config.Chain.ClConfig.SlotsPerEpoch, utils.Config.Chain.ClConfig.SlotsPerEpoch)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest existing block heights for slots in epoch %d: %w", epoch, err)
	}
	return res, nil
}

func (d *DataAccessService) GetLatestExchangeRates(ctx context.Context) ([]t.EthConversionRate, error) {
	result := []t.EthConversionRate{}

	availableCurrencies := utils.Deduplicate(price.GetAvailableCurrencies())
	for _, code := range availableCurrencies {
		rate := price.GetPrice(d.config.Frontend.MainCurrency, code)
		result = append(result, t.EthConversionRate{
			Currency: price.GetCurrencyLabel(code),
			Code:     code,
			Symbol:   price.GetCurrencySymbol(code),
			Rate:     rate,
		})
	}

	return result, nil
}
