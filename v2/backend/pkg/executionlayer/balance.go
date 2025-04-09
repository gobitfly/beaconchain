package executionlayer

import (
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/executionlayer/evm"
)

type BalanceStore interface {
	UpdateBalance(chainID string, balances []db2.Balance) error
}

type BalanceUpdateStore interface {
	GetPairsToUpdate(chainID string, batchSize int64) ([]db2.Pair, error)
	DeletePairs(chainID string, pairs []db2.Pair) error
}

type BalanceUpdater struct {
	updates BalanceUpdateStore
	store   BalanceStore
	batcher evm.Batcher
}

func NewBalanceUpdater(updates BalanceUpdateStore, store BalanceStore, batcher evm.Batcher) BalanceUpdater {
	return BalanceUpdater{
		updates: updates,
		store:   store,
		batcher: batcher,
	}
}

func (u BalanceUpdater) UpdateBalances(chainID string, batchSize int64) ([]db2.Balance, error) {
	pairs, err := u.updates.GetPairsToUpdate(chainID, batchSize)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve metadata updates from bigtable: %w", err)
	}

	values, err := evm.BalanceForPairs(u.batcher, pairs)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve balances from node: %w", err)
	}

	var balances []db2.Balance
	for i, value := range values {
		balances = append(balances, db2.Balance{
			Pair:  pairs[i],
			Value: value,
		})
	}
	if err := u.store.UpdateBalance(chainID, balances); err != nil {
		return nil, fmt.Errorf("cannot save balances to bigtable: %w", err)
	}
	if err := u.updates.DeletePairs(chainID, pairs); err != nil {
		return nil, fmt.Errorf("cannot delete metadata updates from bigtable: %w", err)
	}
	return balances, nil
}
