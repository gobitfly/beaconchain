package db2

import (
	"context"
	"math/big"
	"os"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/raw"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/rawtest"
)

func TestRawWithBackend(t *testing.T) {
	raw, backend := rawtest.NewRandSeededStore(t)
	blocks, err := raw.ReadBlocksByNumber(uint64(backend.ChainID), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range blocks {
		expectedBlock, err := backend.Client().BlockByNumber(context.Background(), big.NewInt(b.BlockNumber))
		if err != nil {
			t.Fatal(err)
		}
		expectedReceipts, err := backend.Client().BlockReceipts(context.Background(), rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(b.BlockNumber)))
		if err != nil {
			t.Fatal(err)
		}
		block, receipts, _, err := EthParse(b)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := block.Number().String(), expectedBlock.Number().String(); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := block.Hash().String(), expectedBlock.Hash().String(); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := block.TxHash().String(), expectedBlock.TxHash().String(); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := block.UncleHash().String(), expectedBlock.UncleHash().String(); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := block.ReceiptHash().String(), expectedBlock.ReceiptHash().String(); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if len(expectedReceipts) != 0 {
			if got, want := receipts, expectedReceipts; !reflect.DeepEqual(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		}
	}
}

func TestRawRemoteRealCondition(t *testing.T) {
	remote := os.Getenv("REMOTE_URL")
	if remote == "" {
		t.Skip("skipping test, set REMOTE_URL")
	}

	client := database.NewRemoteClient(remote)
	db := raw.NewStore(client)
	block, err := db.ReadBlockByNumber(1, 6008149)
	if err != nil {
		panic(err)
	}

	ethBlock, receipts, traces, err := EthParse(block)
	if err != nil {
		t.Errorf("failed to parse block: %v", err)
	}
	for i, transaction := range ethBlock.Transactions() {
		if got, want := receipts[i].TxHash, transaction.Hash(); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		if got, want := traces[i].TxHash, transaction.Hash().Hex(); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}
