package rawtest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gobitfly/beaconchain/internal/th"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/databasetest"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/raw"
)

func NewRandSeededStore(t *testing.T) (raw.Store, *th.BlockchainBackend) {
	t.Helper()
	client, admin := databasetest.NewBigTable(t)
	bt, err := database.NewBigTableWithClient(context.Background(), client, admin, raw.Schema)
	if err != nil {
		t.Fatal(err)
	}

	db := raw.NewStore(database.Wrap(bt, raw.BlocksRawTable))

	backend := th.NewBackend(t)
	for i := 0; i < 10; i++ {
		temp := th.CreateEOA(t)
		backend.FundOneEther(t, temp.From)
	}
	lastBlock, err := backend.Client().BlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var blocks []raw.FullBlockData
	for i := uint64(0); i <= lastBlock; i++ {
		blocks = append(blocks, makeRawBlock(t, backend.Endpoint, uint64(backend.ChainID), i))
	}
	if err := db.AddBlocks(blocks); err != nil {
		t.Fatal(err)
	}

	return db, backend
}

func makeRawBlock(t *testing.T, endpoint string, chainID uint64, block uint64) raw.FullBlockData {
	getReceipts := `{"jsonrpc":"2.0","method":"eth_getBlockReceipts","params":["0x%x"],"id":%d}`
	getBlock := `{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", true],"id":%d}`
	getTraces := `{"jsonrpc":"2.0","method":"debug_traceBlockByNumber","params":["0x%x", {"tracer": "callTracer"}],"id":%d}`
	id := 1

	return raw.FullBlockData{
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
