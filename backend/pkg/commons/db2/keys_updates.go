package db2

import (
	"context"
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

const (
	balanceKey = "B"
)

func blockKeysMutation(chainID string, blockNumber uint64, blockHash []byte, keys string) map[string][]database.Item {
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

func markBalanceUpdate(chainID string, address []byte, token []byte, cache database.RemoteCache) map[string][]database.Item {
	items := make(map[string][]database.Item)

	key := fmt.Sprintf("%s:%s:%x", chainID, balanceKey, address) // format is B: for balance update as chainid:prefix:address (token id will be encoded as column name)
	keyCache := fmt.Sprintf("%s:%x", key, token)
	if _, err := cache.Get(context.Background(), keyCache); err != nil {
		items[key] = []database.Item{
			{
				Family: defaultFamily,
				Column: fmt.Sprintf("%x", token),
			},
		}
		_ = cache.Set(context.Background(), keyCache, []byte{0x1}, utils.Day*2)
	}
	return items
}
