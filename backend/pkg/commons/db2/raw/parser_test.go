package raw

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/gobitfly/beaconchain/internal/th"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database/databasetest"
)

func TestRawWithBackend(t *testing.T) {
	raw, backend := newRandSeededStore(t)
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
		block, receipts, _, err := GethParse(b)
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
	db := NewStore(client)
	block, err := db.ReadBlockByNumber(1, 6008149)
	if err != nil {
		panic(err)
	}

	ethBlock, receipts, traces, err := GethParse(block)
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

func newRandSeededStore(t *testing.T) (Store, *th.BlockchainBackend) {
	t.Helper()
	client, admin := databasetest.NewBigTable(t)
	bt, err := database.NewBigTableWithClient(context.Background(), client, admin, Schema)
	if err != nil {
		t.Fatal(err)
	}

	db := NewStore(database.Wrap(bt, Table))

	backend := th.NewBackend(t)
	for i := 0; i < 10; i++ {
		temp := th.CreateEOA(t)
		backend.FundOneEther(t, temp.From)
	}
	lastBlock, err := backend.Client().BlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var blocks []FullBlockData
	for i := uint64(0); i <= lastBlock; i++ {
		blocks = append(blocks, makeRawBlock(t, backend.Endpoint, uint64(backend.ChainID), i))
	}
	if err := db.AddBlocks(blocks); err != nil {
		t.Fatal(err)
	}

	return db, backend
}

func makeRawBlock(t *testing.T, endpoint string, chainID uint64, block uint64) FullBlockData {
	getReceipts := `{"jsonrpc":"2.0","method":"eth_getBlockReceipts","params":["0x%x"],"id":%d}`
	getBlock := `{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", true],"id":%d}`
	getTraces := `{"jsonrpc":"2.0","method":"debug_traceBlockByNumber","params":["0x%x", {"tracer": "callTracer"}],"id":%d}`
	id := 1

	return FullBlockData{
		ChainID:          chainID,
		BlockNumber:      int64(block),
		BlockHash:        nil,
		BlockUnclesCount: 0,
		BlockTxs:         nil,
		Block:            httpCall(t, endpoint, fmt.Sprintf(getBlock, block, id)),
		Receipts:         httpCall(t, endpoint, fmt.Sprintf(getReceipts, block, id)),
		Traces:           httpCall(t, endpoint, fmt.Sprintf(getTraces, block, id)),
		Uncles:           nil,
	}
}

func httpCall(t *testing.T, endpoint string, body string) []byte {
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return b
}
