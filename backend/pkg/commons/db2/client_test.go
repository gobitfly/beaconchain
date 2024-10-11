package db2

import (
	"context"
	"math/big"
	"net/http"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/store"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/storetest"
)

func TestBigTableClient(t *testing.T) {
	tests := []struct {
		name   string
		number int64
		block  FullBlockRawData
	}{
		{
			name:   "test block",
			number: testBlockNumber,
			block:  testFullBlock,
		},
	}

	client, admin := storetest.NewBigTable(t)
	bg, err := store.NewBigTableWithClient(context.Background(), client, admin, raw)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawStore := NewRawStore(store.Wrap(bg, blocRawTable, ""))
			if err := rawStore.AddBlocks([]FullBlockRawData{tt.block}); err != nil {
				t.Fatal(err)
			}

			rpcClient, err := rpc.DialOptions(context.Background(), "http://foo.bar", rpc.WithHTTPClient(&http.Client{
				Transport: NewBigTableEthRaw(http.DefaultTransport, rawStore),
			}))
			if err != nil {
				t.Fatal(err)
			}
			ethClient := ethclient.NewClient(rpcClient)

			block, err := ethClient.BlockByNumber(context.Background(), big.NewInt(tt.number))
			if err != nil {
				t.Fatalf("BlockByNumber() error = %v", err)
			}
			if got, want := block.Number().Int64(), tt.number; got != want {
				t.Errorf("got %v, want %v", got, want)
			}

			receipts, err := ethClient.BlockReceipts(context.Background(), rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(tt.number)))
			if err != nil {
				t.Fatalf("BlockReceipts() error = %v", err)
			}
			if len(receipts) == 0 {
				t.Errorf("receipts should not be empty")
			}

			var traces []GethTraceCallResultWrapper
			if err := rpcClient.Call(&traces, "debug_traceBlockByNumber", hexutil.EncodeBig(block.Number()), gethTracerArg); err != nil {
				t.Fatalf("debug_traceBlockByNumber() error = %v", err)
			}
			if len(traces) == 0 {
				t.Errorf("traces should not be empty")
			}
		})
	}
}

// TODO import those 3 from somewhere
var gethTracerArg = map[string]string{
	"tracer": "callTracer",
}

type GethTraceCallResultWrapper struct {
	Result *GethTraceCallResult `json:"result,omitempty"`
}

type GethTraceCallResult struct {
	TransactionPosition int                    `json:"transaction_position,omitempty"`
	Time                string                 `json:"time,omitempty"`
	GasUsed             string                 `json:"gas_used,omitempty"`
	From                common.Address         `json:"from,omitempty"`
	To                  common.Address         `json:"to,omitempty"`
	Value               string                 `json:"value,omitempty"`
	Gas                 string                 `json:"gas,omitempty"`
	Input               string                 `json:"input,omitempty"`
	Output              string                 `json:"output,omitempty"`
	Error               string                 `json:"error,omitempty"`
	Type                string                 `json:"type,omitempty"`
	Calls               []*GethTraceCallResult `json:"calls,omitempty"`
}
