package modules

import (
	"context"
	"testing"
	"time"

	dbmocks "github.com/gobitfly/beaconchain/pkg/commons/db/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

func TestSyncCommitteesCountExport(t *testing.T) {
	tests := []struct {
		name                     string
		mockRowCountResponse     uint64
		mockLatestFinalizedEpoch uint64
		mockDBPeriod             uint64
		mockCountSoFar           float64
		mockTotalValidatorsCount uint64
		mockNoRecordsInDB        bool
	}{
		{
			name:                     "no records in db",
			mockRowCountResponse:     0,
			mockLatestFinalizedEpoch: 0,
			mockNoRecordsInDB:        true,
		},
		{
			name:                     "records exist in db",
			mockRowCountResponse:     1,
			mockLatestFinalizedEpoch: 1,
			mockDBPeriod:             1,
			mockCountSoFar:           1.0,
			mockTotalValidatorsCount: 1,
		},
	}

	utils.Config = &types.Config{
		Chain: types.Chain{
			ClConfig: types.ClChainConfig{
				AltairForkEpoch:              1,
				EpochsPerSyncCommitteePeriod: 32,
			},
		},
		DeploymentType: "development",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConsDBClient := new(dbmocks.ConsensusDBI)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			exporter := syncCommitteesCountExporter{
				db:    mockConsDBClient,
				delay: 0,
				ctx:   ctx,
			}

			if tt.mockNoRecordsInDB {
				mockConsDBClient.On("GetSyncCommitteesCountPerValidator").Return(tt.mockRowCountResponse, nil)
				mockConsDBClient.On("GetLatestFinalizedEpoch").Return(tt.mockLatestFinalizedEpoch, nil)
				mockConsDBClient.On("SaveSyncCommitteesCount", tt.mockDBPeriod, tt.mockCountSoFar).Return(nil)

				exporter.Export()

				mockConsDBClient.AssertCalled(t, "GetSyncCommitteesCountPerValidator")
				mockConsDBClient.AssertCalled(t, "GetLatestFinalizedEpoch")
				mockConsDBClient.AssertCalled(t, "SaveSyncCommitteesCount", tt.mockDBPeriod, tt.mockCountSoFar)

				return
			}

			mockConsDBClient.On("GetSyncCommitteesCountPerValidator").Return(tt.mockRowCountResponse, nil)
			mockConsDBClient.On("GetLatestFinalizedEpoch").Return(tt.mockLatestFinalizedEpoch, nil)
			mockConsDBClient.On("GetTotalPeriodSyncCommitteesCountPerValidator").Return(tt.mockDBPeriod, nil)
			mockConsDBClient.On("GetCountSoFarSyncCommitteesCountPerValidator", tt.mockDBPeriod).Return(tt.mockCountSoFar, nil)

			exporter.Export()

			mockConsDBClient.AssertCalled(t, "GetSyncCommitteesCountPerValidator")
			mockConsDBClient.AssertCalled(t, "GetLatestFinalizedEpoch")
			mockConsDBClient.AssertCalled(t, "GetTotalPeriodSyncCommitteesCountPerValidator")
			mockConsDBClient.AssertCalled(t, "GetCountSoFarSyncCommitteesCountPerValidator", tt.mockDBPeriod)
		})
	}
}
