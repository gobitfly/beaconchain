package raw

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database/databasetest"
	"github.com/gobitfly/beaconchain/pkg/commons/types/geth"
)

const (
	chainID uint64 = 1
)

func TestBigTableClientRealCondition(t *testing.T) {
	project := os.Getenv("BIGTABLE_PROJECT")
	instance := os.Getenv("BIGTABLE_INSTANCE")
	if project == "" || instance == "" {
		t.Skip("skipping test, set BIGTABLE_PROJECT and BIGTABLE_INSTANCE")
	}

	tests := []struct {
		name  string
		block int64
	}{
		{
			name:  "test block",
			block: testBlockNumber,
		},
		{
			name:  "test block two uncles",
			block: testTwoUnclesBlockNumber,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bt, err := database.NewBigTable(project, instance, nil)
			if err != nil {
				t.Fatal(err)
			}

			rawStore := NewStore(database.Wrap(bt, Table))
			rpcClient, err := rpc.DialOptions(context.Background(), "https://foo.bar", rpc.WithHTTPClient(&http.Client{
				Transport: NewBigTableEthRaw(rawStore, chainID),
			}))
			if err != nil {
				t.Fatal(err)
			}
			ethClient := ethclient.NewClient(rpcClient)

			block, err := ethClient.BlockByNumber(context.Background(), big.NewInt(tt.block))
			if err != nil {
				t.Fatalf("BlockByNumber() error = %v", err)
			}
			if got, want := block.Number().Int64(), tt.block; got != want {
				t.Errorf("got %v, want %v", got, want)
			}

			receipts, err := ethClient.BlockReceipts(context.Background(), rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(tt.block)))
			if err != nil {
				t.Fatalf("BlockReceipts() error = %v", err)
			}
			if len(block.Transactions()) != 0 && len(receipts) == 0 {
				t.Errorf("receipts should not be empty")
			}
			var traces []geth.Trace
			if err := rpcClient.Call(&traces, "debug_traceBlockByNumber", hexutil.EncodeBig(block.Number()), geth.Tracer); err != nil {
				t.Fatalf("debug_traceBlockByNumber() error = %v", err)
			}
			if len(block.Transactions()) != 0 && len(traces) == 0 {
				t.Errorf("traces should not be empty")
			}
		})
	}
}

func TestBigTableClient(t *testing.T) {
	tests := []struct {
		name  string
		block FullBlockData
	}{
		{
			name:  "test block",
			block: testFullBlock,
		},
		{
			name:  "two uncles",
			block: testTwoUnclesFullBlock,
		},
	}

	client, admin := databasetest.NewBigTable(t)
	bt, err := database.NewBigTableWithClient(context.Background(), client, admin, Schema)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawStore := NewStore(database.Wrap(bt, Table))
			if err := rawStore.AddBlocks([]FullBlockData{tt.block}); err != nil {
				t.Fatal(err)
			}

			rpcClient, err := rpc.DialOptions(context.Background(), "https://foo.bar", rpc.WithHTTPClient(&http.Client{
				Transport: NewBigTableEthRaw(WithCache(rawStore), tt.block.ChainID),
			}))
			if err != nil {
				t.Fatal(err)
			}
			ethClient := ethclient.NewClient(rpcClient)

			block, err := ethClient.BlockByNumber(context.Background(), big.NewInt(tt.block.BlockNumber))
			if err != nil {
				t.Fatalf("BlockByNumber() error = %v", err)
			}
			if got, want := block.Number().Int64(), tt.block.BlockNumber; got != want {
				t.Errorf("got %v, want %v", got, want)
			}

			receipts, err := ethClient.BlockReceipts(context.Background(), rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(tt.block.BlockNumber)))
			if err != nil {
				t.Fatalf("BlockReceipts() error = %v", err)
			}
			if len(block.Transactions()) != 0 && len(receipts) == 0 {
				t.Errorf("receipts should not be empty")
			}

			var traces []geth.Trace
			if err := rpcClient.Call(&traces, "debug_traceBlockByNumber", hexutil.EncodeBig(block.Number()), geth.Tracer); err != nil {
				t.Fatalf("debug_traceBlockByNumber() error = %v", err)
			}
			if len(block.Transactions()) != 0 && len(traces) == 0 {
				t.Errorf("traces should not be empty")
			}
		})
	}
}

func TestWithFallback(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedErr bool
	}{
		{
			name: "fallback with json.SyntaxError",
			err:  &json.SyntaxError{},
		},
		{
			name: "fallback with ErrNotFoundInCache",
			err:  ErrNotFoundInCache,
		},
		{
			name: "fallback with ErrMethodNotSupported",
			err:  ErrMethodNotSupported,
		},
		{
			name: "fallback with database.ErrNotFound",
			err:  database.ErrNotFound,
		},
		{
			name:        "return err for other",
			err:         fmt.Errorf("some error"),
			expectedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roundTripper := NewWithFallback(
				stubRoundTripper{err: tt.err},
				stubRoundTripper{resp: &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}},
			)
			resp, err := roundTripper.RoundTrip(httptest.NewRequest(http.MethodGet, "/", nil))
			if tt.expectedErr {
				if err == nil {
					t.Fatal("expecting error, got none")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("got %v, want %v", resp.StatusCode, http.StatusOK)
			}
		})
	}
}

type stubRoundTripper struct {
	err  error
	resp *http.Response
}

func (stub stubRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	if stub.err != nil {
		return nil, stub.err
	}
	return stub.resp, nil
}

func TestBigTableClientWithFallbackRealCondition(t *testing.T) {
	node := os.Getenv("ETH1_ERIGON_ENDPOINT")
	if node == "" {
		t.Skip("skipping test, set ETH1_ERIGON_ENDPOINT")
	}

	tests := []struct {
		name  string
		block FullBlockData
	}{
		{
			name:  "test block",
			block: testFullBlock,
		},
	}

	client, admin := databasetest.NewBigTable(t)
	bt, err := database.NewBigTableWithClient(context.Background(), client, admin, Schema)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawStore := NewStore(database.Wrap(bt, Table))

			rpcClient, err := rpc.DialOptions(context.Background(), node, rpc.WithHTTPClient(&http.Client{
				Transport: NewWithFallback(NewBigTableEthRaw(rawStore, tt.block.ChainID), http.DefaultTransport),
			}))
			if err != nil {
				t.Fatal(err)
			}
			ethClient := ethclient.NewClient(rpcClient)

			balance, err := ethClient.BalanceAt(context.Background(), common.Address{}, big.NewInt(tt.block.BlockNumber))
			if err != nil {
				t.Fatal(err)
			}
			if balance == nil {
				t.Errorf("empty balance")
			}

			block, err := ethClient.BlockByNumber(context.Background(), big.NewInt(tt.block.BlockNumber))
			if err != nil {
				t.Fatalf("BlockByNumber() error = %v", err)
			}
			if got, want := block.Number().Int64(), tt.block.BlockNumber; got != want {
				t.Errorf("got %v, want %v", got, want)
			}

			receipts, err := ethClient.BlockReceipts(context.Background(), rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(tt.block.BlockNumber)))
			if err != nil {
				t.Fatalf("BlockReceipts() error = %v", err)
			}
			if len(block.Transactions()) != 0 && len(receipts) == 0 {
				t.Errorf("receipts should not be empty")
			}

			var traces []geth.Trace
			if err := rpcClient.Call(&traces, "debug_traceBlockByNumber", hexutil.EncodeBig(block.Number()), geth.Tracer); err != nil {
				t.Fatalf("debug_traceBlockByNumber() error = %v", err)
			}
			if len(block.Transactions()) != 0 && len(traces) == 0 {
				t.Errorf("traces should not be empty")
			}
		})
	}
}

func benchmarkBlockRetrieval(b *testing.B, ethClient *ethclient.Client, rpcClient *rpc.Client) {
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		blockTestNumber := int64(20978000 + b.N)
		_, err := ethClient.BlockByNumber(context.Background(), big.NewInt(blockTestNumber))
		if err != nil {
			b.Fatalf("BlockByNumber() error = %v", err)
		}

		if _, err := ethClient.BlockReceipts(context.Background(), rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(blockTestNumber))); err != nil {
			b.Fatalf("BlockReceipts() error = %v", err)
		}

		var traces []geth.Trace
		if err := rpcClient.Call(&traces, "debug_traceBlockByNumber", hexutil.EncodeBig(big.NewInt(blockTestNumber)), geth.Tracer); err != nil {
			b.Fatalf("debug_traceBlockByNumber() error = %v", err)
		}
	}
}

func BenchmarkErigonNode(b *testing.B) {
	node := os.Getenv("ETH1_ERIGON_ENDPOINT")
	if node == "" {
		b.Skip("skipping test, please set ETH1_ERIGON_ENDPOINT")
	}

	rpcClient, err := rpc.DialOptions(context.Background(), node)
	if err != nil {
		b.Fatal(err)
	}

	benchmarkBlockRetrieval(b, ethclient.NewClient(rpcClient), rpcClient)
}

func BenchmarkRawBigTable(b *testing.B) {
	project := os.Getenv("BIGTABLE_PROJECT")
	instance := os.Getenv("BIGTABLE_INSTANCE")
	if project == "" || instance == "" {
		b.Skip("skipping test, set BIGTABLE_PROJECT and BIGTABLE_INSTANCE")
	}

	bt, err := database.NewBigTable(project, instance, nil)
	if err != nil {
		b.Fatal(err)
	}

	rawStore := WithCache(NewStore(database.Wrap(bt, Table)))
	rpcClient, err := rpc.DialOptions(context.Background(), "https://foo.bar", rpc.WithHTTPClient(&http.Client{
		Transport: NewBigTableEthRaw(rawStore, chainID),
	}))
	if err != nil {
		b.Fatal(err)
	}

	benchmarkBlockRetrieval(b, ethclient.NewClient(rpcClient), rpcClient)
}

func BenchmarkAll(b *testing.B) {
	b.Run("BenchmarkErigonNode", func(b *testing.B) {
		BenchmarkErigonNode(b)
	})
	b.Run("BenchmarkRawBigTable", func(b *testing.B) {
		BenchmarkRawBigTable(b)
	})
}
