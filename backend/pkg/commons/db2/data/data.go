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

// Store encapsulate all the interaction with the data store which contains all the indexed objects
type Store struct {
	db database.Database
}

func NewStore(db database.Database) Store {
	return Store{
		db: db,
	}
}

type TransferWithIndexes struct {
	Indexed  *types.Eth1ERC20Indexed
	TxIndex  int
	LogIndex int
}

type IndexedBlock struct {
	ChainID         string
	Block           *types.Eth1BlockIndexed
	Transactions    []*types.Eth1TransactionIndexed
	Internals       []InternalWithIndexes
	ERC20Transfer   []TransferWithIndexes
	ERC1155Transfer []ERC1155TransferWithIndexes
	ERC721Transfer  []ERC721TransferWithIndexes
	Blobs           []BlobWithIndex
	Uncles          []UncleWithIndexes
	Withdrawals     []*types.Eth1WithdrawalIndexed
	ENS             []ENSLog
}

func (store Store) AddIndexedBlock(block IndexedBlock) ([]string, error) {
	updates := make(map[string][]database.Item)

	update, err := TransactionsToItems(block.ChainID, block.Transactions)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = BlockToItems(block.ChainID, block.Block)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = InternalsToItems(block.ChainID, block.Internals)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = ERC20TransfersToItems(block.ChainID, block.ERC20Transfer)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = ERC1155TransfersToItems(block.ChainID, block.ERC1155Transfer)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = ERC721TransfersToItems(block.ChainID, block.ERC721Transfer)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = BlobToItems(block.ChainID, block.Blobs)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = UnclesToItems(block.ChainID, block.Uncles)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = WithdrawalToItems(block.ChainID, block.Withdrawals)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	update, err = ENSToItems(block.ChainID, block.ENS)
	if err != nil {
		return nil, err
	}
	maps.Copy(updates, update)

	if err := store.db.BulkAdd(updates); err != nil {
		return nil, err
	}

	return maps.Keys(updates), nil
}

// Get allows to retrieve Transactions and ERC20 Transfers from the Store
// some Option can be applied to add filter on the search
// This is a WIP functionality, used only in tests for now
// but it will the base for account dashboard functionalities
func (store Store) Get(chainIDs []string, addresses []common.Address, prefixes map[string]map[string]string, limit int64, opts ...Option) ([]*Interaction, map[string]map[string]string, error) {
	sources := map[formatType]unMarshalInteraction{
		typeTx:       unMarshalTx,
		typeTransfer: unMarshalTransfer,
	}
	options := apply(opts)
	if options.onlyTransfers {
		delete(sources, typeTx)
	}
	if options.onlyTxs {
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
