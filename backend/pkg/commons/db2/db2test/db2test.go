package db2test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/data"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database/databasetest"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/raw"
)

func NewDataStore(t *testing.T) data.Store {
	client, admin := databasetest.NewBigTable(t)
	db, err := database.NewBigTableWithClient(context.Background(), client, admin, data.Schema)
	if err != nil {
		t.Fatal(err)
	}
	return data.NewStore(database.Wrap(db, data.Table))
}

func NewRawStore(t *testing.T) raw.Store {
	client, admin := databasetest.NewBigTable(t)
	db, err := database.NewBigTableWithClient(context.Background(), client, admin, raw.Schema)
	if err != nil {
		t.Fatal(err)
	}
	return raw.NewStore(database.Wrap(db, raw.Table))
}

func NewStores(t *testing.T) (raw.Store, data.Store) {
	return NewRawStore(t), NewDataStore(t)
}

func AddBlockToRawStore(t *testing.T, store raw.Store, endpoint string, chainID uint64, blocks []uint64) {
	var fullBlocks []raw.FullBlockData
	for _, block := range blocks {
		fullBlocks = append(fullBlocks, makeRawBlock(t, endpoint, chainID, block))
	}
	if err := store.AddBlocks(fullBlocks); err != nil {
		t.Fatal(err)
	}
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
