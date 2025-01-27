package data

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/proto"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type db interface {
	Add(table string, rows []any) error
	Read(query string, scan func(driver.Rows) error) error
	ClearTable(table string) error
}

type Store struct {
	db db
}

func NewStore(db db) Store {
	return Store{
		db: db,
	}
}

/*func (store Store) AddItems(items map[string][]database.Item) error {
	return store.db.BulkAdd(items)
}
*/
/*func (store Store) AddBlockERC20Transfers(chainID string, transactions []TransferWithIndexes) error {
	items, err := BlockERC20TransfersToItemsV2(chainID, transactions)
	if err != nil {
		return err
	}
	return store.db.BulkAdd(items)
}*/

func (store Store) AddBlockTransactions(chainID string, transactions []*types.Eth1TransactionIndexed) error {
	var dbTransactions []*DBTransaction
	for i, transaction := range transactions {
		dbTransactions = append(dbTransactions, &DBTransaction{
			ChainID:            chainID,
			TxHash:             hex.EncodeToString(transaction.Hash),
			TxIndex:            uint64(i),
			BlockNumber:        transaction.BlockNumber,
			Timestamp:          transaction.Time.AsTime(),
			Method:             hex.EncodeToString(transaction.MethodId),
			From:               hex.EncodeToString(transaction.From),
			To:                 hex.EncodeToString(transaction.To),
			Value:              new(big.Int).SetBytes(transaction.Value).String(),
			TxFee:              new(big.Int).SetBytes(transaction.TxFee),
			GasPrice:           new(big.Int).SetBytes(transaction.GasPrice).Uint64(),
			IsContractCreation: transaction.IsContractCreation,
			ErrorMsg:           transaction.ErrorMsg,
			BlobTxFee:          new(big.Int).SetBytes(transaction.BlobTxFee),
			BlobGasPrice:       new(big.Int).SetBytes(transaction.BlobGasPrice).Uint64(),
			Status:             transaction.Status.String(),
			Type:               "transaction", // TODO put correct value
			InsertedAt:         time.Time{},
		})
	}
	if err := store.db.Add(TableTransactions, toAnyArray(dbTransactions)); err != nil {
		return fmt.Errorf("canot add transactions: %w", err)
	}
	return nil
}

func (store Store) Get(addresses []common.Address, prefixes map[string]string, limit int64, opts ...Option) ([]DBTransaction, map[string]string, error) {
	options := apply(opts)

	filter, err := newQueryFilterV3(options)
	if err != nil {
		return nil, nil, err
	}
	databaseOptions := []database.Option{
		database.WithLimit(limit),
		database.WithOpenRange(true),
	}
	if options.statsReporter != nil {
		databaseOptions = append(databaseOptions, database.WithStats(options.statsReporter))
	}
	interactions, err := store.getBy(addresses, prefixes, filter, databaseOptions)
	if err != nil {
		return nil, nil, err
	}

	/*sort.Sort(byTimeDesc(interactions))
	if int64(len(interactions)) > limit {
		interactions = interactions[:limit]
	}

	var res []*Interaction
	if prefixes == nil {
		prefixes = make(map[string]string)
	}
	for i := 0; i < len(interactions); i++ {
		prefixes[interactions[i].root] = interactions[i].key
		res = append(res, interactions[i].Interaction)
	}*/
	return interactions, nil, nil
}

/*func (store Store) getBy(addresses []common.Address, prefixes map[string]string, condition filter, databaseOptions []database.Option) ([]*interactionWithInfo, error) {
	var g errgroup.Group
	var interactions []*interactionWithInfo
	var mu sync.Mutex
	for _, address := range addresses {
		g.Go(func() error {
			root := condition.get(address)
			prefix := root
			if prefixes != nil && prefixes[root] != "" {
				prefix = prefixes[root]
			}
			upper := condition.limit(root)
			rowKeyFilter := condition.rowKeyFilter(prefix)
			if rowKeyFilter != "" {
				databaseOptions = append(databaseOptions, database.WithRowKeyFilter(rowKeyFilter))
			}
			indexRows, err := store.db.GetRowsRange(upper, prefix, databaseOptions...)
			if err != nil {
				if errors.Is(err, database.ErrNotFound) {
					return nil
				}
				return err
			}
			txKeys := make(map[string]string)
			for _, row := range indexRows {
				for key := range row.Values {
					txKey := strings.TrimPrefix(key, fmt.Sprintf("%s:", defaultFamily))
					txKeys[txKey] = row.Key
				}
			}
			txRows, err := store.db.GetRowsWithKeys(maps.Keys(txKeys))
			if err != nil {
				return err
			}
			for _, row := range txRows {
				parts := strings.Split(row.Key, ":")
				unMarshal := unMarshalTx
				if parts[0] == "ERC20" {
					unMarshal = unMarshalTransfer
				}
				interaction, err := unMarshal(row.Values[fmt.Sprintf("%s:%s", defaultFamily, dataColumn)])
				if err != nil {
					return err
				}
				interaction.ChainID = parts[1]
				mu.Lock()
				interactions = append(interactions, &interactionWithInfo{
					Interaction: interaction,
					chainID:     parts[1],
					root:        root,
					key:         txKeys[row.Key],
				})
				mu.Unlock()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return interactions, nil
}
*/

func (store Store) getBy(addresses []common.Address, prefixes map[string]string, condition filter, databaseOptions []database.Option) ([]DBTransaction, error) {
	var transactions []DBTransaction

	if err := store.db.Read(fmt.Sprintf(`SELECT * FROM %s`, TableTransactions), database.ScanArray(&transactions)); err != nil {
		return nil, err
	}

	return transactions, nil
}

type interactionWithInfo struct {
	*Interaction
	chainID string
	root    string
	key     string
}

type byTimeDesc []*interactionWithInfo

func (c byTimeDesc) Len() int      { return len(c) }
func (c byTimeDesc) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c byTimeDesc) Less(i, j int) bool {
	t1 := c[i].Interaction.Time
	t2 := c[j].Interaction.Time
	if t1.Equal(t2) {
		return c[i].key < c[j].key
	}
	return t1.After(t2)
}

type Interaction struct {
	ChainID string
	Hash    []byte
	Method  []byte
	Time    time.Time
	Type    string
	Value   []byte
	Asset   string
	From    string
	To      string
}

var erc20Transfer, _ = hex.DecodeString("a9059cbb")

func unMarshalTx(b []byte) (*Interaction, error) {
	tx := &types.Eth1TransactionIndexed{}
	if err := proto.Unmarshal(b, tx); err != nil {
		return nil, err
	}
	return parseTx(tx), nil
}

func unMarshalTransfer(b []byte) (*Interaction, error) {
	tx := &types.Eth1ERC20Indexed{}
	if err := proto.Unmarshal(b, tx); err != nil {
		return nil, err
	}
	return parseTransfer(tx), nil
}

func parseTransfer(transfer *types.Eth1ERC20Indexed) *Interaction {
	return &Interaction{
		ChainID: "",
		Hash:    transfer.ParentHash,
		Method:  erc20Transfer,
		Time:    transfer.Time.AsTime(),
		Type:    "ERC20",
		Value:   transfer.Value,
		Asset:   hex.EncodeToString(transfer.TokenAddress),
		From:    hex.EncodeToString(transfer.From),
		To:      hex.EncodeToString(transfer.To),
	}
}

func parseTx(tx *types.Eth1TransactionIndexed) *Interaction {
	return &Interaction{
		ChainID: "",
		Hash:    tx.Hash,
		Method:  tx.MethodId,
		Time:    tx.Time.AsTime(),
		Type:    "Transaction",
		Value:   tx.Value,
		Asset:   "ETH",
		From:    hex.EncodeToString(tx.From),
		To:      hex.EncodeToString(tx.To),
	}
}

func toAnyArray[T any](array []T) []any {
	ret := make([]any, len(array))
	for i, v := range array {
		ret[i] = v
	}
	return ret
}
