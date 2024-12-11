package raw

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/gobitfly/beaconchain/pkg/commons/chain"
)

func TestSigner(t *testing.T) {
	tests := []struct {
		name    string
		chainID *big.Int
		tx      *types.Transaction
		sender  common.Address
		wantErr bool
	}{
		{
			name:    "default",
			chainID: chain.IDs.Mainnet,
			sender:  common.HexToAddress("0x1234000000000000000000000000000000000000"),
		},
		{
			name:    "default ok with chain ID = 0 (pre EIP-155)",
			chainID: big.NewInt(0),
			sender:  common.HexToAddress("0x1234000000000000000000000000000000000000"),
		},
		{
			name:    "default error with address 0",
			chainID: chain.IDs.Mainnet,
			sender:  common.Address{},
			wantErr: true,
		},
		{
			name:    "op ok with address 0",
			chainID: chain.IDs.Optimistic,
			sender:  common.Address{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := types.NewTx(&types.LegacyTx{})
			setSender(tt.chainID, tx, tt.sender, common.Hash{1})
			sender, err := types.Sender(SignerForChainID(tt.chainID, common.Hash{1}), tx)
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("failed to get tx sender: %v", err)
				}
				return
			}
			if got, want := sender, tt.sender; got.Cmp(want) != 0 {
				t.Errorf("got sender %v, want %v", got, want)
			}
		})
	}
}

func mustGetEnv(key string) func(t *testing.T) string {
	return func(t *testing.T) string {
		val := os.Getenv(key)
		if val == "" {
			t.Skipf("skipping test, set %s", key)
		}
		return val
	}
}

func TestSignerRealCondition(t *testing.T) {
	tests := []struct {
		name      string
		chainID   *big.Int
		txHash    common.Hash
		blockHash common.Hash
		txIndex   uint
		getUrl    func(t *testing.T) string
	}{
		{
			name:      "op without signer",
			chainID:   chain.IDs.Optimistic,
			txHash:    common.HexToHash("0xfdb23a026aebea63fb0e56d82bc27d61cc1a4e49c34f88928b371a38742d3bb9"),
			blockHash: common.HexToHash("0x239ca8170b66e63659800f92079a30838d52dc90cf71b901cedf394226e87856"),
			txIndex:   0,
			getUrl:    mustGetEnv("NODE_URL_OP"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ethclient.Dial(tt.getUrl(t))
			if err != nil {
				t.Fatal(err)
			}
			tx, _, err := c.TransactionByHash(context.Background(), tt.txHash)
			if err != nil {
				t.Fatal(err)
			}
			sender, err := c.TransactionSender(context.Background(), tx, tt.blockHash, tt.txIndex)
			if err != nil {
				t.Fatal(err)
			}
			setSender(tt.chainID, tx, sender, tt.blockHash)
			got, err := types.Sender(SignerForChainID(tt.chainID, tt.blockHash), tx)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := got, sender; got.Cmp(want) != 0 {
				t.Errorf("got sender %v, want %v", got, want)
			}
		})
	}
}
