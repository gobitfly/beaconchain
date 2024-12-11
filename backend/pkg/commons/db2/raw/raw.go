package raw

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/hexutil"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

type compressor interface {
	compress(src []byte) ([]byte, error)
	decompress(src []byte) ([]byte, error)
}

type Store struct {
	db         database.Database
	compressor compressor
}

func NewStore(store database.Database) Store {
	return Store{
		db:         store,
		compressor: gzipCompressor{},
	}
}

func (store Store) AddBlocks(blocks []FullBlockData) error {
	itemsByKey := make(map[string][]database.Item)
	for _, fullBlock := range blocks {
		if len(fullBlock.Block) == 0 || len(fullBlock.BlockTxs) != 0 && len(fullBlock.Traces) == 0 {
			return fmt.Errorf("block %d: empty data", fullBlock.BlockNumber)
		}
		key := blockKey(fullBlock.ChainID, fullBlock.BlockNumber)

		block, err := store.compressor.compress(fullBlock.Block)
		if err != nil {
			return fmt.Errorf("cannot compress block %d: %w", fullBlock.BlockNumber, err)
		}
		receipts, err := store.compressor.compress(fullBlock.Receipts)
		if err != nil {
			return fmt.Errorf("cannot compress receipts %d: %w", fullBlock.BlockNumber, err)
		}
		traces, err := store.compressor.compress(fullBlock.Traces)
		if err != nil {
			return fmt.Errorf("cannot compress traces %d: %w", fullBlock.BlockNumber, err)
		}
		itemsByKey[key] = []database.Item{
			{
				Family: BT_COLUMNFAMILY_BLOCK,
				Column: BT_COLUMN_BLOCK,
				Data:   block,
			},
			{
				Family: BT_COLUMNFAMILY_RECEIPTS,
				Column: BT_COLUMN_RECEIPTS,
				Data:   receipts,
			},
			{
				Family: BT_COLUMNFAMILY_TRACES,
				Column: BT_COLUMN_TRACES,
				Data:   traces,
			},
		}
		if len(fullBlock.Receipts) == 0 {
			// todo move that log higher up
			log.Warn(fmt.Sprintf("empty receipts at block %d lRec %d lTxs %d", fullBlock.BlockNumber, len(fullBlock.Receipts), len(fullBlock.BlockTxs)))
		}
		if fullBlock.BlockUnclesCount > 0 {
			uncles, err := store.compressor.compress(fullBlock.Uncles)
			if err != nil {
				return fmt.Errorf("cannot compress block %d: %w", fullBlock.BlockNumber, err)
			}
			itemsByKey[key] = append(itemsByKey[key], database.Item{
				Family: BT_COLUMNFAMILY_UNCLES,
				Column: BT_COLUMN_UNCLES,
				Data:   uncles,
			})
		}
	}
	return store.db.BulkAdd(itemsByKey)
}

func (store Store) ReadBlockByNumber(chainID uint64, number int64) (*FullBlockData, error) {
	return store.readBlock(chainID, number)
}

func (store Store) ReadBlockByHash(chainID uint64, hash string) (*FullBlockData, error) {
	// todo use sql store to retrieve hash
	return nil, fmt.Errorf("ReadBlockByHash not implemented")
}

func (store Store) readBlock(chainID uint64, number int64) (*FullBlockData, error) {
	key := blockKey(chainID, number)
	row, err := store.db.GetRow(key)
	if err != nil {
		return nil, err
	}
	return store.parseRow(chainID, number, row.Values)
}

func (store Store) parseRow(chainID uint64, number int64, data map[string][]byte) (*FullBlockData, error) {
	block, err := store.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_BLOCK, BT_COLUMN_BLOCK)])
	if err != nil {
		return nil, fmt.Errorf("cannot decompress block %d: %w", number, err)
	}
	receipts, err := store.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_RECEIPTS, BT_COLUMN_RECEIPTS)])
	if err != nil {
		return nil, fmt.Errorf("cannot decompress receipts %d: %w", number, err)
	}
	traces, err := store.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_TRACES, BT_COLUMN_TRACES)])
	if err != nil {
		return nil, fmt.Errorf("cannot decompress traces %d: %w", number, err)
	}
	uncles, err := store.compressor.decompress(data[fmt.Sprintf("%s:%s", BT_COLUMNFAMILY_UNCLES, BT_COLUMN_UNCLES)])
	if err != nil {
		return nil, fmt.Errorf("cannot decompress uncles %d: %w", number, err)
	}
	return &FullBlockData{
		ChainID:          chainID,
		BlockNumber:      number,
		BlockHash:        nil,
		BlockUnclesCount: 0,
		BlockTxs:         nil,
		Block:            block,
		Receipts:         receipts,
		Traces:           traces,
		Uncles:           uncles,
	}, nil
}

func (store Store) ReadBlocksByNumber(chainID uint64, start, end int64) ([]*FullBlockData, error) {
	rows, err := store.db.GetRowsRange(blockKey(chainID, start), blockKey(chainID, end))
	if err != nil {
		return nil, err
	}
	blocks := make([]*FullBlockData, 0, end-start+1)
	for _, row := range rows {
		block, err := store.parseRow(chainID, blockKeyToNumber(chainID, row.Key), row.Values)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func blockKey(chainID uint64, number int64) string {
	return fmt.Sprintf("%d:%12d", chainID, MAX_EL_BLOCK_NUMBER-number)
}

func blockKeyToNumber(chainID uint64, key string) int64 {
	key = strings.TrimPrefix(key, fmt.Sprintf("%d:", chainID))
	reversed, _ := new(big.Int).SetString(key, 10)

	return MAX_EL_BLOCK_NUMBER - reversed.Int64()
}

type FullBlockData struct {
	ChainID uint64

	BlockNumber      int64
	BlockHash        hexutil.Bytes
	BlockUnclesCount int
	BlockTxs         []string

	Block    hexutil.Bytes
	Receipts hexutil.Bytes
	Traces   hexutil.Bytes
	Uncles   hexutil.Bytes
}
