package evm_node_indexer

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	gethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

func TestUpdateBlockNumberWithRewindingNode(t *testing.T) {
	utils.Config = &types.Config{
		Chain: types.Chain{
			ClConfig: types.ClChainConfig{
				SecondsPerSlot: 0,
			},
		},
	}
	elClient = &fakeEthClient{
		chainID: big.NewInt(1),
		blocks:  []uint64{126192682, 126186030},
	}
	err := updateBlockNumber(false, time.Second, nil, nil, nil)
	if err == nil {
		t.Fatal("expected error got nil")
	}
	if !strings.Contains(err.Error(), "node is rewinding") {
		t.Fatalf("error message does not contain 'node is rewinding', got %s", err.Error())
	}
}

type fakeEthClient struct {
	chainID *big.Int
	blocks  []uint64
}

func (f *fakeEthClient) SubscribeNewHead(ctx context.Context, ch chan<- *gethtypes.Header) (ethereum.Subscription, error) {
	panic("implement me")
}

func (f *fakeEthClient) ChainID(ctx context.Context) (*big.Int, error) {
	return f.chainID, nil
}

func (f *fakeEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	block := f.blocks[0]
	f.blocks = f.blocks[1:]
	return block, nil
}
