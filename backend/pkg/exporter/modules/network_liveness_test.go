package modules

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	cachemocks "github.com/gobitfly/beaconchain/pkg/commons/cache/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/db/mocks"
	rpcmocks "github.com/gobitfly/beaconchain/pkg/commons/rpc/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

func TestNetworkLivenessExport(t *testing.T) {
	tests := []struct {
		name                  string
		mockChainHeadResponse *types.ChainHead
		prevHeadEpoch         uint64
		depositChainID        uint64
		sameHeadEpoch         bool
		nodeIsSynced          bool
		mockChainHeadError    error
	}{
		{
			name: "different head epoch and node is synced",
			mockChainHeadResponse: &types.ChainHead{
				HeadEpoch:      12345678,
				FinalizedEpoch: 12345678,
			},
			prevHeadEpoch:  12345677,
			depositChainID: 1,
			nodeIsSynced:   true,
		},
		{
			name: "same head epoch",
			mockChainHeadResponse: &types.ChainHead{
				HeadEpoch:      12345678,
				FinalizedEpoch: 12345678,
			},
			prevHeadEpoch: 12345678,
			sameHeadEpoch: true,
		},
		{
			name: "node is not synced",
			mockChainHeadResponse: &types.ChainHead{
				HeadEpoch:      1,
				FinalizedEpoch: 1,
			},
		},
	}

	utils.Config = &types.Config{
		Chain: types.Chain{
			GenesisTimestamp: 1581957675,
			ClConfig: types.ClChainConfig{
				SecondsPerSlot: 1,
				SlotsPerEpoch:  32,
				DepositChainID: 1,
			},
		},
		DeploymentType: "development",
	}

	mockConsDBClient := new(mocks.ConsensusDBI)
	mockRPCClient := new(rpcmocks.EpochClient)
	cachemocks := new(cachemocks.RemoteCache)
	tieredCache := cache.NewTieredCache(cachemocks, 1000)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	updater := networkLivenessUpdater{
		client: mockRPCClient,
		db:     mockConsDBClient,
		ctx:    ctx,
		cache:  tieredCache,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			latestNodeEpochKey := fmt.Sprintf("%d:frontend:latestNodeFinalizedEpoch", tt.depositChainID)
			latestNodeFinalizedEpochKey := fmt.Sprintf("%d:frontend:latestFinalized", tt.depositChainID)

			mockConsDBClient.On("GetNetworkLivenessPreviousHeadEpoch").Return(tt.prevHeadEpoch, nil)
			mockRPCClient.On("GetChainHead").Return(tt.mockChainHeadResponse, tt.mockChainHeadError)
			if !tt.sameHeadEpoch && tt.nodeIsSynced {
				mockConsDBClient.On("SaveNetworkLivenessData", tt.mockChainHeadResponse).Return(nil)
				cachemocks.On("SetUint64", "mock.Anything", latestNodeEpochKey, tt.mockChainHeadResponse.HeadEpoch, time.Hour*24).Return(nil)
				cachemocks.On("SetUint64", "mock.Anything", latestNodeFinalizedEpochKey, tt.mockChainHeadResponse.FinalizedEpoch, time.Hour*24).Return(nil)
			}

			updater.Export()

			mockConsDBClient.AssertCalled(t, "GetNetworkLivenessPreviousHeadEpoch")
			mockRPCClient.AssertCalled(t, "GetChainHead")
			if !tt.sameHeadEpoch && tt.nodeIsSynced {
				mockConsDBClient.AssertCalled(t, "SaveNetworkLivenessData", tt.mockChainHeadResponse)
				cachemocks.AssertCalled(t, "SetUint64", "mock.Anything", latestNodeEpochKey, tt.mockChainHeadResponse.HeadEpoch, time.Hour*24)
				cachemocks.AssertCalled(t, "SetUint64", "mock.Anything", latestNodeFinalizedEpochKey, tt.mockChainHeadResponse.FinalizedEpoch, time.Hour*24)
			}
		})
	}
}
