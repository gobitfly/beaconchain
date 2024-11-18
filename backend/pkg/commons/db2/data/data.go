package data

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type Store struct {
	db database.Database
}

func NewStore(store database.Database) Store {
	return Store{
		db: store,
	}
}

type TransferWithIndexes struct {
	Indexed  *types.Eth1ERC20Indexed
	TxIndex  int
	LogIndex int
}

func (store Store) BlockERC20TransfersToItems(chainID string, transfers []TransferWithIndexes) (map[string][]database.Item, error) {
	items := make(map[string][]database.Item)
	for _, transfer := range transfers {
		b, err := proto.Marshal(transfer.Indexed)
		if err != nil {
			return nil, err
		}
		key := keyERC20(chainID, transfer.Indexed.ParentHash, transfer.LogIndex)
		item := []database.Item{{Family: defaultFamily, Column: key}}
		items[key] = []database.Item{{Family: defaultFamily, Column: dataColumn, Data: b}}

		items[keyERC20Time(chainID, transfer.Indexed, transfer.Indexed.From, transfer.TxIndex, transfer.LogIndex)] = item
		items[keyERC20Time(chainID, transfer.Indexed, transfer.Indexed.To, transfer.TxIndex, transfer.LogIndex)] = item

		items[keyERC20ContractAllTime(chainID, transfer.Indexed, transfer.TxIndex, transfer.LogIndex)] = item
		items[keyERC20ContractTime(chainID, transfer.Indexed, transfer.Indexed.From, transfer.TxIndex, transfer.LogIndex)] = item
		items[keyERC20ContractTime(chainID, transfer.Indexed, transfer.Indexed.To, transfer.TxIndex, transfer.LogIndex)] = item

		items[keyERC20To(chainID, transfer.Indexed, transfer.TxIndex, transfer.LogIndex)] = item
		items[keyERC20From(chainID, transfer.Indexed, transfer.TxIndex, transfer.LogIndex)] = item
		items[keyERC20Sent(chainID, transfer.Indexed, transfer.TxIndex, transfer.LogIndex)] = item
		items[keyERC20Received(chainID, transfer.Indexed, transfer.TxIndex, transfer.LogIndex)] = item
	}
	return items, nil
}

func (store Store) AddBlockERC20Transfers(chainID string, transactions []TransferWithIndexes) error {
	items, err := store.BlockERC20TransfersToItems(chainID, transactions)
	if err != nil {
		return err
	}
	return store.db.BulkAdd(items)
}

func (store Store) BlockTransactionsToItems(chainID string, transactions []*types.Eth1TransactionIndexed) (map[string][]database.Item, error) {
	items := make(map[string][]database.Item)
	for i, transaction := range transactions {
		b, err := proto.Marshal(transaction)
		if err != nil {
			return nil, err
		}
		key := keyTx(chainID, transaction.GetHash())
		item := []database.Item{{Family: defaultFamily, Column: key}}
		items[key] = []database.Item{{Family: defaultFamily, Column: dataColumn, Data: b}}
		items[keyTxSent(chainID, transaction, i)] = item
		items[keyTxReceived(chainID, transaction, i)] = item

		items[keyTxTime(chainID, transaction, transaction.To, i)] = item
		items[keyTxBlock(chainID, transaction, transaction.To, i)] = item
		items[keyTxMethod(chainID, transaction, transaction.To, i)] = item

		items[keyTxTime(chainID, transaction, transaction.From, i)] = item
		items[keyTxBlock(chainID, transaction, transaction.From, i)] = item
		items[keyTxMethod(chainID, transaction, transaction.From, i)] = item

		if transaction.ErrorMsg != "" {
			items[keyTxError(chainID, transaction, transaction.To, i)] = item
			items[keyTxError(chainID, transaction, transaction.From, i)] = item
		}

		if transaction.IsContractCreation {
			items[keyTxContractCreation(chainID, transaction, transaction.To, i)] = item
			items[keyTxContractCreation(chainID, transaction, transaction.From, i)] = item
		}
	}
	return items, nil
}

func (store Store) AddBlockTransactions(chainID string, transactions []*types.Eth1TransactionIndexed) error {
	items, err := store.BlockTransactionsToItems(chainID, transactions)
	if err != nil {
		return err
	}
	return store.db.BulkAdd(items)
}

func (store Store) Get(chainIDs []string, addresses []common.Address, prefixes map[string]map[string]string, limit int64, opts ...Option) ([]*Interaction, map[string]map[string]string, error) {
	sources := map[formatType]unMarshalInteraction{
		typeTx:       unMarshalTx,
		typeTransfer: unMarshalTransfer,
	}
	options := apply(opts)
	if options.ignoreTxs {
		delete(sources, typeTx)
	}
	if options.ignoreTransfers {
		delete(sources, typeTransfer)
	}
	var interactions []*interactionWithInfo
	for interactionType, unMarshalFunc := range sources {
		filter, err := makeFilters(options, interactionType)
		if err != nil {
			return nil, nil, err
		}
		temp, err := store.getBy(unMarshalFunc, chainIDs, addresses, prefixes, limit, filter)
		if err != nil {
			return nil, nil, err
		}
		interactions = append(interactions, temp...)
	}
	sort.Sort(byTimeDesc(interactions))
	if int64(len(interactions)) > limit {
		interactions = interactions[:limit]
	}

	var res []*Interaction
	if prefixes == nil {
		prefixes = make(map[string]map[string]string)
	}
	for i := 0; i < len(interactions); i++ {
		if prefixes[interactions[i].chainID] == nil {
			prefixes[interactions[i].chainID] = make(map[string]string)
		}
		prefixes[interactions[i].chainID][interactions[i].root] = interactions[i].key
		res = append(res, interactions[i].Interaction)
	}
	return res, prefixes, nil
}

func (store Store) getBy(unMarshal unMarshalInteraction, chainIDs []string, addresses []common.Address, prefixes map[string]map[string]string, limit int64, condition filter) ([]*interactionWithInfo, error) {
	var interactions []*interactionWithInfo
	for _, chainID := range chainIDs {
		for _, address := range addresses {
			root := condition.get(chainID, address)
			prefix := root
			if prefixes != nil && prefixes[chainID] != nil && prefixes[chainID][root] != "" {
				prefix = prefixes[chainID][root]
			}
			upper := condition.limit(root)
			indexRows, err := store.db.GetRowsRange(upper, prefix, database.WithLimit(limit), database.WithOpenRange(true))
			if err != nil {
				if errors.Is(err, database.ErrNotFound) {
					continue
				}
				return nil, err
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
				return nil, err
			}
			for _, row := range txRows {
				interaction, err := unMarshal(row.Values[fmt.Sprintf("%s:%s", defaultFamily, dataColumn)])
				if err != nil {
					return nil, err
				}
				interaction.ChainID = chainID
				interactions = append(interactions, &interactionWithInfo{
					Interaction: interaction,
					chainID:     chainID,
					root:        root,
					key:         txKeys[row.Key],
				})
			}
		}
	}
	return interactions, nil
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

type unMarshalInteraction func(b []byte) (*Interaction, error)

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
