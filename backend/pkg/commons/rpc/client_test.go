package rpc

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/gobitfly/beaconchain/internal/th"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/db2test"
)

func TestCmpNodeAndRawStoreClient(t *testing.T) {
	backend := th.NewBackend(t)

	nodeClient, err := NewNodeClient(backend.Endpoint)
	if err != nil {
		t.Fatal(err)
	}

	rawStore := db2test.NewRawStore(t)
	rawStoreClient := RawStoreClient{
		chainID: big.NewInt(int64(backend.ChainID)),
		store:   rawStore,
	}

	for i := 0; i < 1; i++ {
		receiver := th.CreateEOA(t).From
		tx := backend.MakeTx(t, backend.BankAccount, &receiver, big.NewInt(1), nil)
		if err := backend.Client().SendTransaction(context.Background(), tx); err != nil {
			t.Fatal(err)
		}
	}
	backend.Commit()
	lastBlock, err := backend.Client().BlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	db2test.AddBlockToRawStore(t, rawStore, backend.Endpoint, uint64(backend.ChainID), []uint64{lastBlock})
	rawStoreBlock, err := rawStoreClient.GetBlock(int64(lastBlock), "geth")
	if err != nil {
		t.Fatal(err)
	}

	nodeBlock, err := nodeClient.GetBlock(int64(lastBlock), "geth")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(nodeBlock, rawStoreBlock, protocmp.Transform()); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}
