package rpc

import (
	"bytes"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gobitfly/beaconchain/internal/th"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	aliceAddress = common.HexToAddress("0x95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5")
	bobAddress   = common.HexToAddress("0x388C818CA8B9251b393131C08a736A67ccB19297")
	token        = common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	token2       = common.HexToAddress("0xabcdef1234567890abcdef1234567890abcdef12")
)

// TestGetReceiver tests the getReceiver function, which extracts the recipient
// address from a transaction.
func TestGetReceiver(t *testing.T) {
	tests := []struct {
		name     string
		tx       *gethtypes.Transaction
		expected []byte
	}{
		{
			name:     "transaction with recipient",
			tx:       gethtypes.NewTransaction(0, aliceAddress, big.NewInt(100), 100000, big.NewInt(10), []byte{}),
			expected: aliceAddress.Bytes(),
		},
		{
			name:     "contract creation transaction",
			tx:       gethtypes.NewContractCreation(0, big.NewInt(100), 100000, big.NewInt(10), []byte{}),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getReceiver(tt.tx)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetBlockWithdrawals tests the getBlockWithdrawals function, which
// converts a list of withdrawals from a block into a list of Eth1Withdrawal type.
func TestGetBlockWithdrawals(t *testing.T) {
	tests := []struct {
		name           string
		blkWithdrawals gethtypes.Withdrawals
		expected       []*types.Eth1Withdrawal
	}{
		{
			name: "block with one withdrawal",
			blkWithdrawals: gethtypes.Withdrawals{
				{
					Index:     1,
					Validator: 2,
					Address:   aliceAddress,
					Amount:    100,
				},
			},
			expected: []*types.Eth1Withdrawal{
				{
					Index:          1,
					ValidatorIndex: 2,
					Address:        aliceAddress.Bytes(),
					Amount:         big.NewInt(100).Bytes(),
				},
			},
		},
		{
			name: "block with two withdrawals",
			blkWithdrawals: gethtypes.Withdrawals{
				{
					Index:     1,
					Validator: 2,
					Address:   aliceAddress,
					Amount:    100,
				},
				{
					Index:     2,
					Validator: 3,
					Address:   bobAddress,
					Amount:    200,
				},
			},
			expected: []*types.Eth1Withdrawal{
				{
					Index:          1,
					ValidatorIndex: 2,
					Address:        aliceAddress.Bytes(),
					Amount:         big.NewInt(100).Bytes(),
				},
				{
					Index:          2,
					ValidatorIndex: 3,
					Address:        bobAddress.Bytes(),
					Amount:         big.NewInt(200).Bytes(),
				},
			},
		},
		{
			name:           "block with no withdrawals",
			blkWithdrawals: gethtypes.Withdrawals{},
			expected:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBlockWithdrawals(tt.blkWithdrawals)
			if len(result) != len(tt.expected) {
				t.Fatalf("got %v withdrawals, want %v withdrawals", len(result), len(tt.expected))
			}
			for i, withdrawal := range result {
				if withdrawal.Index != tt.expected[i].Index {
					t.Errorf("got Index %v, want %v", withdrawal.Index, tt.expected[i].Index)
				}
				if withdrawal.ValidatorIndex != tt.expected[i].ValidatorIndex {
					t.Errorf("got ValidatorIndex %v, want %v", withdrawal.ValidatorIndex, tt.expected[i].ValidatorIndex)
				}
				if !bytes.Equal(withdrawal.Address, tt.expected[i].Address) {
					t.Errorf("got Address %v, want %v", withdrawal.Address, tt.expected[i].Address)
				}
				if !bytes.Equal(withdrawal.Amount, tt.expected[i].Amount) {
					t.Errorf("got Amount %v, want %v", withdrawal.Amount, tt.expected[i].Amount)
				}
			}
		})
	}
}

// TestGetLogsFromReceipts tests the getLogsFromReceipts function, which converts
// a list of logs from a receipt into a list of Eth1Log type
func TestGetLogsFromReceipts(t *testing.T) {
	tests := []struct {
		name     string
		logs     []*gethtypes.Log
		expected []*types.Eth1Log
	}{
		{
			name: "receipt with one log",
			logs: []*gethtypes.Log{
				{
					Address: aliceAddress,
					Topics:  []common.Hash{common.HexToHash("0x123"), common.HexToHash("0x234")},
					Data:    []byte("data"),
					Removed: false,
				},
			},
			expected: []*types.Eth1Log{
				{
					Address: aliceAddress.Bytes(),
					Topics:  [][]byte{common.HexToHash("0x123").Bytes(), common.HexToHash("0x234").Bytes()},
					Data:    []byte("data"),
					Removed: false,
				},
			},
		},
		{
			name: "receipt with two logs",
			logs: []*gethtypes.Log{
				{
					Address: aliceAddress,
					Topics:  []common.Hash{common.HexToHash("0x123"), common.HexToHash("0x234")},
					Data:    []byte("data"),
					Removed: false,
				},
				{
					Address: bobAddress,
					Topics:  []common.Hash{common.HexToHash("0x123"), common.HexToHash("0x234")},
					Data:    []byte("data"),
					Removed: false,
				},
			},
			expected: []*types.Eth1Log{
				{
					Address: aliceAddress.Bytes(),
					Topics:  [][]byte{common.HexToHash("0x123").Bytes(), common.HexToHash("0x234").Bytes()},
					Data:    []byte("data"),
					Removed: false,
				},
				{
					Address: bobAddress.Bytes(),
					Topics:  [][]byte{common.HexToHash("0x123").Bytes(), common.HexToHash("0x234").Bytes()},
					Data:    []byte("data"),
					Removed: false,
				},
			},
		},
		{
			name:     "receipt with no logs",
			logs:     []*gethtypes.Log{},
			expected: nil,
		},
		{
			name: "receipt with logs but no topics",
			logs: []*gethtypes.Log{
				{
					Address: aliceAddress,
					Topics:  []common.Hash{},
					Data:    []byte("data"),
					Removed: false,
				},
			},
			expected: []*types.Eth1Log{
				{
					Address: aliceAddress.Bytes(),
					Topics:  [][]byte{},
					Data:    []byte("data"),
					Removed: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getLogsFromReceipts(tt.logs)
			if len(result) != len(tt.expected) {
				t.Fatalf("got %v logs, want %v logs", len(result), len(tt.expected))
			}
			for i, log := range result {
				if !bytes.Equal(log.Address, tt.expected[i].Address) {
					t.Errorf("got Address %v, want %v", log.Address, tt.expected[i].Address)
				}
				if !bytes.Equal(log.Data, tt.expected[i].Data) {
					t.Errorf("got Data %v, want %v", log.Data, tt.expected[i].Data)
				}
				if log.Removed != tt.expected[i].Removed {
					t.Errorf("got Removed %v, want %v", log.Removed, tt.expected[i].Removed)
				}
				for j, topic := range log.Topics {
					if !bytes.Equal(topic, tt.expected[i].Topics[j]) {
						t.Errorf("got Topic %v, want %v", topic, tt.expected[i].Topics[j])
					}
				}
			}
		})
	}
}

// TestGetTokens tests the getTokens function, which converts a list of token
// addresses from a transaction into a list of common.Address type
func TestGetTokens(t *testing.T) {
	tests := []struct {
		name     string
		tokenStr []string
		expected []common.Address
	}{
		{
			name:     "single token address",
			tokenStr: []string{token.Hex()},
			expected: []common.Address{token},
		},
		{
			name:     "two token addresses",
			tokenStr: []string{token.Hex(), token2.Hex()},
			expected: []common.Address{token, token2},
		},
		{
			name:     "empty token addresses",
			tokenStr: []string{},
			expected: []common.Address{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTokens(tt.tokenStr)
			if len(result) != len(tt.expected) {
				t.Fatalf("got %v tokens, want %v tokens", len(result), len(tt.expected))
			}
			for i, token := range result {
				if token != tt.expected[i] {
					t.Errorf("got token %v, want %v", token, tt.expected[i])
				}
			}
		})
	}
}

// TestParseAddressBalance tests the parseAddressBalance function, which converts
// a list of token addresses, address and a list of balances into a list of
// Eth1AddressBalance type
func TestParseAddressBalance(t *testing.T) {
	tests := []struct {
		name   string
		tokens []common.
			Address
		address     string
		balances    []*big.Int
		expected    []*types.Eth1AddressBalance
		expectError bool
	}{
		{
			name:    "valid single token and balance",
			tokens:  []common.Address{token},
			address: aliceAddress.Hex(),
			balances: []*big.Int{
				big.NewInt(100),
			},
			expected: []*types.Eth1AddressBalance{
				{
					Address: aliceAddress.Bytes(),
					Token:   token.Bytes(),
					Balance: big.NewInt(100).Bytes(),
				},
			},
			expectError: false,
		},
		{
			name:    "valid two tokens and balances",
			tokens:  []common.Address{token, token2},
			address: aliceAddress.Hex(),
			balances: []*big.Int{
				big.NewInt(100),
				big.NewInt(200),
			},
			expected: []*types.Eth1AddressBalance{
				{
					Address: aliceAddress.Bytes(),
					Token:   token.Bytes(),
					Balance: big.NewInt(100).Bytes(),
				},
				{
					Address: aliceAddress.Bytes(),
					Token:   token2.Bytes(),
					Balance: big.NewInt(200).Bytes(),
				},
			},
			expectError: false,
		},
		{
			name:    "no tokens",
			tokens:  []common.Address{},
			address: aliceAddress.Hex(),
			balances: []*big.Int{
				big.NewInt(100),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:        "no balance",
			tokens:      []common.Address{token},
			address:     aliceAddress.Hex(),
			balances:    []*big.Int{},
			expected:    nil,
			expectError: true,
		},
		{
			name:    "single token and two balances",
			tokens:  []common.Address{token},
			address: aliceAddress.Hex(),
			balances: []*big.Int{
				big.NewInt(100),
				big.NewInt(200),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:    "invalid address format",
			tokens:  []common.Address{token},
			address: "invalid-address",
			balances: []*big.Int{
				big.NewInt(100),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:    "empty token address",
			tokens:  []common.Address{common.HexToAddress("")},
			address: aliceAddress.Hex(),
			balances: []*big.Int{
				big.NewInt(100),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name:    "nil balance",
			tokens:  []common.Address{token},
			address: aliceAddress.Hex(),
			balances: []*big.Int{
				nil,
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAddressBalance(tt.tokens, tt.address, tt.balances)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
				if result != nil {
					t.Errorf("expected result to be nil on error, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(result) != len(tt.expected) {
					t.Fatalf("got %v balances, want %v balances", len(result), len(tt.expected))
				}
				for i, res := range result {
					if !bytes.Equal(res.Address, tt.expected[i].Address) {
						t.Errorf("got Address %v, want %v", res.Address, tt.expected[i].Address)
					}
					if !bytes.Equal(res.Token, tt.expected[i].Token) {
						t.Errorf("got Token %v, want %v", res.Token, tt.expected[i].Token)
					}
					if !bytes.Equal(res.Balance, tt.expected[i].Balance) {
						t.Errorf("got Balance %v, want %v", res.Balance, tt.expected[i].Balance)
					}
				}
			}
		})
	}
}

// TestGetBlockUncles tests the getBlockUncles function which extracts a list of
// uncles from a block into a list of Eth1Block type
func TestGetBlockUncles(t *testing.T) {
	var (
		hash1 = common.HexToHash("0x1234567890abcdef1234567890abcdef12345678")
		hash2 = common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef12")
		hash3 = common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef13")
		hash4 = common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef14")
		hash5 = common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef15")
		hash6 = common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef16")
	)
	tests := []struct {
		name      string
		blkUncles []*gethtypes.Header
		expected  []*types.Eth1Block
	}{
		{
			name: "block with single uncle",
			blkUncles: []*gethtypes.Header{
				{
					Number:      big.NewInt(1),
					ParentHash:  hash1,
					UncleHash:   hash2,
					Coinbase:    common.HexToAddress("0xcoinbase"),
					Root:        hash3,
					TxHash:      hash4,
					ReceiptHash: hash5,
					Difficulty:  big.NewInt(100),
					GasLimit:    1000000,
					GasUsed:     500000,
					Time:        123456789,
					Extra:       []byte("extra"),
					MixDigest:   hash6,
					Bloom:       gethtypes.Bloom{},
				},
			},
			expected: []*types.Eth1Block{
				{
					Hash:        hash2.Bytes(),
					ParentHash:  hash1.Bytes(),
					UncleHash:   hash2.Bytes(),
					Coinbase:    common.HexToAddress("0xcoinbase").Bytes(),
					Root:        hash3.Bytes(),
					TxHash:      hash4.Bytes(),
					ReceiptHash: hash5.Bytes(),
					Difficulty:  big.NewInt(100).Bytes(),
					Number:      1,
					GasLimit:    1000000,
					GasUsed:     500000,
					Time:        timestamppb.New(time.Unix(123456789, 0)),
					Extra:       []byte("extra"),
					MixDigest:   hash6.Bytes(),
					Bloom:       gethtypes.Bloom{}.Bytes(),
				},
			},
		},
		{
			name: "block with two uncles",
			blkUncles: []*gethtypes.Header{
				{
					Number:      big.NewInt(1),
					ParentHash:  hash1,
					UncleHash:   hash2,
					Coinbase:    common.HexToAddress("0xcoinbase"),
					Root:        hash3,
					TxHash:      hash4,
					ReceiptHash: hash5,
					Difficulty:  big.NewInt(10),
					GasLimit:    1000000,
					GasUsed:     500000,
					Time:        123456789,
					Extra:       []byte("extra"),
					MixDigest:   hash6,
					Bloom:       gethtypes.Bloom{},
				},
				{
					Number:      big.NewInt(1),
					ParentHash:  hash2,
					UncleHash:   hash3,
					Coinbase:    common.HexToAddress("0xcoinbase"),
					Root:        hash4,
					TxHash:      hash5,
					ReceiptHash: hash6,
					Difficulty:  big.NewInt(100),
					GasLimit:    1000000,
					GasUsed:     500000,
					Time:        987654321,
					Extra:       []byte("extra"),
					MixDigest:   hash1,
					Bloom:       gethtypes.Bloom{},
				},
			},
			expected: []*types.Eth1Block{
				{
					Hash:        hash2.Bytes(),
					ParentHash:  hash1.Bytes(),
					UncleHash:   hash2.Bytes(),
					Coinbase:    common.HexToAddress("0xcoinbase").Bytes(),
					Root:        hash3.Bytes(),
					TxHash:      hash4.Bytes(),
					ReceiptHash: hash5.Bytes(),
					Difficulty:  big.NewInt(10).Bytes(),
					Number:      1,
					GasLimit:    1000000,
					GasUsed:     500000,
					Time:        timestamppb.New(time.Unix(123456789, 0)),
					Extra:       []byte("extra"),
					MixDigest:   hash6.Bytes(),
					Bloom:       gethtypes.Bloom{}.Bytes(),
				},
				{
					Hash:        hash3.Bytes(),
					ParentHash:  hash2.Bytes(),
					UncleHash:   hash3.Bytes(),
					Coinbase:    common.HexToAddress("0xcoinbase").Bytes(),
					Root:        hash4.Bytes(),
					TxHash:      hash5.Bytes(),
					ReceiptHash: hash6.Bytes(),
					Difficulty:  big.NewInt(100).Bytes(),
					Number:      1,
					GasLimit:    1000000,
					GasUsed:     500000,
					Time:        timestamppb.New(time.Unix(987654321, 0)),
					Extra:       []byte("extra"),
					MixDigest:   hash1.Bytes(),
					Bloom:       gethtypes.Bloom{}.Bytes(),
				},
			},
		},
		{
			name:      "block with no uncles",
			blkUncles: []*gethtypes.Header{},
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBlockUncles(tt.blkUncles)
			if len(result) != len(tt.expected) {
				t.Fatalf("got %v uncles, want %v uncles", len(result), len(tt.expected))
			}
			for i, uncle := range result {
				if !bytes.Equal(uncle.ParentHash, tt.expected[i].ParentHash) {
					t.Errorf("got ParentHash %v, want %v", uncle.ParentHash, tt.expected[i].ParentHash)
				}
				if !bytes.Equal(uncle.UncleHash, tt.expected[i].UncleHash) {
					t.Errorf("got Hash %v, want %v", uncle.UncleHash, tt.expected[i].UncleHash)
				}
				if !bytes.Equal(uncle.Coinbase, tt.expected[i].Coinbase) {
					t.Errorf("got Coinbase %v, want %v", uncle.Coinbase, tt.expected[i].Coinbase)
				}
				if !bytes.Equal(uncle.Root, tt.expected[i].Root) {
					t.Errorf("got Root %v, want %v", uncle.Root, tt.expected[i].Root)
				}
				if !bytes.Equal(uncle.TxHash, tt.expected[i].TxHash) {
					t.Errorf("got TxHash %v, want %v", uncle.TxHash, tt.expected[i].TxHash)
				}
				if !bytes.Equal(uncle.ReceiptHash, tt.expected[i].ReceiptHash) {
					t.Errorf("got ReceiptHash %v, want %v", uncle.ReceiptHash, tt.expected[i].ReceiptHash)
				}
				if !bytes.Equal(uncle.Difficulty, tt.expected[i].Difficulty) {
					t.Errorf("got Difficulty %v, want %v", uncle.Difficulty, tt.expected[i].Difficulty)
				}
				if uncle.Number != tt.expected[i].Number {
					t.Errorf("got Number %v, want %v", uncle.Number, tt.expected[i].Number)
				}
				if uncle.GasLimit != tt.expected[i].GasLimit {
					t.Errorf("got GasLimit %v, want %v", uncle.GasLimit, tt.expected[i].GasLimit)
				}
				if uncle.GasUsed != tt.expected[i].GasUsed {
					t.Errorf("got GasUsed %v, want %v", uncle.GasUsed, tt.expected[i].GasUsed)
				}
				if !uncle.Time.AsTime().Equal(tt.expected[i].Time.AsTime()) {
					t.Errorf("got Time %v, want %v", uncle.Time.AsTime(), tt.expected[i].Time.AsTime())
				}
				if !bytes.Equal(uncle.Extra, tt.expected[i].Extra) {
					t.Errorf("got Extra %v, want %v", uncle.Extra, tt.expected[i].Extra)
				}
				if !bytes.Equal(uncle.MixDigest, tt.expected[i].MixDigest) {
					t.Errorf("got MixDigest %v, want %v", uncle.MixDigest, tt.expected[i].MixDigest)
				}
				if !bytes.Equal(uncle.Bloom, tt.expected[i].Bloom) {
					t.Errorf("got Bloom %v, want %v", uncle.Bloom, tt.expected[i].Bloom)
				}
			}
		})
	}
}

// TestGetERC20ContractSymbol tests the getERC20ContractSymbol function which extracts
// the symbol of an ERC20 contract
func TestGetERC20ContractSymbol(t *testing.T) {
	backend := th.NewBackend(t)
	ret := &types.ERC20Metadata{}

	contractAddress, _ := backend.DeployERC20(t, "usdt", "USDT", backend.BankAccount.From)
	contractMetadata, err := backend.NewERC20Metadata(t, contractAddress)
	if err != nil {
		t.Fatalf("could not create contract: %v", err)
	}

	test := []struct {
		name        string
		err         error
		expected    string
		expectedErr error
	}{
		{
			name:        "valid symbol",
			err:         nil,
			expected:    "USDT",
			expectedErr: nil,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			err := getERC20ContractSymbol(contractMetadata, ret)
			if err != nil && tt.expectedErr == nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && tt.expectedErr != nil {
				t.Fatalf("expected error: %v, got nil", tt.expectedErr)
			}
			if ret.Symbol != tt.expected {
				t.Errorf("got Symbol %v, want %v", ret.Symbol, tt.expected)
			}
		})
	}

	contractAddress2, _ := backend.DeployERC20(t, "usdt", "", backend.BankAccount.From)
	contractMetadata2, err := backend.NewERC20Metadata(t, contractAddress2)
	if err != nil {
		t.Fatalf("could not create contract: %v", err)
	}

	test2 := []struct {
		name        string
		err         error
		expected    string
		expectedErr error
	}{
		{
			name:        "empty symbol",
			err:         nil,
			expected:    "",
			expectedErr: nil,
		},
	}

	for _, tt := range test2 {
		t.Run(tt.name, func(t *testing.T) {
			err := getERC20ContractSymbol(contractMetadata2, ret)
			if err != nil && tt.expectedErr == nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && tt.expectedErr != nil {
				t.Fatalf("expected error: %v, got nil", tt.expectedErr)
			}
			if ret.Symbol != tt.expected {
				t.Errorf("got Symbol %v, want %v", ret.Symbol, tt.expected)
			}
		})
	}
}

// TestGetERC20ContractTotalSupply tests the getERC20ContractTotalSupply function which
// extracts the total supply of an ERC20 contract
func TestGetERC20ContractTotalSupply(t *testing.T) {
	backend := th.NewBackend(t)
	ret := &types.ERC20Metadata{}

	contractAddress, _ := backend.DeployERC20(t, "usdt", "USDT", backend.BankAccount.From)
	contractMetadata, err := backend.NewERC20Metadata(t, contractAddress)
	if err != nil {
		t.Fatalf("could not create contract: %v", err)
	}

	tests := []struct {
		name        string
		err         error
		expected    []byte
		expectedErr error
	}{
		{
			name:        "valid total supply",
			err:         nil,
			expected:    big.NewInt(100).Bytes(),
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := getERC20ContractTotalSupply(contractMetadata, ret)
			if err != nil && tt.expectedErr == nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && tt.expectedErr != nil {
				t.Fatalf("expected error: %v, got nil", tt.expectedErr)
			}
			if !bytes.Equal(ret.TotalSupply, tt.expected) {
				t.Errorf("got TotalSupply %v, want %v", ret.TotalSupply, tt.expected)
			}
		})
	}
}

// TestGetERC20ContractDecimals tests the getERC20ContractDecimals function which
// extracts the decimals of an ERC20 contract
func TestGetERC20ContractDecimals(t *testing.T) {
	backend := th.NewBackend(t)
	ret := &types.ERC20Metadata{}

	contractAddress, _ := backend.DeployERC20(t, "usdt", "USDT", backend.BankAccount.From)
	contractMetadata, err := backend.NewERC20Metadata(t, contractAddress)
	if err != nil {
		t.Fatalf("could not create contract: %v", err)
	}

	tests := []struct {
		name        string
		err         error
		expected    []byte
		expectedErr error
	}{
		{
			name:        "valid decimals",
			err:         nil,
			expected:    big.NewInt(18).Bytes(),
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := getERC20ContractDecimals(contractMetadata, ret)
			if err != nil && tt.expectedErr == nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && tt.expectedErr != nil {
				t.Fatalf("expected error: %v, got nil", tt.expectedErr)
			}
			if !bytes.Equal(ret.Decimals, tt.expected) {
				t.Errorf("got Decimals %v, want %v", ret.Decimals, tt.expected)
			}
		})
	}
}
