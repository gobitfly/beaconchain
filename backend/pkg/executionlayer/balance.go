package executionlayer

import (
	"fmt"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/metadata"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/metadataupdates"
	"github.com/gobitfly/beaconchain/pkg/executionlayer/evm"
)

type UpdatesStore interface {
	GetPairsToUpdate(chainID string, batchSize int64) ([]metadataupdates.Pair, error)
	DeletePairs(chainID string, pairs []metadataupdates.Pair) error
}

type MetadataStore interface {
	UpdateBalance(chainID string, balances []metadata.Balance) error
}

type BalanceUpdater struct {
	chainID  string
	updates  UpdatesStore
	metadata MetadataStore
	batcher  evm.Batcher
}

func NewBalanceUpdater(chainID string, updates UpdatesStore, store MetadataStore, batcher evm.Batcher) BalanceUpdater {
	return BalanceUpdater{
		chainID:  chainID,
		updates:  updates,
		metadata: store,
		batcher:  batcher,
	}
}

func (u BalanceUpdater) UpdateBalances(batchSize int64) ([]metadata.Balance, error) {
	pairs, err := u.updates.GetPairsToUpdate(u.chainID, batchSize)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve metadata updates from bigtable: %w", err)
	}
	values, err := evm.BalanceForPairs(u.batcher, pairs)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve balances from node: %w", err)
	}
	var balances []metadata.Balance
	for i, value := range values {
		balances = append(balances, metadata.Balance{
			Pair:  pairs[i],
			Value: value,
		})
	}
	if err := u.metadata.UpdateBalance(u.chainID, balances); err != nil {
		return nil, fmt.Errorf("cannot save balances to bigtable: %w", err)
	}
	if err := u.updates.DeletePairs(u.chainID, pairs); err != nil {
		return nil, fmt.Errorf("cannot delete metadata updates from bigtable: %w", err)
	}
	return balances, nil
}
