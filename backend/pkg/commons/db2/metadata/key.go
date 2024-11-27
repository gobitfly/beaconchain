package metadata

import (
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

const (
	METADATA_UPDATES_FAMILY_BLOCKS = "blocks"
	maxExecutionLayerBlockNumber   = 1000000000
	DEFAULT_FAMILY                 = "f"
)

type Cache interface {
	Set(key, value []byte, expireSeconds int) (err error)
	Get(key []byte) (value []byte, err error)
}

// todo proper implement this function - Tangui
func BlockKeysMutation(chainID string, blockNumber uint64, blockHash []byte, keys string) map[string][]database.Item {
	items := make(map[string][]database.Item)
	key := fmt.Sprintf("%s:BLOCK:%s:%x", chainID, reversedPaddedBlockNumber(blockNumber), blockHash)
	items[key] = []database.Item{
		{
			Family: METADATA_UPDATES_FAMILY_BLOCKS,
			Column: "keys",
			Data:   []byte(keys),
		},
	}
	return items
}

func MarkBalanceUpdate(chainID string, address []byte, token []byte, cache Cache) map[string][]database.Item {
	items := make(map[string][]database.Item)

	key := fmt.Sprintf("%s:B:%x", chainID, address) // format is B: for balance update as chainid:prefix:address (token id will be encoded as column name)
	keyCache := []byte(fmt.Sprintf("%s:B:%x:%x", chainID, address, token))
	if _, err := cache.Get(keyCache); err != nil {
		items[key] = []database.Item{
			{
				Family: DEFAULT_FAMILY,
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
