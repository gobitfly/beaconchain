package db2

import (
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
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

func markBalanceUpdate(chainID string, address []byte, token []byte, cache CachedBalanceUpdates) map[string][]database.Item {
	key := fmt.Sprintf("%s:%s:%x", chainID, balanceKey, address) // format is B: for balance update as chainid:prefix:address (token id will be encoded as column name)
	if cache.Add(chainID, address, token) {
		return nil
	}

	return map[string][]database.Item{
		key: {
			{
				Family: defaultFamily,
				Column: fmt.Sprintf("%x", token),
			},
		},
	}
}
