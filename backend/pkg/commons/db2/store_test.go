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
	store := NewStoreV1FromBigtable(bt, CachedBalanceUpdates{database.NoopCache{}})

	t.Run("block range", func(t *testing.T) {
		defer bt.Close()
		if err := store.SaveBlock("1", &types.Eth1Block{
			Number: 10,
		}); err != nil {
			t.Fatal(err)
		}

		if err := store.SaveBlock("1", &types.Eth1Block{
			Number: 11,
		}); err != nil {
			t.Fatal(err)
		}

		blocks, err := store.GetBlocksRange("1", 10, 11)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(blocks), 2; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
