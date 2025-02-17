package modules

import (
	"context"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
)

func TestCreateNetworkStatusReport(t *testing.T) {
	tests := []struct {
		name         string
		slotDuration time.Duration
	}{
		{
			name:         "valid network status report",
			slotDuration: time.Second * 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportFunc := createNetworkStatusReport(tt.slotDuration)
			if reportFunc == nil {
				t.Error("expected a non-nil function")
			}
		})
	}
}

func TestHandleNetworkSuccess(t *testing.T) {
	tests := []struct {
		name         string
		slotDuration time.Duration
	}{
		{
			name:         "valid network success",
			slotDuration: time.Second * 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusReport := func(status constants.StatusType, metadata map[string]string) {}
			handleNetworkSuccess(tt.slotDuration, statusReport)
		})
	}
}

func TestIsNodeSynced(t *testing.T) {
	tests := []struct {
		name           string
		headEpoch      uint64
		epochDuration  time.Duration
		expectedResult bool
	}{
		{
			name:           "Test node is synced",
			headEpoch:      100,
			epochDuration:  time.Second * 12,
			expectedResult: true,
		},
		{
			name:           "Test node is not synced",
			headEpoch:      1000000,
			epochDuration:  time.Second * 12,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.Config = &types.Config{
				Chain: types.Chain{
					GenesisTimestamp: 1581957675,
					ClConfig: types.ClChainConfig{
						SecondsPerSlot: 12,
						SlotsPerEpoch:  32,
					},
				},
			}

			isSynced := isNodeSynced(tt.headEpoch, tt.epochDuration)
			if isSynced != tt.expectedResult {
				t.Errorf("isNodeSynced() = %v, want %v", isSynced, tt.expectedResult)
			}
		})
	}
}

func TestUpdateCache(t *testing.T) {
	tests := []struct {
		name                 string
		head                 *types.ChainHead
		latestEpoch          uint64
		latestFinalizedEpoch uint64
		expectedError        bool
	}{
		{
			name: "successful cache update",
			head: &types.ChainHead{
				HeadEpoch:      10,
				FinalizedEpoch: 10,
			},
			latestEpoch:          10,
			latestFinalizedEpoch: 10,
			expectedError:        false,
		},
		{
			name:          "cache update error",
			head:          &types.ChainHead{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			utils.Config = &types.Config{
				Chain: types.Chain{
					ClConfig: types.ClChainConfig{
						AltairForkEpoch:              8,
						EpochsPerSyncCommitteePeriod: 1,
						SlotsPerEpoch:                32,
						DepositChainID:               1,
					},
				},
			}

			cache.TieredCache = &cache.TieredCacheBase{
				LocalGoCache: freecache.NewCache(100 * 1024 * 1024), // 100 MB
				RemoteCache:  &MockRemoteCache{},
			}

			cache.LatestNodeEpoch.Set(tt.latestEpoch)
			cache.LatestNodeFinalizedEpoch.Set(tt.latestFinalizedEpoch)

			err := updateCache(tt.head)
			if tt.expectedError {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}

		})
	}
}

type MockRemoteCache struct{}

func (m *MockRemoteCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return nil
}

func (m *MockRemoteCache) GetUint64(ctx context.Context, key string) (uint64, error) {
	return 0, nil
}

func (m *MockRemoteCache) SetUint64(ctx context.Context, key string, value uint64, expiration time.Duration) error {
	return nil
}

func (m *MockRemoteCache) Get(ctx context.Context, key string, dest any) (any, error) {
	return 0, nil
}

func (m *MockRemoteCache) GetBool(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (m *MockRemoteCache) GetString(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (m *MockRemoteCache) SetBool(ctx context.Context, key string, value bool, expiration time.Duration) error {
	return nil
}

func (m *MockRemoteCache) SetString(ctx context.Context, key string, value string, expiration time.Duration) error {
	return nil
}
