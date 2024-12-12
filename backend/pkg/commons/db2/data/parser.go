package data

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

func BlockTransactionsToItemsV2(chainID string, transactions []*types.Eth1TransactionIndexed) (map[string][]database.Item, error) {
	if len(transactions) > txPerBlockLimit {
		return nil, fmt.Errorf("unexpected number of transactions in block expected at most %d but got: %v, chainID: %s, block: %d", txPerBlockLimit, len(transactions), chainID, transactions[0].BlockNumber)
	}

	items := make(map[string][]database.Item)
	for i, transaction := range transactions {
		b, err := proto.Marshal(transaction)
		if err != nil {
			return nil, err
		}
		key, indexes := transactionKeysV3(chainID, transaction, i)
		items[key] = []database.Item{{Family: defaultFamily, Column: dataColumn, Data: b}}
		for _, index := range indexes {
			items[index] = []database.Item{{Family: defaultFamily, Column: key}}
		}
	}
	return items, nil
}

func BlockERC20TransfersToItemsV2(chainID string, transfers []TransferWithIndexes) (map[string][]database.Item, error) {
	items := make(map[string][]database.Item)
	for _, transfer := range transfers {
		if transfer.TxIndex > txPerBlockLimit {
			return nil, fmt.Errorf("unexpected number of transactions in block expected at most %d but got: %v, chainID: %s, block: %d", txPerBlockLimit, transfer.TxIndex, chainID, transfers[0].Indexed.BlockNumber)
		}
		if transfer.LogIndex > logPerTxLimit {
			return nil, fmt.Errorf("unexpected number of logs in block expected at most %d but got: %v, chainID: %s, block: %d", logPerTxLimit, transfer.LogIndex, chainID, transfers[0].Indexed.BlockNumber)
		}
		b, err := proto.Marshal(transfer.Indexed)
		if err != nil {
			return nil, err
		}
		key, indexes := transferKeysV3(chainID, transfer.Indexed, transfer.TxIndex, transfer.LogIndex)
		items[key] = []database.Item{{Family: defaultFamily, Column: dataColumn, Data: b}}
		for _, index := range indexes {
			items[index] = []database.Item{{Family: defaultFamily, Column: key}}
		}
	}
	return items, nil
}

type TransferWithIndexes struct {
	Indexed  *types.Eth1ERC20Indexed
	TxIndex  int
	LogIndex int
}
