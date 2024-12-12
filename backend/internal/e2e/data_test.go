package e2e

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/gobitfly/beaconchain/internal/th"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/data"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/db2test"
	"github.com/gobitfly/beaconchain/pkg/commons/indexer"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

func TestStoreWithBackend(t *testing.T) {
	store := db2test.NewDataStore(t)

	backend := th.NewBackend(t)
	usdtAddress, usdt := backend.DeployToken(t, "usdt", "usdt", backend.BankAccount.From)

	transform := indexer.NewTransformer(indexer.NoopCache{})
	indexer := indexer.New(store, transform.Tx, transform.ERC20)

	client, err := rpc.NewErigonClient(backend.Endpoint)
	if err != nil {
		t.Fatal(err)
	}

	var addresses []common.Address
	for i := 0; i < 1; i++ {
		temp := th.CreateEOA(t)
		addresses = append(addresses, temp.From)
		for j := 0; j < 50; j++ {
			if err := backend.Client().SendTransaction(context.Background(), backend.MakeTx(t, backend.BankAccount, &temp.From, big.NewInt(1), nil)); err != nil {
				t.Fatal(err)
			}
			if _, err := usdt.Mint(backend.BankAccount.TransactOpts, temp.From, big.NewInt(1)); err != nil {
				t.Fatal(i, j, err)
			}
			backend.Commit()

			lastBlock, err := backend.Client().BlockNumber(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			block, _, err := client.GetBlock(int64(lastBlock), "geth")
			if err != nil {
				t.Fatal(err)
			}
			if err := indexer.IndexBlocks(fmt.Sprintf("%d", backend.ChainID), []*types.Eth1Block{block}); err != nil {
				t.Fatal(err)
			}
		}
	}

	tests := []struct {
		name    string
		address common.Address
		opts    []data.Option
	}{
		{
			name:    "no filters",
			address: addresses[0],
		},
		{
			name:    "method",
			address: backend.BankAccount.From,
			opts:    []data.Option{data.ByMethod("40c10f19")},
		},
		{
			name:    "asset",
			address: addresses[0],
			opts:    []data.Option{data.ByAsset(usdtAddress)},
		},
		{
			name:    "received",
			address: addresses[0],
			opts:    []data.Option{data.OnlyReceived()},
		},
		{
			name:    "sent",
			address: backend.BankAccount.From,
			opts:    []data.Option{data.OnlySent()},
		},
		{
			name:    "asset sent",
			address: backend.BankAccount.From,
			opts:    []data.Option{data.ByAsset(usdtAddress), data.OnlySent()},
		},
		{
			name:    "asset received",
			address: backend.BankAccount.From,
			opts:    []data.Option{data.ByAsset(usdtAddress), data.OnlyReceived()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			efficiencies := make(map[string]map[string]string)
			opts := []data.Option{data.WithDatabaseStats(getEfficiencies(efficiencies))}
			opts = append(opts, tt.opts...)
			_, _, err := store.Get([]common.Address{tt.address}, nil, 25, opts...)
			if err != nil {
				t.Fatal(err)
			}
			/*for _, interaction := range interactions {
				t.Log(interaction.Type, interaction.ChainID, "0x"+interaction.From, "0x"+interaction.To, "0x"+hex.EncodeToString(interaction.Hash), interaction.Time)
			}*/
			/*		if got, want := len(efficiencies), len(addresses); got != want {
					t.Errorf("got %d want %d", got, want)
				}*/
			for _, stats := range efficiencies {
				t.Log(stats[database.KeyStatRowsReturned], "/", stats[database.KeyStatRowsSeen])
				/*if got, want := efficiency, int64(1); got != want {
					t.Errorf("efficiency for %s: got %d, want %d", rowRange, got, want)
				}*/
			}
		})
	}
}

func getEfficiencies(efficiencies map[string]map[string]string) func(msg string, args ...any) {
	return func(msg string, args ...any) {
		stats := make(map[string]string)
		var rowRange string
		for i := 0; i < len(args); i = i + 2 {
			if args[i].(string) == database.KeyStatEfficiency {
				stats[database.KeyStatEfficiency] = fmt.Sprintf("%d", args[i+1].(int64))
			}
			if args[i].(string) == database.KeyStatRange {
				rowRange = args[i+1].(string)
			}
			if args[i].(string) == database.KeyStatRowsSeen {
				stats[database.KeyStatRowsSeen] = fmt.Sprintf("%d", args[i+1].(int64))
			}
			if args[i].(string) == database.KeyStatRowsReturned {
				stats[database.KeyStatRowsReturned] = fmt.Sprintf("%d", args[i+1].(int64))
			}
		}
		efficiencies[rowRange] = stats
	}
}
