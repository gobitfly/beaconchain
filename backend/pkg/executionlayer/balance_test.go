package executionlayer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/executionlayer/evm"
)

func TestBalanceUpdater_UpdateBalancesErr(t *testing.T) {
	tests := []struct {
		name    string
		store   BalanceStore
		batcher evm.Batcher
		wantErr string
	}{
		{
			name:    "update metadata error read",
			store:   stubBalanceStore{readErr: fmt.Errorf("some error")},
			batcher: stubBatcher{},
			wantErr: "cannot retrieve metadata updates",
		},
		{
			name:    "update metadata error delete",
			store:   stubBalanceStore{deleteErr: fmt.Errorf("some error")},
			batcher: stubBatcher{},
			wantErr: "cannot delete metadata updates",
		},
		{
			name:    "metadata error",
			store:   stubBalanceStore{updateErr: fmt.Errorf("some error")},
			batcher: stubBatcher{},
			wantErr: "cannot save balances",
		},
		{
			name:    "node error",
			store:   stubBalanceStore{},
			batcher: stubBatcher{err: fmt.Errorf("some error")},
			wantErr: "cannot retrieve balances",
		},
		{
			name:    "no error",
			store:   stubBalanceStore{},
			batcher: stubBatcher{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewBalanceUpdater("", tt.store, tt.batcher)
			_, err := u.UpdateBalances(0)
			if err == nil {
				if tt.wantErr != "" {
					t.Fatalf("UpdateBalances() expected an error")
				}
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("'%s' should contains '%s'", err.Error(), tt.wantErr)
			}
		})
	}
}

type stubBalanceStore struct {
	readErr   error
	deleteErr error
	updateErr error
}

func (s stubBalanceStore) GetPairsToUpdate(chainID string, batchSize int64) ([]db2.Pair, error) {
	return nil, s.readErr
}

func (s stubBalanceStore) DeletePairs(chainID string, pairs []db2.Pair) error {
	return s.deleteErr
}

func (s stubBalanceStore) UpdateBalance(chainID string, balances []db2.Balance) error {
	return s.updateErr
}

type stubBatcher struct {
	err error
}

func (s stubBatcher) Batch(elements []evm.BatchElement) ([]evm.BatchResponse, error) {
	return nil, s.err
}
