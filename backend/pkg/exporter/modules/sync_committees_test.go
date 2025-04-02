package modules

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/cache"
	cachemocks "github.com/gobitfly/beaconchain/pkg/commons/cache/mocks"
	dbmocks "github.com/gobitfly/beaconchain/pkg/commons/db2/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
)

func TestSyncCommitteesExport(t *testing.T) {
	tests := []struct {
		name                      string
		mockDBPeriods             []uint64
		mockStateID               string
		mockDepositChainID        uint64
		mockEpoch                 uint64
		mockSyncCommitteeResponse *constypes.StandardSyncCommittee
		mockSyncCommitteesData    []types.SyncCommittee
		mockSyncCommitteesData2   []types.SyncCommittee
		mockNoRecordsInDB         bool
	}{
		{
			name:               "no records in db",
			mockDBPeriods:      nil,
			mockStateID:        "10",
			mockEpoch:          1,
			mockDepositChainID: 1,
			mockSyncCommitteeResponse: &constypes.StandardSyncCommittee{
				Validators: []constypes.Uint64Str{
					1,
				},
			},
			mockSyncCommitteesData: []types.SyncCommittee{
				{
					Period:         0,
					ValidatorIndex: 1,
					CommitteeIndex: 0,
				},
			},
			mockSyncCommitteesData2: []types.SyncCommittee{
				{
					Period:         1,
					ValidatorIndex: 1,
					CommitteeIndex: 0,
				},
			},
			mockNoRecordsInDB: true,
		},
		{
			name:               "records exist in db",
			mockDBPeriods:      []uint64{1},
			mockStateID:        "10",
			mockEpoch:          1,
			mockDepositChainID: 1,
			mockSyncCommitteeResponse: &constypes.StandardSyncCommittee{
				Validators: []constypes.Uint64Str{
					1,
				},
			},
			mockSyncCommitteesData: []types.SyncCommittee{
				{
					Period:         0,
					ValidatorIndex: 1,
					CommitteeIndex: 0,
				},
			},
		},
	}

	mockConsDBClient := new(dbmocks.ConsensusRepository)
	mockClient := new(mocks.SyncCommitteeClient)
	cachemocks := new(cachemocks.RemoteCache)
	tieredCache := cache.NewTieredCache(cachemocks, 1000)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	exporter := syncCommitteesExporter{
		client: mockClient,
		db:     mockConsDBClient,
		delay:  0,
		ctx:    ctx,
		cache:  tieredCache,
	}

	utils.Config = &types.Config{
		Chain: types.Chain{
			ClConfig: types.ClChainConfig{
				AltairForkEpoch:              1,
				EpochsPerSyncCommitteePeriod: 32,
				SlotsPerEpoch:                10,
				DepositChainID:               1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			latestNodeFinalizedEpochKey := fmt.Sprintf("%d:frontend:latestFinalized", tt.mockDepositChainID)

			mockConsDBClient.On("GetSyncCommitteesPeriods").Return(tt.mockDBPeriods, nil)
			cachemocks.On("GetUint64", "mock.Anything", latestNodeFinalizedEpochKey).Return(tt.mockEpoch, nil)
			mockClient.On("GetSyncCommittee", tt.mockStateID, tt.mockEpoch).Return(tt.mockSyncCommitteeResponse, nil)
			mockConsDBClient.On("SaveSyncCommitteeData", tt.mockSyncCommitteesData).Return(nil)
			mockConsDBClient.On("SaveSyncCommitteeData", tt.mockSyncCommitteesData2).Return(nil)

			exporter.Export()

			mockConsDBClient.AssertCalled(t, "GetSyncCommitteesPeriods")
			mockClient.AssertCalled(t, "GetSyncCommittee", tt.mockStateID, tt.mockEpoch)
			mockConsDBClient.AssertCalled(t, "SaveSyncCommitteeData", tt.mockSyncCommitteesData)
			if tt.mockNoRecordsInDB {
				mockConsDBClient.AssertCalled(t, "SaveSyncCommitteeData", tt.mockSyncCommitteesData2)
			}
			mockClient.AssertCalled(t, "GetSyncCommittee", tt.mockStateID, tt.mockEpoch)
			cachemocks.AssertCalled(t, "GetUint64", "mock.Anything", latestNodeFinalizedEpochKey)
		})
	}
}
