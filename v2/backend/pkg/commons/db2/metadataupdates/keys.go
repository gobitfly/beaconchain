package metadataupdates

import (
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

const (
	balanceKey = "B"

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

func MarkBalanceUpdate(chainID string, address []byte, token []byte, cache Cache) map[string][]database.Item {
	items := make(map[string][]database.Item)

	key := fmt.Sprintf("%s:%s:%x", chainID, balanceKey, address) // format is B: for balance update as chainid:prefix:address (token id will be encoded as column name)
	keyCache := []byte(fmt.Sprintf("%s:%x", key, token))
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
