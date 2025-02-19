package metadataupdates

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/data"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

// Store encapsulate all the interaction with the metadata_updates store, it contains:
//   - the mutation for a block (in case of reorg)
//   - the balance update keys (balance to update, record is deleted once balance is updated)
//   - the contract keys (to track if an address is a contract)
//
// also have caching to prevent putting multiple balance to update for the same address / token pair
// clearing the cache is the responsibility of the caller for now
type Store struct {
	db    database.Database
	cache Cache
}

func NewStore(db database.Database, cache Cache) Store {
	return Store{
		db:    db,
		cache: cache,
	}
}

type IndexedBlock struct {
	ChainID       string
	Block         *types.Eth1BlockIndexed
	Transactions  []*types.Eth1TransactionIndexed
	Internals     []data.InternalWithIndexes
	ERC20Transfer []data.TransferWithIndexes
	Blobs         []data.BlobWithIndex
	Contracts     []ContractUpdateWithAddress
	Uncles        []data.UncleWithIndexes
	Withdrawals   []*types.Eth1WithdrawalIndexed
}

// AddIndexedBlock extracts all the data to save from an IndexedBlock
// Moreover it takes a list of keys as param to save the different keys update in the data.Store in case of a reorg
func (store Store) AddIndexedBlock(number uint64, hash []byte, block IndexedBlock, keys []string) error {
	updates := make(map[string][]database.Item)
	if len(keys) > 0 {
		metaKeys := strings.Join(keys, ",") // save block keys in order to be able to handle chain reorgs
		updates = BlockKeysMutation(block.ChainID, number, hash, metaKeys)
	}

	// mark sender and recipient for balance update
	for _, transaction := range block.Transactions {
		mergeItems(updates, MarkBalanceUpdate(block.ChainID, transaction.From, []byte{0x0}, store.cache))
		mergeItems(updates, MarkBalanceUpdate(block.ChainID, transaction.To, []byte{0x0}, store.cache))
	}
	for _, transaction := range block.Internals {
		mergeItems(updates, MarkBalanceUpdate(block.ChainID, transaction.Indexed.From, []byte{0x0}, store.cache))
		mergeItems(updates, MarkBalanceUpdate(block.ChainID, transaction.Indexed.To, []byte{0x0}, store.cache))
	}
	for _, transaction := range block.ERC20Transfer {
		mergeItems(updates, MarkBalanceUpdate(block.ChainID, transaction.Indexed.From, transaction.Indexed.TokenAddress, store.cache))
		mergeItems(updates, MarkBalanceUpdate(block.ChainID, transaction.Indexed.To, transaction.Indexed.TokenAddress, store.cache))
	}
	for _, transaction := range block.Blobs {
		mergeItems(updates, MarkBalanceUpdate(block.ChainID, transaction.Indexed.From, []byte{0x0}, store.cache))
		mergeItems(updates, MarkBalanceUpdate(block.ChainID, transaction.Indexed.To, []byte{0x0}, store.cache))
	}
	for _, transaction := range block.Withdrawals {
		mergeItems(updates, MarkBalanceUpdate(block.ChainID, transaction.Address, []byte{0x0}, store.cache))
	}
	if block.Block != nil {
		mergeItems(updates, MarkBalanceUpdate(block.ChainID, block.Block.Coinbase, []byte{0x0}, store.cache))
	}
	for _, uncle := range block.Uncles {
		mergeItems(updates, MarkBalanceUpdate(block.ChainID, uncle.Coinbase, []byte{0x0}, store.cache))
	}

	update, err := ContractUpdate(number, block.ChainID, block.Contracts)
	if err != nil {
		return err
	}
	maps.Copy(updates, update)
	if err := store.db.BulkAdd(updates); err != nil {
		return err
	}
	return nil
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

type Pair struct {
	Address common.Address
	Token   common.Address
}

func (store Store) GetPairsToUpdate(chainID string, batchSize int64) ([]Pair, error) {
	key := fmt.Sprintf("%s:%s", chainID, balanceKey)
	rows, err := store.db.GetRowsRange(toSuccessor(key), fmt.Sprintf("%s:", key), database.WithLimit(batchSize))
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

func addressFromKey(key string) common.Address {
	return common.HexToAddress(strings.Split(key, ":")[2])
}

func tokenFromColumn(column string) common.Address {
	return common.HexToAddress(strings.Split(column, ":")[1])
}

func (store Store) DeletePairs(chainID string, pairs []Pair) error {
	var keys []string
	for _, pair := range pairs {
		keys = append(keys, fmt.Sprintf("%s:%s:%x", chainID, balanceKey, pair.Address.Bytes()))
	}
	slices.Sort(keys)
	keys = slices.Compact(keys)
	return store.db.DeleteRowsWithKeys(keys)
}

// toSuccessor add suffix ";" has it comes after ":" in the ascii order
// this is a simple way to have an infinite bound limit
// prefix must be a real prefix and not a key
func toSuccessor(prefix string) string {
	return prefix + ";"
}
