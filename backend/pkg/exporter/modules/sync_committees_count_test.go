package modules

import (
	"context"
	"testing"
	"time"

	dbmocks "github.com/gobitfly/beaconchain/pkg/commons/db2/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

func TestSyncCommitteesCountExport(t *testing.T) {
	utils.Config = &types.Config{
		Chain: types.Chain{
			ClConfig: types.ClChainConfig{
				AltairForkEpoch:              1,
				EpochsPerSyncCommitteePeriod: 32,
			},
		},
	}

	t.Run("records exist in db", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		mockConsDBClient := new(dbmocks.ConsensusRepository)
		exporter := syncCommitteesCountExporter{
			db:    mockConsDBClient,
			delay: 0,
			ctx:   ctx,
		}

		mockConsDBClient.On("GetSyncCommitteesCountPerValidator").Return(uint64(1), nil)
		mockConsDBClient.On("GetLatestFinalizedEpoch").Return(uint64(1), nil)
		mockConsDBClient.On("GetTotalPeriodSyncCommitteesCountPerValidator").Return(uint64(1), nil)
		mockConsDBClient.On("GetCountSoFarSyncCommitteesCountPerValidator", uint64(1)).Return(float64(1.0), nil)

		exporter.Export()

		mockConsDBClient.AssertCalled(t, "GetSyncCommitteesCountPerValidator")
		mockConsDBClient.AssertCalled(t, "GetLatestFinalizedEpoch")
		mockConsDBClient.AssertCalled(t, "GetTotalPeriodSyncCommitteesCountPerValidator")
		mockConsDBClient.AssertCalled(t, "GetCountSoFarSyncCommitteesCountPerValidator", uint64(1))
	})

	t.Run("no records in db", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		mockConsDBClient := new(dbmocks.ConsensusRepository)
		exporter := syncCommitteesCountExporter{
			db:    mockConsDBClient,
			delay: 0,
			ctx:   ctx,
		}

		mockConsDBClient.On("GetSyncCommitteesCountPerValidator").Return(uint64(0), nil)
		mockConsDBClient.On("GetLatestFinalizedEpoch").Return(uint64(0), nil)
		mockConsDBClient.On("SaveSyncCommitteesCount", uint64(0), float64(0.0)).Return(nil)

		exporter.Export()

		mockConsDBClient.AssertCalled(t, "GetSyncCommitteesCountPerValidator")
		mockConsDBClient.AssertCalled(t, "GetLatestFinalizedEpoch")
		mockConsDBClient.AssertCalled(t, "SaveSyncCommitteesCount", uint64(0), float64(0.0))
	})
}
