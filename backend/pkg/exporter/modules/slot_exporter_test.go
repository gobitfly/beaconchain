package modules

import (
	"fmt"
	"sync"
	"testing"
	"time"

	rpcmocks "github.com/gobitfly/beaconchain/pkg/commons/rpc/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	dbmocks "github.com/gobitfly/beaconchain/pkg/exporter/db/mocks"
	"github.com/jmoiron/sqlx"
)

func TestSlotExporterOnHead(t *testing.T) {
	t.Run("genesis slot", func(t *testing.T) {
		mockChainHead := &types.ChainHead{
			HeadSlot:       1,
			HeadEpoch:      1,
			FinalizedSlot:  1,
			FinalizedEpoch: 1,
		}
		mockDBSlots := []uint64{}
		mockDBLastSlot := uint64(0)
		mockNonFinalSlots := []*edb.NonFinalizedSlotsRow{}
		mockBlock := &types.Block{
			Status: 1,
			EpochAssignments: &types.EpochAssignments{
				ProposerAssignments: map[uint64]uint64{
					0: 1,
				},
				AttestorAssignments: map[string]uint64{
					"0": 1,
				},
				SyncAssignments: []uint64{
					1,
				},
			},
			Slot: 1,
			Validators: []*types.Validator{
				{
					Index:                      0,
					PublicKey:                  []byte{1, 2, 3},
					WithdrawalCredentials:      []byte{5, 6, 7},
					EffectiveBalance:           32,
					Status:                     "1",
					Slashed:                    false,
					ActivationEligibilityEpoch: 2,
					ActivationEpoch:            2,
					ExitEpoch:                  2,
					WithdrawableEpoch:          2,
				},
			},
			AttestationDuties: map[types.ValidatorIndex][]types.Slot{
				0: {1},
			},
			SyncDuties: map[types.ValidatorIndex]bool{
				0: true,
			},
		}

		utils.Config = &types.Config{
			Chain: types.Chain{
				ClConfig: types.ClChainConfig{
					DepositChainID: 1,
					SlotsPerEpoch:  32,
				},
			},
		}

		mockDB := new(dbmocks.SlotExporterDBRepository)
		mockBT := new(dbmocks.SlotExporterBTRepository)
		mockClient := new(rpcmocks.SlotExporterClient)
		mockCache := new(dbmocks.SlotExporterCacheRepository)
		mockTx := new(sqlx.Tx)

		// mock DB calls
		mockDB.On("BeginTx").Return(mockTx, nil)
		mockDB.On("RollbackTx", mockTx).Return(nil)
		mockDB.On("CommitTx", mockTx).Return(nil)
		mockDB.On("GetAllSlots", mockTx).Return(mockDBSlots, nil)
		mockDB.On("GetLastSlot", mockTx).Return(mockDBLastSlot, nil)
		mockDB.On("SaveBlock", mockBlock, false, mockTx).Return(nil)
		mockDB.On("SaveEpoch", utils.EpochOfSlot(mockChainHead.HeadEpoch), mockBlock.Validators, mockTx).Return(nil)
		mockDB.On("GetAllNonFinalizedSlots").Return(mockNonFinalSlots, nil)

		// mock BT calls
		mockBT.On("SaveAttestationDuties", "mock.Anything").Return(nil)
		mockBT.On("SaveSyncCommitteeDuties", "mock.Anything").Return(nil)
		mockBT.On("SaveValidatorBalances", utils.EpochOfSlot(mockChainHead.HeadEpoch), mockBlock.Validators).Return(nil)

		// mock Client calls
		mockClient.On("GetChainHead").Return(mockChainHead, nil)
		mockClient.On("GetBlockBySlot", mockChainHead.HeadSlot).Return(mockBlock, nil)

		// mock Cache calls
		mockCache.On("GetLatestEpoch").Return(mockChainHead.HeadEpoch-1, nil)
		mockCache.On("GetLatestSlot").Return(mockChainHead.HeadSlot-1, nil)
		mockCache.On("GetLatestFinalizedEpoch").Return(mockChainHead.FinalizedEpoch-1, nil)
		mockCache.On("GetLatestProposedSlot").Return(mockBlock.Slot-1, nil)
		mockCache.On("SetLatestSlot", mockChainHead.HeadSlot).Return(nil)
		mockCache.On("SetLatestProposedSlot", mockChainHead.HeadSlot).Return(nil)

		exporter := &slotExporter{
			Client:   mockClient,
			cache:    mockCache,
			db:       mockDB,
			bt:       mockBT,
			firstRun: true,
		}

		err := exporter.OnHead(nil)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		mockDB.AssertExpectations(t)
		mockBT.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("head epoch", func(t *testing.T) {
		mockChainHead := &types.ChainHead{
			HeadSlot:       111,
			HeadEpoch:      3,
			FinalizedSlot:  111,
			FinalizedEpoch: 3,
		}
		mockDBLastSlot := uint64(110)
		mockNonFinalSlots := []*edb.NonFinalizedSlotsRow{
			{
				Slot:      111,
				BlockRoot: []byte{1, 2, 3, 4, 5},
				Finalized: true,
			},
		}
		mockCurrentValidators := []*types.Validator{
			{
				Index:                      0,
				PublicKey:                  []byte{1, 2, 3},
				WithdrawalCredentials:      []byte{4, 5, 6},
				EffectiveBalance:           32,
				Status:                     "2",
				Slashed:                    true,
				ActivationEligibilityEpoch: 3,
				ActivationEpoch:            3,
				ExitEpoch:                  3,
				WithdrawableEpoch:          3,
			},
		}
		var mockBlock = &types.Block{
			Status: 1,
			EpochAssignments: &types.EpochAssignments{
				ProposerAssignments: map[uint64]uint64{
					0: 1,
				},
				AttestorAssignments: map[string]uint64{
					"0": 1,
				},
				SyncAssignments: []uint64{
					1,
				},
			},
			Slot: 111,
			Validators: []*types.Validator{
				{
					Index:                      0,
					PublicKey:                  []byte{1, 2, 3},
					WithdrawalCredentials:      []byte{5, 6, 7},
					EffectiveBalance:           32,
					Status:                     "1",
					Slashed:                    false,
					ActivationEligibilityEpoch: 4,
					ActivationEpoch:            4,
					ExitEpoch:                  4,
					WithdrawableEpoch:          4,
				},
			},
			AttestationDuties: map[types.ValidatorIndex][]types.Slot{
				0: {1},
			},
			SyncDuties: map[types.ValidatorIndex]bool{
				0: true,
			},
		}
		mockActivationEpochVal := []edb.ActivationEpochValidator{
			{
				ActivationEpoch: 3,
				ValidatorIndex:  0,
			},
		}
		mockUpdateCount := 7
		queries := fmt.Sprintf("UPDATE validators SET status = %s WHERE validatorindex = %d;\n", "2", 0)
		mockBalances := map[uint64]uint64{
			0: 1,
		}
		mockValidatorBalance := map[uint64][]*types.ValidatorBalance{
			1: {
				{
					Epoch:            3,
					Balance:          32,
					EffectiveBalance: 32,
				},
			},
		}
		mockParticipationStats := &types.ValidatorParticipation{
			Epoch:                   3,
			GlobalParticipationRate: 0.5,
			VotedEther:              1,
		}
		mockLastAttestationCache := map[uint64]uint64{
			0: 1,
		}

		utils.Config = &types.Config{
			Chain: types.Chain{
				ClConfig: types.ClChainConfig{
					DepositChainID: 1,
					SlotsPerEpoch:  32,
				},
			},
		}

		mockDB := new(dbmocks.SlotExporterDBRepository)
		mockBT := new(dbmocks.SlotExporterBTRepository)
		mockClient := new(rpcmocks.SlotExporterClient)
		mockCache := new(dbmocks.SlotExporterCacheRepository)
		mockTx := new(sqlx.Tx)

		// mock DB calls
		mockDB.On("BeginTx").Return(mockTx, nil)
		mockDB.On("RollbackTx", mockTx).Return(nil)
		mockDB.On("CommitTx", mockTx).Return(nil)
		mockDB.On("GetLastSlot", mockTx).Return(mockDBLastSlot, nil)
		mockDB.On("SaveBlock", mockBlock, false, mockTx).Return(nil)
		mockDB.On("GetValidatorsCurrentState", mockTx).Return(mockCurrentValidators, nil)
		mockDB.On("GetValidatorsWithMissingBalances", 10000, mockTx).Return(mockActivationEpochVal, nil)
		mockDB.On("AnalyzeValidatorsTable", mockTx).Return(nil)
		mockDB.On("CacheBlockDepositLookup").Return(nil)
		mockDB.On("UpdateQueueDeposits", mockTx).Return(nil)
		mockDB.On("SaveEpoch", mockChainHead.HeadEpoch, mockBlock.Validators, mockTx).Return(nil)
		mockDB.On("GetAllNonFinalizedSlots").Return(mockNonFinalSlots, nil)
		mockDB.On("UpdateActivationEpochBalance", mockActivationEpochVal[0].ValidatorIndex, mockBalances[0], mockTx).Return(nil)
		mockDB.On("UpdateEpochStatus", mockParticipationStats, mockTx).Return(nil)
		mockDB.On("PrepareValidatorsUpdate", mockCurrentValidators[0], mockBlock.Validators[0], mockTx).Return(mockUpdateCount, queries, nil)
		mockDB.On("SaveValidatorsFieldsUpdate", queries, mockUpdateCount, mockTx).Return(nil)

		// mock BT calls
		mockBT.On("SaveAttestationDuties", "mock.Anything").Return(nil)
		mockBT.On("SaveSyncCommitteeDuties", "mock.Anything").Return(nil)
		mockBT.On("SaveValidatorBalances", mockChainHead.HeadEpoch, mockBlock.Validators).Return(nil)
		mockBT.On("GetLastAttestationCacheMux").Return(&sync.Mutex{})
		mockBT.On("GetLastAttestationCache").Return(mockLastAttestationCache)
		mockBT.On("GetValidatorBalanceHistory",
			[]uint64{mockActivationEpochVal[0].ValidatorIndex},
			mockActivationEpochVal[0].ActivationEpoch,
			mockActivationEpochVal[0].ActivationEpoch).Return(mockValidatorBalance, nil)

		// mock Client calls
		mockClient.On("GetChainHead").Return(mockChainHead, nil)
		mockClient.On("GetBlockBySlot", mockChainHead.HeadSlot).Return(mockBlock, nil)
		mockClient.On("GetBlockHeader", mockNonFinalSlots[0].Slot).Return(&constypes.StandardBeaconHeaderResponse{}, nil)
		mockClient.On("GetEpochAssignments", mockChainHead.HeadEpoch+1).Return(mockBlock.EpochAssignments, nil)
		mockClient.On("GetBalancesForEpoch", int64(mockActivationEpochVal[0].ActivationEpoch)).Return(mockBalances, nil)
		mockClient.On("GetValidatorParticipation", mockChainHead.HeadEpoch-1).Return(mockParticipationStats, nil)

		// mock Cache calls
		mockCache.On("GetLatestEpoch").Return(mockChainHead.HeadEpoch-1, nil)
		mockCache.On("GetLatestSlot").Return(mockChainHead.HeadSlot-1, nil)
		mockCache.On("GetLatestFinalizedEpoch").Return(mockChainHead.FinalizedEpoch-1, nil)
		mockCache.On("GetLatestProposedSlot").Return(mockBlock.Slot-1, nil)
		mockCache.On("SetLatestEpoch", mockChainHead.HeadEpoch).Return(nil)
		mockCache.On("SetLatestSlot", mockChainHead.HeadSlot).Return(nil)
		mockCache.On("SetValidatorMapping", "mock.Anything", time.Duration(0)).Return(nil)
		mockCache.On("SetLatestProposedSlot", mockChainHead.HeadSlot).Return(nil)

		exporter := &slotExporter{
			Client:   mockClient,
			cache:    mockCache,
			db:       mockDB,
			bt:       mockBT,
			firstRun: false,
		}

		err := exporter.OnHead(nil)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		mockDB.AssertExpectations(t)
		mockBT.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}
