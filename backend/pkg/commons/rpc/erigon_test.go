package rpc

import (
	"bytes"
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"

	"github.com/gobitfly/beaconchain/internal/th"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

var (
	johnAddress = common.HexToAddress("0x6d2e03b7EfFEae98BD302A9F836D0d6Ab0002766")
	tomAddress  = common.HexToAddress("0x10e4597ff93cbee194f4879f8f1d54a370db6969")
)

// TestGetInternalTxs tests the getInternalTxs function
// which returns the internal transactions of a transaction at a given position
func TestGetInternalTxs(t *testing.T) {
	tests := []struct {
		name       string
		traceIndex int
		traces     []*Eth1InternalTransactionWithPosition
		txPosition int
		expected   []*types.Eth1InternalTransaction
	}{
		{
			name:       "single internal transaction with tx position",
			traceIndex: 0,
			traces: []*Eth1InternalTransactionWithPosition{
				{
					txPosition: 0,
					Eth1InternalTransaction: types.Eth1InternalTransaction{
						From:  johnAddress.Bytes(),
						To:    tomAddress.Bytes(),
						Value: big.NewInt(100).Bytes(),
					},
				},
			},
			txPosition: 0,
			expected: []*types.Eth1InternalTransaction{
				{
					From:  johnAddress.Bytes(),
					To:    tomAddress.Bytes(),
					Value: big.NewInt(100).Bytes(),
				},
			},
		},
		{
			name:       "two internal transactions with tx position",
			traceIndex: 1,
			traces: []*Eth1InternalTransactionWithPosition{
				{
					txPosition: 0,
					Eth1InternalTransaction: types.Eth1InternalTransaction{
						From:  johnAddress.Bytes(),
						To:    tomAddress.Bytes(),
						Value: big.NewInt(100).Bytes(),
					},
				},
				{
					txPosition: 1,
					Eth1InternalTransaction: types.Eth1InternalTransaction{
						From:  tomAddress.Bytes(),
						To:    johnAddress.Bytes(),
						Value: big.NewInt(200).Bytes(),
					},
				},
			},
			txPosition: 1,
			expected: []*types.Eth1InternalTransaction{
				{
					From:  tomAddress.Bytes(),
					To:    johnAddress.Bytes(),
					Value: big.NewInt(200).Bytes(),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getInternalTxs(tt.traceIndex, tt.traces, tt.txPosition)
			if len(result) != len(tt.expected) {
				t.Fatalf("got %v internal transactions, want %v internal transactions", len(result), len(tt.expected))
			}
			for i, itx := range result {
				if !bytes.Equal(itx.From, tt.expected[i].From) {
					t.Errorf("got From %v, want %v", itx.From, tt.expected[i].From)
				}
				if !bytes.Equal(itx.To, tt.expected[i].To) {
					t.Errorf("got To %v, want %v", itx.To, tt.expected[i].To)
				}
				if !bytes.Equal(itx.Value, tt.expected[i].Value) {
					t.Errorf("got Value %v, want %v", itx.Value, tt.expected[i].Value)
				}
			}
		})
	}
}

// TestGetBlobVersionedHashes tests the getBlobVersionedHashes function
// which extracts the blob versioned hashes from a transaction
func TestGetBlobVersionedHashes(t *testing.T) {
	tests := []struct {
		name     string
		tx       *gethtypes.Transaction
		expected [][]byte
	}{
		{
			name: "transaction with blob versioned hashes",
			tx: gethtypes.NewTx(&gethtypes.BlobTx{
				BlobHashes: []common.Hash{common.HexToHash("0x12345"), common.HexToHash("0x23456")},
			}),
			expected: [][]byte{common.HexToHash("0x12345").Bytes(), common.HexToHash("0x23456").Bytes()},
		},
		{
			name:     "transaction without blob versioned hashes",
			tx:       gethtypes.NewTransaction(0, johnAddress, big.NewInt(100), 100000, big.NewInt(10), []byte{}),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBlobVersionedHashes(tt.tx)
			if len(result) != len(tt.expected) {
				t.Fatalf("got %v hashes, want %v hashes", len(result), len(tt.expected))
			}
			for i, hash := range result {
				if !bytes.Equal(hash, tt.expected[i]) {
					t.Errorf("got Hash %v, want %v", hash, tt.expected[i])
				}
			}
		})
	}
}

// TestGetMaxFeePerBlobGas tests the getMaxFeePerBlobGas function
// which extracts the max fee per blob gas from a transaction
func TestGetMaxFeePerBlobGas(t *testing.T) {
	tests := []struct {
		name     string
		tx       *gethtypes.Transaction
		expected []byte
	}{
		{
			name: "transaction with blob gas fee cap",
			tx: gethtypes.NewTx(&gethtypes.BlobTx{
				BlobFeeCap: uint256.NewInt(10),
			}),
			expected: big.NewInt(10).Bytes(),
		},
		{
			name:     "transaction with no blob gas fee cap",
			tx:       gethtypes.NewTransaction(0, johnAddress, big.NewInt(100), 100000, big.NewInt(10), []byte{}),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMaxFeePerBlobGas(tt.tx)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetBlobGasPrice tests the getBlobGasPrice function
// which extracts the blob gas price from a receipt
func TestGetBlobGasPrice(t *testing.T) {
	tests := []struct {
		name     string
		receipt  *gethtypes.Receipt
		expected []byte
	}{
		{
			name:     "receipt with blob gas price",
			receipt:  &gethtypes.Receipt{BlobGasPrice: big.NewInt(10)},
			expected: big.NewInt(10).Bytes(),
		},
		{
			name:     "receipt with no blob gas price",
			receipt:  &gethtypes.Receipt{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBlobGasPrice(tt.receipt)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetBaseFee tests the getBaseFee function
// which extracts the base fee from a block
func TestGetBaseFee(t *testing.T) {
	tests := []struct {
		name     string
		block    *gethtypes.Block
		expected []byte
	}{
		{
			name:     "block with base fee",
			block:    gethtypes.NewBlockWithHeader(&gethtypes.Header{BaseFee: big.NewInt(10)}),
			expected: big.NewInt(10).Bytes(),
		},
		{
			name:     "block with no base fee",
			block:    gethtypes.NewBlockWithHeader(&gethtypes.Header{}),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBaseFee(tt.block)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetBlobGasUsed tests the getBlobGasUsed function
// which extracts the blob gas used from a block
func TestGetBlobGasUsed(t *testing.T) {
	tests := []struct {
		name     string
		block    *gethtypes.Block
		expected uint64
	}{
		{
			name:     "block with blob gas used",
			block:    gethtypes.NewBlockWithHeader(&gethtypes.Header{BlobGasUsed: toPtr(uint64(100))}),
			expected: 100,
		},
		{
			name:     "block with no blob gas used",
			block:    gethtypes.NewBlockWithHeader(&gethtypes.Header{}),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBlobGasUsed(tt.block)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetExcessBlobGas tests the getExcessBlobGas function
// which extracts the excess blob gas from a block
func TestGetExcessBlobGas(t *testing.T) {
	tests := []struct {
		name     string
		block    *gethtypes.Block
		expected uint64
	}{
		{
			name:     "block with excess blob gas",
			block:    gethtypes.NewBlockWithHeader(&gethtypes.Header{ExcessBlobGas: toPtr(uint64(100))}),
			expected: 100,
		},
		{
			name:     "block with no excess blob gas",
			block:    gethtypes.NewBlockWithHeader(&gethtypes.Header{}),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getExcessBlobGas(tt.block)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetSender tests the getSender function
// which extracts the sender of a transaction
func TestGetSender(t *testing.T) {
	backend := th.NewBackend(t)
	defer backend.Close()
	temp := th.CreateEOA(t)
	if err := backend.Client().SendTransaction(context.Background(), backend.MakeTx(t, backend.BankAccount, &temp.From, big.NewInt(1), nil)); err != nil {
		t.Fatal(err)
	}
	backend.Commit()

	lastBlock, err := backend.Client().BlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewErigonClient(backend.Endpoint)
	if err != nil {
		t.Fatal(err)
	}

	block, err := client.ethClient.BlockByNumber(context.Background(), big.NewInt(int64(lastBlock)))
	if err != nil {
		t.Fatal(err)
	}

	if len(block.Transactions()) == 0 {
		t.Fatal("no transactions in the block")
	}

	tests := []struct {
		name        string
		tx          *gethtypes.Transaction
		blockHash   common.Hash
		txPosition  int
		expected    []byte
		expectError bool
	}{
		{
			name:       "valid transaction with sender",
			tx:         block.Transactions()[0],
			blockHash:  block.Hash(),
			txPosition: 0,
			expected:   backend.BankAccount.From.Bytes(),
		},
		{
			name:       "transaction with error retrieving sender",
			tx:         block.Transactions()[0],
			blockHash:  common.HexToHash("0x456"),
			txPosition: 1,
			expected:   common.HexToAddress("abababababababababababababababababababab").Bytes(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.getSender(tt.tx, tt.blockHash, tt.txPosition)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("got %x, want %x", result, tt.expected)
			}
		})
	}
}

// TestGetBlockHash tests the getBlockHash function
// which extracts the hash of a block from receipts or from RPC call
func TestGetBlockHash(t *testing.T) {
	backend := th.NewBackend(t)
	defer backend.Close()
	temp := th.CreateEOA(t)
	if err := backend.Client().SendTransaction(context.Background(), backend.MakeTx(t, backend.BankAccount, &temp.From, big.NewInt(1), nil)); err != nil {
		t.Fatal(err)
	}
	backend.Commit()

	lastBlock, err := backend.Client().BlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewErigonClient(backend.Endpoint)
	if err != nil {
		t.Fatal(err)
	}

	block, err := client.ethClient.BlockByNumber(context.Background(), new(big.Int).SetUint64(lastBlock))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		blockNumber uint64
		receipts    []*gethtypes.Receipt
		expected    common.Hash
		expectError bool
	}{
		{
			name:        "block hash from receipts",
			blockNumber: lastBlock,
			receipts: []*gethtypes.Receipt{
				{BlockHash: common.HexToHash("0x123")},
			},
			expected: common.HexToHash("0x123"),
		},
		{
			name:        "block hash from RPC",
			blockNumber: lastBlock,
			expected:    block.Hash(),
		},
		{
			name:        "RPC call error for block that doesn't exist",
			blockNumber: ^uint64(0), // max uint, block doesn't exist
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.getBlockHash(tt.blockNumber, tt.receipts)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// toPtr returns a pointer
func toPtr[T any](i T) *T {
	return &i
}
