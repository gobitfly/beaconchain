package executionlayer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/metadata"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/metadataupdates"
	"github.com/gobitfly/beaconchain/pkg/executionlayer/evm"
)

func TestBalanceUpdater_UpdateBalancesErr(t *testing.T) {
	tests := []struct {
		name     string
		updates  UpdatesStore
		metadata MetadataStore
		batcher  evm.Batcher
		wantErr  string
	}{
		{
			name:     "update metadata error read",
			updates:  stubUpdates{updateErr: fmt.Errorf("some error")},
			metadata: stubMetadata{},
			batcher:  stubBatcher{},
			wantErr:  "cannot retrieve metadata updates",
		},
		{
			name:     "update metadata error delete",
			updates:  stubUpdates{deleteErr: fmt.Errorf("some error")},
			metadata: stubMetadata{},
			batcher:  stubBatcher{},
			wantErr:  "cannot delete metadata updates",
		},
		{
			name:     "metadata error",
			updates:  stubUpdates{},
			metadata: stubMetadata{err: fmt.Errorf("some error")},
			batcher:  stubBatcher{},
			wantErr:  "cannot save balances",
		},
		{
			name:     "node error",
			updates:  stubUpdates{},
			metadata: stubMetadata{},
			batcher:  stubBatcher{err: fmt.Errorf("some error")},
			wantErr:  "cannot retrieve balances",
		},
		{
			name:     "no error",
			updates:  stubUpdates{},
			metadata: stubMetadata{},
			batcher:  stubBatcher{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := NewBalanceUpdater("", tt.updates, tt.metadata, tt.batcher)
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

type stubUpdates struct {
	updateErr error
	deleteErr error
}

func (s stubUpdates) GetPairsToUpdate(string, int64) ([]metadataupdates.Pair, error) {
	return nil, s.updateErr
}

func (s stubUpdates) DeletePairs(string, []metadataupdates.Pair) error {
	return s.deleteErr
}

type stubMetadata struct {
	err error
}

func (s stubMetadata) UpdateBalance(string, []metadata.Balance) error {
	return s.err
}

type stubBatcher struct {
	err error
}

func (s stubBatcher) Batch(elements []evm.BatchElement) ([]evm.BatchResponse, error) {
	return nil, s.err
}
