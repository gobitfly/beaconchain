package executionlayer

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/exp/maps"

	"github.com/gobitfly/beaconchain/internal/contracts"
	"github.com/gobitfly/beaconchain/internal/th"
	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database/databasetest"
	"github.com/gobitfly/beaconchain/pkg/executionlayer/evm"
)

func TestBalanceUpdater(t *testing.T) {
	btClient, btAdmin := databasetest.NewBigTable(t)
	bt, err := database.NewBigTableWithClient(context.Background(), btClient, btAdmin, db2.Schema)
	if err != nil {
		t.Fatal(err)
	}
	store := db2.NewStoreV1FromBigtable(bt, database.NoopCache{})
	backend := th.NewBackend(t)

	multicall := backend.DeployContract(t, common.FromHex(contracts.MulticallMetaData.Bin))
	batcher := evm.NewMulticallBatcher(backend.Client(), multicall, 0)

	expectedBalance, err := backend.Client().BalanceAt(context.Background(), backend.BankAccount.From, nil)
	if err != nil {
		t.Fatal(err)
	}
	expected := db2.Balance{
		Pair: db2.Pair{
			Address: backend.BankAccount.From,
			Token:   common.Address{},
		},
		Value: expectedBalance,
	}
	updates := newStubUpdatesStore([]db2.Pair{expected.Pair})
	updater := NewBalanceUpdater(fmt.Sprintf("%d", backend.ChainID), updates, store, batcher)

	balances, err := updater.UpdateBalances(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(balances) != 1 {
		t.Fatalf("got %d balances, want 1", len(balances))
	}
	if got, want := balances[0].Pair, expected.Pair; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := balances[0].Value.String(), expectedBalance.String(); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if len(updates.updates) != 0 {
		t.Error("updates were not deleted")
	}
}

type stubUpdatesStore struct {
	updates map[string]db2.Pair
}

func newStubUpdatesStore(pairs []db2.Pair) stubUpdatesStore {
	store := stubUpdatesStore{
		updates: make(map[string]db2.Pair),
	}
	for _, pair := range pairs {
		store.updates[fmt.Sprintf("%s%s", pair.Address, pair.Token)] = pair
	}
	return store
}

func (s stubUpdatesStore) GetPairsToUpdate(chainID string, batchSize int64) ([]db2.Pair, error) {
	return maps.Values(s.updates), nil
}

func (s stubUpdatesStore) DeletePairs(chainID string, pairs []db2.Pair) error {
	for _, pair := range pairs {
		delete(s.updates, fmt.Sprintf("%s%s", pair.Address, pair.Token))
	}
	return nil
}

func TestBalanceUpdater_UpdateBalancesErr(t *testing.T) {
	tests := []struct {
		name    string
		store   stubBalanceStore
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
			u := NewBalanceUpdater("", tt.store, tt.store, tt.batcher)
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
