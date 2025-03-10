package db2test

import (
	"context"
	"testing"

	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database/databasetest"
)

func NewStoreAndCachedLastBlocks(t *testing.T) (db2.StoreV1, db2.CachedLastBlocks) {
	t.Helper()
	client, admin := databasetest.NewBigTable(t)
	db, err := database.NewBigTableWithClient(context.Background(), client, admin, db2.Schema)
	if err != nil {
		t.Fatal(err)
	}
	store := db2.NewStoreV1FromBigtable(db, database.NoopCache{})
	return store, db2.NewCachedLastBlocks(&database.MemCache{}, store)
}
