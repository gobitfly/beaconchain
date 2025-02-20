package modules

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/coocood/freecache"
	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/monitoring/constants"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
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

func TestGetPreviousHeadEpoch(t *testing.T) {
	tests := []struct {
		name              string
		mockRows          *sqlmock.Rows
		mockError         error
		expectedHeadEpoch uint64
		expectedError     bool
	}{
		{
			name:              "valid previous head epoch",
			mockRows:          sqlmock.NewRows([]string{"headepoch"}).AddRow(2),
			mockError:         nil,
			expectedHeadEpoch: 2,
			expectedError:     false,
		},
		{
			name:          "GetNetworkLivenessPreviousHeadEpoch error",
			mockRows:      nil,
			mockError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer dbMock.Close()

			sqlxDB := sqlx.NewDb(dbMock, "sqlmock")
			db.WriterDb = sqlxDB

			// mock GetNetworkLivenessPreviousHeadEpoch query
			query := `SELECT COALESCE\(MAX\(headepoch\), 0\) FROM network_liveness`
			if tt.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(query).WillReturnRows(tt.mockRows)
			}

			prevHeadEpoch, err := getPreviousHeadEpoch()

			if !tt.expectedError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if prevHeadEpoch != tt.expectedHeadEpoch {
					t.Errorf("expected head epoch: %v, got: %v", tt.expectedHeadEpoch, prevHeadEpoch)
				}
			}
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error: %v, got nil", tt.expectedError)
				}
			}
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

			err := cache.LatestNodeEpoch.Set(tt.latestEpoch)
			if err != nil {
				t.Errorf("unexpected latest epoch error: %v", err)
			}

			err = cache.LatestNodeFinalizedEpoch.Set(tt.latestFinalizedEpoch)
			if err != nil {
				t.Errorf("unexpected latest finalized epoch error: %v", err)
			}

			err = updateCache(tt.head)
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
