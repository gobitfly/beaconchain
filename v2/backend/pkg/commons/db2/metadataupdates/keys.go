package metadataupdates

import (
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"

	"google.golang.org/protobuf/proto"
)

const (
	maxExecutionLayerBlockNumber = 1000000000
)

type Cache interface {
	Set(key, value []byte, expireSeconds int) (err error)
	Get(key []byte) (value []byte, err error)
}

func BlockKeysMutation(chainID string, blockNumber uint64, blockHash []byte, keys string) map[string][]database.Item {
	items := make(map[string][]database.Item)
	key := fmt.Sprintf("%s:BLOCK:%s:%x", chainID, reversedPaddedBlockNumber(blockNumber), blockHash)
	items[key] = []database.Item{
		{
			Family: updatesBlockFamily,
			Column: blockKeysColumn,
			Data:   []byte(keys),
		},
	}
	return items
}

type ContractUpdateWithAddress struct {
	Indexed       *types.IsContractUpdate
	Address       []byte
	TxIndex       int
	InternalIndex int
}

func ContractUpdate(blockNumber uint64, chainID string, updates []ContractUpdateWithAddress) (map[string][]database.Item, error) {
	items := make(map[string][]database.Item)
	for _, update := range updates {
		b, err := proto.Marshal(update.Indexed)
		if err != nil {
			return nil, err
		}

		key := fmt.Sprintf("%s:S:%x", chainID, update.Address)
		ts, err := encodeIsContractUpdateTs(blockNumber, uint64(update.TxIndex), uint64(update.InternalIndex))
		if err != nil {
			return nil, fmt.Errorf("error generating bigtable isContract timestamp: %w", err)
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
	return items, nil
}

func MarkBalanceUpdate(chainID string, address []byte, token []byte, cache Cache) map[string][]database.Item {
	items := make(map[string][]database.Item)

	key := fmt.Sprintf("%s:B:%x", chainID, address) // format is B: for balance update as chainid:prefix:address (token id will be encoded as column name)
	keyCache := []byte(fmt.Sprintf("%s:B:%x:%x", chainID, address, token))
	if _, err := cache.Get(keyCache); err != nil {
		items[key] = []database.Item{
			{
				Family: defaultFamily,
				Column: fmt.Sprintf("%x", token),
			},
		}
		_ = cache.Set(keyCache, []byte{0x1}, int((utils.Day * 2).Seconds()))
	}
	return items
}

func reversedPaddedBlockNumber(blockNumber uint64) string {
	return fmt.Sprintf("%09d", maxExecutionLayerBlockNumber-blockNumber)
}
