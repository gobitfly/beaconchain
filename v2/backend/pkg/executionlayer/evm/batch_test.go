package evm

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/gobitfly/beaconchain/internal/contracts"
	"github.com/gobitfly/beaconchain/internal/th"
	"github.com/gobitfly/beaconchain/pkg/commons/chain"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/metadataupdates"
)

func TestNewBatcher(t *testing.T) {
	tests := []struct {
		name             string
		chainID          *big.Int
		multicallAddress *common.Address
		expected         Batcher
	}{
		{
			name:     "select multicall if known",
			chainID:  chain.IDs.Mainnet,
			expected: MulticallBatcher{},
		},
		{
			name:     "select batch if unknown",
			chainID:  big.NewInt(-1),
			expected: RPCBatcher{},
		},
		{
			name:             "select multicall when address provided",
			multicallAddress: pointer(common.HexToAddress("0x1")),
			chainID:          big.NewInt(-1),
			expected:         MulticallBatcher{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poller := NewBatcher(tt.chainID, &ethclient.Client{}, BatcherConfig{
				MulticallAddress: tt.multicallAddress,
			})
			if got, want := reflect.TypeOf(poller), reflect.TypeOf(tt.expected); got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}

func TestBatcher(t *testing.T) {
	b := th.NewBackend(t)
	tokenAddress, token := b.DeployERC20(t, "usdt", "usdt", b.BankAccount.From)
	multicall := b.DeployContract(t, common.FromHex(contracts.MulticallMetaData.Bin))
	b.Commit()

	tests := []struct {
		name    string
		batcher Batcher
	}{
		{
			name:    "rpc batch",
			batcher: NewRPCBatch(b.Client().Client(), 0),
		},
		{
			name:    "multicall",
			batcher: NewMulticallBatcher(b.Client(), multicall, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("Balance", func(t *testing.T) {
				t.Run("erc20", func(t *testing.T) {
					balance, err := token.BalanceOf(nil, b.BankAccount.From)
					if err != nil {
						t.Fatal(err)
					}
					if balance.Cmp(big.NewInt(0)) == 0 {
						t.Fatal("expected balance cannot be zero")
					}

					res, err := BalanceForPairs(tt.batcher, []metadataupdates.Pair{
						{Address: b.BankAccount.From, Token: tokenAddress},
					})
					if err != nil {
						t.Fatal(err)
					}
					if got, want := res[0].String(), balance.String(); got != want {
						t.Errorf("got %v, want %v", got, want)
					}
				})
				t.Run("eth", func(t *testing.T) {
					balance, err := b.Client().BalanceAt(context.Background(), b.BankAccount.From, nil)
					if err != nil {
						t.Fatal(err)
					}
					if balance.Cmp(big.NewInt(0)) == 0 {
						t.Fatal("expected balance cannot be zero")
					}

					res, err := BalanceForPairs(tt.batcher, []metadataupdates.Pair{
						{Address: b.BankAccount.From},
					})
					if err != nil {
						t.Fatal(err)
					}
					if got, want := res[0].String(), balance.String(); got != want {
						t.Errorf("got %v, want %v", got, want)
					}
				})
			})
		})
	}
}

func TestBatchWithLimit(t *testing.T) {
	callCount := 0
	limit := 2

	res, err := batchWithLimit([]BatchElement{
		{Method: "eth_call", To: &common.Address{}},
		{Method: "eth_call", To: &common.Address{}},
		{Method: "eth_call", To: &common.Address{}},
	}, limit, func(elements []BatchElement) ([]BatchResponse, error) {
		callCount++
		res := make([]BatchResponse, len(elements))
		return res, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(res), 3; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := callCount, 2; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func pointer[T any](t T) *T {
	return &t
}
