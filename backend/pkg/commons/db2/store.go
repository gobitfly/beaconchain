package db2

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type StoreV1 struct {
	data     database.Database
	metadata database.Database
	updates  database.Database
	blocks   database.Database
	cache    database.RemoteCache
}

func NewStoreV1(data, metadata, updates, blocks database.Database, cache database.RemoteCache) StoreV1 {
	return StoreV1{
		data:     data,
		metadata: metadata,
		updates:  updates,
		blocks:   blocks,
		cache:    cache,
	}
}

func NewStoreV1FromBigtable(bigtable *database.BigTable, cache database.RemoteCache) StoreV1 {
	return StoreV1{
		data:     database.Wrap(bigtable, DataTable),
		metadata: database.Wrap(bigtable, MetadataTable),
		updates:  database.Wrap(bigtable, UpdatesTable),
		blocks:   database.Wrap(bigtable, BlocksTable),
		cache:    cache,
	}
}

func (store StoreV1) AddIndexedBlock(block IndexedBlock) error {
	updates, err := store.addIndexedBlockInData(block)
	if err != nil {
		return err
	}
	if err := store.addIndexedBlockInUpdates(block, updates); err != nil {
		return err
	}
	if err := store.addIndexedBlockInMetadata(block); err != nil {
		return err
	}
	return nil
}

func (store StoreV1) addIndexedBlockInData(block IndexedBlock) ([]string, error) {
	updates := make(map[string][]database.Item)

	update, err := transactionsToItems(block.ChainID, block.Transactions)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = blockToItems(block.ChainID, block.Block)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = internalsToItems(block.ChainID, block.Internals)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = erc20TransfersToItems(block.ChainID, block.ERC20Transfer)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = erc1155TransfersToItems(block.ChainID, block.ERC1155Transfer)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = erc721TransfersToItems(block.ChainID, block.ERC721Transfer)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = blobToItems(block.ChainID, block.Blobs)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = unclesToItems(block.ChainID, block.Uncles)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = withdrawalToItems(block.ChainID, block.Withdrawals)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	maps.Copy(updates, ensToItems(block.ChainID, block.ENS))

	if err := store.data.BulkAdd(updates); err != nil {
		return nil, err
	}

	return maps.Keys(updates), nil
}

func (store StoreV1) addIndexedBlockInUpdates(block IndexedBlock, keys []string) error {
	updates := make(map[string][]database.Item)
	if len(keys) > 0 {
		metaKeys := strings.Join(keys, ",") // save block keys in order to be able to handle chain reorgs
		updates = blockKeysMutation(block.ChainID, block.Number, block.Hash, metaKeys)
	}

	// mark sender and recipient for balance update
	for _, transaction := range block.Transactions {
		mergeItems(updates, markBalanceUpdate(block.ChainID, transaction.From, []byte{0x0}, store.cache))
		mergeItems(updates, markBalanceUpdate(block.ChainID, transaction.To, []byte{0x0}, store.cache))
	}
	for _, transaction := range block.Internals {
		mergeItems(updates, markBalanceUpdate(block.ChainID, transaction.Indexed.From, []byte{0x0}, store.cache))
		mergeItems(updates, markBalanceUpdate(block.ChainID, transaction.Indexed.To, []byte{0x0}, store.cache))
	}
	for _, transaction := range block.ERC20Transfer {
		mergeItems(updates, markBalanceUpdate(block.ChainID, transaction.Indexed.From, transaction.Indexed.TokenAddress, store.cache))
		mergeItems(updates, markBalanceUpdate(block.ChainID, transaction.Indexed.To, transaction.Indexed.TokenAddress, store.cache))
	}
	for _, transaction := range block.Blobs {
		mergeItems(updates, markBalanceUpdate(block.ChainID, transaction.Indexed.From, []byte{0x0}, store.cache))
		mergeItems(updates, markBalanceUpdate(block.ChainID, transaction.Indexed.To, []byte{0x0}, store.cache))
	}
	for _, transaction := range block.Withdrawals {
		mergeItems(updates, markBalanceUpdate(block.ChainID, transaction.Address, []byte{0x0}, store.cache))
	}
	if block.Block != nil {
		mergeItems(updates, markBalanceUpdate(block.ChainID, block.Block.Coinbase, []byte{0x0}, store.cache))
	}
	for _, uncle := range block.Uncles {
		mergeItems(updates, markBalanceUpdate(block.ChainID, uncle.Coinbase, []byte{0x0}, store.cache))
	}

	if err := store.updates.BulkAdd(updates); err != nil {
		return err
	}
	return nil
}

func (store StoreV1) addIndexedBlockInMetadata(block IndexedBlock) error {
	items := make(map[string][]database.Item)
	for _, update := range block.Contracts {
		b, err := proto.Marshal(update.Indexed)
		if err != nil {
			return err
		}

		key := fmt.Sprintf("%s:S:%x", block.ChainID, update.Address)
		ts, err := encodeIsContractUpdateTs(block.Number, uint64(update.TxIndex), uint64(update.InternalIndex))
		if err != nil {
			return fmt.Errorf("error generating bigtable isContract timestamp: %w", err)
		}
		items[key] = []database.Item{
			{
				Family:    accountFamily,
				Column:    accountIsContractColumn,
				Data:      b,
				Timestamp: &ts,
			},
		}
	}
	if err := store.metadata.BulkAdd(items); err != nil {
		return err
	}
	return nil
}

func (store StoreV1) GetPairsToUpdate(chainID string, batchSize int64) ([]Pair, error) {
	key := fmt.Sprintf("%s:%s", chainID, balanceKey)
	rows, err := store.updates.GetRowsRange(toSuccessor(key), fmt.Sprintf("%s:", key), database.WithLimit(batchSize))
	if err != nil {
		return nil, err
	}
	var pairs []Pair
	for _, row := range rows {
		for col := range row.Values {
			pairs = append(pairs, Pair{
				Address: addressFromKey(row.Key),
				Token:   tokenFromColumn(col),
			})
		}
	}
	return pairs, nil
}

func (store StoreV1) DeletePairs(chainID string, pairs []Pair) error {
	var keys []string
	for _, pair := range pairs {
		keys = append(keys, fmt.Sprintf("%s:%s:%x", chainID, balanceKey, pair.Address.Bytes()))
	}
	slices.Sort(keys)
	keys = slices.Compact(keys)
	return store.updates.DeleteRowsWithKeys(keys)
}

func (store StoreV1) UpdateBalance(chainID string, balances []Balance) error {
	updates := make(map[string][]database.Item)
	for _, balance := range balances {
		key := fmt.Sprintf("%s:%x", chainID, balance.Address)
		updates[key] = append(updates[key], database.Item{
			Family: accountFamily,
			Column: fmt.Sprintf("B:%x", balance.Token),
			Data:   balance.Value.Bytes(),
		})
	}
	return store.metadata.BulkAdd(updates)
}

func (store StoreV1) UpdateToken(chainID string, tokens []*types.ERC20TokenPrice) error {
	updates := make(map[string][]database.Item)
	for _, token := range tokens {
		key := fmt.Sprintf("%s:%x", chainID, token.Token)
		updates[key] = append(updates[key], database.Item{
			Family: erc20MetadataFamily,
			Column: erc20ColumnPrice,
			Data:   token.Price,
		})
		updates[key] = append(updates[key], database.Item{
			Family: erc20MetadataFamily,
			Column: erc20ColumnTotalSupply,
			Data:   token.TotalSupply,
		})
	}
	return store.metadata.BulkAdd(updates)
}

func (store StoreV1) TokenPrice(chainID string, token common.Address) (*types.ERC20TokenPrice, error) {
	key := fmt.Sprintf("%s:%x", chainID, token.Bytes())
	row, err := store.metadata.GetRow(key)
	if err != nil {
		return nil, err
	}
	return &types.ERC20TokenPrice{
		Token:       token.Bytes(),
		Price:       row.Values[fmt.Sprintf("%s:%s", erc20MetadataFamily, erc20ColumnPrice)],
		TotalSupply: row.Values[fmt.Sprintf("%s:%s", erc20MetadataFamily, erc20ColumnTotalSupply)],
	}, nil
}

func (store StoreV1) SaveBlock(chainID string, block *types.Eth1Block) error {
	b, err := proto.Marshal(block)
	if err != nil {
		return err
	}

	if err := store.blocks.BulkAdd(map[string][]database.Item{
		blockKey(chainID, block.Number): {
			{
				Family: defaultBlocksFamily,
				Column: blocksDataColumn,
				Data:   b,
			},
		},
	}); err != nil {
		return err
	}
	return nil
}

func (store StoreV1) GetBlock(chainID string, number uint64) (*types.Eth1Block, error) {
	row, err := store.blocks.GetRow(blockKey(chainID, number))
	if err != nil {
		return nil, err
	}
	var block types.Eth1Block
	if err := proto.Unmarshal(row.Values[fmt.Sprintf("%s:%s", defaultBlocksFamily, blocksDataColumn)], &block); err != nil {
		return nil, err
	}
	return &block, nil
}

func (store StoreV1) GetBlocksRange(chainID string, start, end uint64) ([]*types.Eth1Block, error) {
	if end < start {
		return nil, fmt.Errorf("invalid block range provided (high: %v, low: %v)", end, start)
	}

	rows, err := store.blocks.GetRowsRange(blockKey(chainID, end), blockKey(chainID, start))
	if err != nil {
		return nil, err
	}
	var blocks []*types.Eth1Block
	for _, row := range rows {
		var block types.Eth1Block
		if err := proto.Unmarshal(row.Values[fmt.Sprintf("%s:%s", defaultBlocksFamily, blocksDataColumn)], &block); err != nil {
			return nil, err
		}
		blocks = append(blocks, &block)
	}
	return blocks, nil
}

// RevertBlock
// - revert contract updates in metadata table
// - retrieve and delete all indexing keys concerning that block in data table
// - delete the block in blocks table
func (store StoreV1) RevertBlock(chainID string, number uint64, blockHash []byte) error {
	if err := store.revertContractUpdate(chainID, number); err != nil {
		return err
	}
	if err := store.deleteBlockKeys(chainID, number, blockHash); err != nil {
		return err
	}
	if err := store.deleteBlock(chainID, number); err != nil {
		return err
	}
	return nil
}

func (store StoreV1) revertContractUpdate(chainID string, number uint64) error {
	start, err := encodeIsContractUpdateTs(number, 0, 0)
	if err != nil {
		return err
	}
	end, err := encodeIsContractUpdateTs(number+1, 0, 0)
	if err != nil {
		return err
	}

	// handle contract state updates
	// TODO: this is potentially very resources consuming
	rows, err := store.metadata.Read(fmt.Sprintf("%s:S:", chainID),
		database.WithFamilyFilter(accountFamily),
		database.WithColumnFilter(accountIsContractColumn),
		database.WithTimestampRangeFilter(start, end-1),
	)
	if err != nil {
		return err
	}
	var keys []string
	for _, row := range rows {
		keys = append(keys, row.Key)
	}
	if err := store.metadata.DeleteRowsWithKeys(keys,
		database.WithFamilyFilter(accountFamily),
		database.WithColumnFilter(accountIsContractColumn),
		database.WithTimestampRangeFilter(start, end),
	); err != nil {
		return err
	}
	return nil
}

func (store StoreV1) deleteBlockKeys(chainID string, number uint64, blockHash []byte) error {
	keys, err := store.getBlockKeys(chainID, number, blockHash)
	if err != nil {
		return err
	}
	if err := store.data.DeleteRowsWithKeys(keys); err != nil {
		return err
	}
	return nil
}

func (store StoreV1) deleteBlock(chainID string, number uint64) error {
	if err := store.blocks.DeleteRowsWithKeys([]string{blockKey(chainID, number)}); err != nil {
		return err
	}
	return nil
}

func (store StoreV1) getBlockKeys(chainID string, blockNumber uint64, blockHash []byte) ([]string, error) {
	row, err := store.updates.GetRow(fmt.Sprintf("%s:BLOCK:%s:%x", chainID, reversedPaddedBlockNumber(blockNumber), blockHash))
	if err != nil {
		return nil, err
	}
	return strings.Split(string(row.Values[fmt.Sprintf("%s:%s", updatesBlockFamily, blockKeysColumn)]), ","), nil
}

func (store StoreV1) GetLastBlockInDataTable(chainID string) (uint64, error) {
	prefix := chainID + ":B:"
	rows, err := store.data.Read(prefix, database.WithLimit(1), database.WithoutValue())
	if err != nil {
		return 0, err
	}
	if len(rows) != 1 {
		return 0, nil
	}
	reversedLastBlockStr := strings.TrimPrefix(rows[0].Key, prefix)
	reversedLastBlock, ok := new(big.Int).SetString(reversedLastBlockStr, 10)
	if !ok {
		return 0, fmt.Errorf("failed to parse last block from string: %s", reversedLastBlockStr)
	}
	lastBlock := maxExecutionLayerBlockNumber - reversedLastBlock.Uint64()
	return lastBlock, nil
}

func (store StoreV1) GetLastBlockInBlocksTable(chainID string) (uint64, error) {
	prefix := chainID + ":"
	rows, err := store.blocks.Read(prefix, database.WithLimit(1), database.WithoutValue())
	if err != nil {
		return 0, err
	}
	if len(rows) != 1 {
		return 0, nil
	}
	reversedLastBlockStr := strings.TrimPrefix(rows[0].Key, prefix)
	reversedLastBlock, ok := new(big.Int).SetString(reversedLastBlockStr, 10)
	if !ok {
		return 0, fmt.Errorf("failed to parse last block from string: %s", reversedLastBlockStr)
	}
	lastBlock := maxExecutionLayerBlockNumber - reversedLastBlock.Uint64()
	return lastBlock, nil
}

func (store StoreV1) GetENSUpdate(chainID string, batchSize int64) ([]ENSLog, error) {
	key := fmt.Sprintf("%s:ENS:V", chainID)
	rows, err := store.data.Read(key, database.WithoutValue(), database.WithLimit(batchSize))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	var ensLogs []ENSLog
	for _, row := range rows {
		split := strings.Split(row.Key, ":")
		ensType := split[3]
		value := split[4]

		var log ENSLog
		switch ensType {
		case "H":
			nameHash, err := hex.DecodeString(value)
			if err != nil {
				return nil, fmt.Errorf("cannot decode name hash for row %s: %w", row.Key, err)
			}
			log.Node = toPointer(toByte32(nameHash))
		case "A":
			ownerHash, err := hex.DecodeString(value)
			if err != nil {
				return nil, fmt.Errorf("cannot decode address hash for row %s: %w", row.Key, err)
			}
			log.Owner = toPointer(common.BytesToAddress(ownerHash))
		case "N":
			log.Name = toPointer(value)
		default:
			return nil, fmt.Errorf("unknown ens type for row %s", row.Key)
		}
		ensLogs = append(ensLogs, log)
	}
	return ensLogs, nil
}

func (store StoreV1) DeleteENSUpdate(chainID string, logs []ENSLog) error {
	keys := maps.Keys(ensToItems(chainID, logs))
	return store.data.DeleteRowsWithKeys(keys)
}

func blockKey(chainID string, number uint64) string {
	return fmt.Sprintf("%s:%s", chainID, reversedPaddedBlockNumber(number))
}

func mergeItems(dest map[string][]database.Item, source map[string][]database.Item) {
	for key, items := range source {
		if _, ok := dest[key]; !ok {
			dest[key] = items
			continue
		}
		dest[key] = append(dest[key], items...)
	}
}

func toByte32(source []byte) [32]byte {
	var dest [32]byte
	copy(dest[:], source)
	return dest
}

func toPointer[T any](i T) *T {
	return &i
}
