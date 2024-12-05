package data_test

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/data"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
)

var chainIDs = []string{
	"1",
	"17000",
	"100",
}

var addresses = []common.Address{
	common.HexToAddress("0x95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5"),
	common.HexToAddress("0x388C818CA8B9251b393131C08a736A67ccB19297"),
	common.HexToAddress("0x6d2e03b7EfFEae98BD302A9F836D0d6Ab0002766"),
	common.HexToAddress("0x10e4597ff93cbee194f4879f8f1d54a370db6969"),
}

func dbFromEnv(t *testing.T, table string) database.Database {
	project := os.Getenv("BIGTABLE_PROJECT")
	instance := os.Getenv("BIGTABLE_INSTANCE")
	remote := os.Getenv("REMOTE_URL")
	if project != "" && instance != "" {
		db, err := database.NewBigTable(project, instance, nil)
		if err != nil {
			t.Fatal(err)
		}
		return database.Wrap(db, table)
	}
	if remote != "" {
		return database.NewRemoteClient(remote)
	}
	t.Skip("skipping test, set BIGTABLE_PROJECT and BIGTABLE_INSTANCE or REMOTE_URL")
	return nil
}

func TestStoreExternal(t *testing.T) {
	db := dbFromEnv(t, data.Table)
	store := data.NewStore(db)

	/*
		list of filter
		data.ByMethod(method)
		data.ByAsset(asset)
		data.OnlyReceived()
		data.OnlySent()
		data.IgnoreTransactions()
		data.IgnoreTransfers()
		data.WithTimeRange(from, to)
	*/
	tests := []struct {
		name      string
		limit     int64
		chainIDs  []string
		scroll    int
		addresses []common.Address
		opts      []data.Option
	}{
		{
			name:      "all no filter",
			limit:     25,
			scroll:    2,
			chainIDs:  chainIDs,
			addresses: addresses,
		},
		{
			name:      "tx with erc20 transfer method",
			limit:     10,
			chainIDs:  chainIDs,
			addresses: addresses,
			opts: []data.Option{
				data.OnlyTransactions(),
				data.ByMethod("a9059cbb"),
			},
		},
		{
			name:      "transfer",
			limit:     10,
			chainIDs:  chainIDs,
			addresses: addresses,
			opts: []data.Option{
				data.OnlyTransactions(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lastPrefixes map[string]string
			for i := 0; i < tt.scroll+1; i++ {
				interactions, prefixes, err := store.Get(addresses, lastPrefixes, tt.limit, tt.opts...)
				if err != nil {
					t.Fatal(err)
				}
				for _, interaction := range interactions {
					t.Log(interaction.ChainID, "0x"+interaction.From, "0x"+interaction.To, "0x"+hex.EncodeToString(interaction.Hash), interaction.Time)
				}
				lastPrefixes = prefixes
			}
		})
	}
}
