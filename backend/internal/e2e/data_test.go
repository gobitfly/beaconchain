package e2e

import (
	"context"
	"encoding/hex"
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
	_, usdt := backend.DeployToken(t, "usdt", "usdt", backend.BankAccount.From)

	transform := indexer.NewTransformer(indexer.NoopCache{})
	indexer := indexer.New(store, transform.Tx, transform.ERC20)

	client, err := rpc.NewErigonClient(backend.Endpoint)
	if err != nil {
		t.Fatal(err)
	}

	var addresses []common.Address
	for i := 0; i < 10; i++ {
		temp := th.CreateEOA(t)
		addresses = append(addresses, temp.From)
		for j := 0; j < 25; j++ {
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

	t.Run("get interactions", func(t *testing.T) {
		efficiencies := make(map[string]int64)
		interactions, _, err := store.Get(addresses, nil, 50, data.WithDatabaseStats(getEfficiencies(efficiencies)))
		if err != nil {
			t.Fatal(err)
		}
		for _, interaction := range interactions {
			t.Log(interaction.Type, interaction.ChainID, "0x"+interaction.From, "0x"+interaction.To, "0x"+hex.EncodeToString(interaction.Hash), interaction.Time)
		}
		if got, want := len(efficiencies), len(addresses); got != want {
			t.Errorf("got %d want %d", got, want)
		}
		for rowRange, efficiency := range efficiencies {
			if got, want := efficiency, int64(1); got != want {
				t.Errorf("efficiency for %s: got %d, want %d", rowRange, got, want)
			}
		}
	})
}

func getEfficiencies(efficiencies map[string]int64) func(msg string, args ...any) {
	return func(msg string, args ...any) {
		var efficiency int64
		var rowRange string
		for i := 0; i < len(args); i = i + 2 {
			if args[i].(string) == database.KeyStatEfficiency {
				efficiency = args[i+1].(int64)
			}
			if args[i].(string) == database.KeyStatRange {
				rowRange = args[i+1].(string)
			}
		}
		efficiencies[rowRange] = efficiency
	}
}
