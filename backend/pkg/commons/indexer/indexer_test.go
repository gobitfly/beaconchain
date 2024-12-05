package indexer

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/gobitfly/beaconchain/internal/th"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/data"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database/databasetest"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

func TestIndexer(t *testing.T) {
	btClient, btAdmin := databasetest.NewBigTable(t)

	bigtable, err := database.NewBigTableWithClient(context.Background(), btClient, btAdmin, data.Schema)
	if err != nil {
		t.Fatal(err)
	}

	transform := NewTransformer(NoopCache{})

	indexer := New(data.NewStore(database.Wrap(bigtable, data.Table)), transform.Tx, transform.ERC20)

	backend := th.NewBackend(t)
	client, err := rpc.NewErigonClient(backend.Endpoint)
	if err != nil {
		t.Fatal(err)
	}
	_, usdt := backend.DeployToken(t, "usdt", "usdt", backend.BankAccount.From)

	for i := 0; i < 10; i++ {
		temp := th.CreateEOA(t)
		if err := backend.Client().SendTransaction(context.Background(), backend.MakeTx(t, backend.BankAccount, &temp.From, big.NewInt(1), nil)); err != nil {
			t.Fatal(err)
		}
		if _, err := usdt.Mint(backend.BankAccount.TransactOpts, temp.From, big.NewInt(1)); err != nil {
			t.Fatal(err)
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

	rows, err := bigtable.Read(data.Table, "")
	if err != nil {
		t.Fatal(err)
	}

	if len(rows) == 0 {
		t.Errorf("no rows found")
	}
}
