package db2

import (
	"context"
	"testing"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database/databasetest"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

func TestStoreV1(t *testing.T) {
	btClient, btAdmin := databasetest.NewBigTable(t)
	bt, err := database.NewBigTableWithClient(context.Background(), btClient, btAdmin, Schema)
	if err != nil {
		t.Fatal(err)
	}
	defer bt.Close()
	store := NewStoreV1FromBigtable(bt, CachedBalanceUpdates{database.NoopCache{}})

	t.Run("block range", func(t *testing.T) {
		tests := []struct {
			name       string
			start, end uint64
		}{
			{
				name:  "normal",
				start: 10,
				end:   20,
			},
			{
				name:  "genesis",
				start: 0,
				end:   1,
			},
			{
				name:  "start and end at genesis ",
				start: 0,
				end:   0,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				for i := tt.start; i <= tt.end; i++ {
					if err := store.SaveBlock("1", &types.Eth1Block{
						Number: i,
					}); err != nil {
						t.Fatal(err)
					}
				}
				blocks, err := store.GetBlocksRange("1", tt.start, tt.end)
				if err != nil {
					t.Fatal(err)
				}
				if got, want := len(blocks), int(tt.end-tt.start)+1; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
			})
		}
	})
}
