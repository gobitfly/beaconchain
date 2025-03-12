package modules

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/gob"
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
	"github.com/klauspost/pgzip"
)

func TestSlotExporter(t *testing.T) {
	utils.Config = &types.Config{
		Chain: types.Chain{
			ClConfig: types.ClChainConfig{
				DepositChainID: 1,
				SlotsPerEpoch:  32,
			},
		},
	}

	t.Run("genesis slot", func(t *testing.T) {
		mockChainHead := &types.ChainHead{
			HeadSlot:       1,
			HeadEpoch:      1,
			FinalizedSlot:  1,
			FinalizedEpoch: 1,
		}
		mockDBSlots := []uint64{0}
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
					Status:                     "pending",
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
		mockDB.On("SaveEpoch", utils.EpochOfSlot(mockBlock.Slot), mockBlock.Validators, mockTx).Return(nil)
		mockDB.On("GetAllNonFinalizedSlots").Return(mockNonFinalSlots, nil)

		// mock BT calls
		mockBT.On("SaveAttestationDuties", "mock.Anything").Return(nil)
		mockBT.On("SaveSyncCommitteeDuties", "mock.Anything").Return(nil)
		mockBT.On("SaveValidatorBalances", utils.EpochOfSlot(mockBlock.Slot), mockBlock.Validators).Return(nil)

		// mock Client calls
		mockClient.On("GetChainHead").Return(mockChainHead, nil)
		mockClient.On("GetBlockBySlot", mockDBSlots[0]+1).Return(mockBlock, nil)

		// mock Cache calls
		mockCache.On("GetLatestEpoch").Return(uint64(0), nil)
		mockCache.On("GetLatestSlot").Return(uint64(0), nil)
		mockCache.On("GetLatestFinalizedEpoch").Return(uint64(0), nil)
		mockCache.On("GetLatestProposedSlot").Return(uint64(0), nil)
		mockCache.On("SetLatestSlot", mockChainHead.HeadSlot).Return(nil)
		mockCache.On("SetLatestProposedSlot", mockBlock.Slot).Return(nil)

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
					Balance:                    0,
					EffectiveBalance:           32,
					Status:                     "pending",
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
		mockCachedValidator := &types.RedisCachedValidatorsMapping{
			Epoch: types.Epoch(utils.EpochOfSlot(mockBlock.Slot)),
			Mapping: []*types.CachedValidator{
				{
					PublicKey:                  mockBlock.Validators[0].PublicKey,
					WithdrawalCredentials:      mockBlock.Validators[0].WithdrawalCredentials,
					Balance:                    mockBlock.Validators[0].Balance,
					EffectiveBalance:           mockBlock.Validators[0].EffectiveBalance,
					Status:                     mockBlock.Validators[0].Status,
					Slashed:                    mockBlock.Validators[0].Slashed,
					ActivationEligibilityEpoch: sql.NullInt64{Int64: int64(mockBlock.Validators[0].ActivationEligibilityEpoch), Valid: true},
					ActivationEpoch:            sql.NullInt64{Int64: int64(mockBlock.Validators[0].ActivationEpoch), Valid: true},
					ExitEpoch:                  sql.NullInt64{Int64: int64(mockBlock.Validators[0].ExitEpoch), Valid: true},
					WithdrawableEpoch:          sql.NullInt64{Int64: int64(mockBlock.Validators[0].WithdrawableEpoch), Valid: true},
					Queues: types.QueuesMetadata{
						ActivationIndex: sql.NullInt64{Int64: int64(mockBlock.Validators[0].Index), Valid: true},
					},
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
		mockDB.On("SaveEpoch", utils.EpochOfSlot(mockBlock.Slot), mockBlock.Validators, mockTx).Return(nil)
		mockDB.On("GetAllNonFinalizedSlots").Return(mockNonFinalSlots, nil)
		mockDB.On("UpdateActivationEpochBalance", mockActivationEpochVal[0].ValidatorIndex, mockBalances[0], mockTx).Return(nil)
		mockDB.On("UpdateEpochStatus", mockParticipationStats, mockTx).Return(nil)
		mockDB.On("PrepareValidatorsUpdate", mockCurrentValidators[0], mockBlock.Validators[0], mockTx).Return(mockUpdateCount, queries, nil)
		mockDB.On("SaveValidatorsFieldsUpdate", queries, mockUpdateCount, mockTx).Return(nil)

		// mock BT calls
		mockBT.On("SaveAttestationDuties", "mock.Anything").Return(nil)
		mockBT.On("SaveSyncCommitteeDuties", "mock.Anything").Return(nil)
		mockBT.On("SaveValidatorBalances", utils.EpochOfSlot(mockBlock.Slot), mockBlock.Validators).Return(nil)
		mockBT.On("GetLastAttestationCacheMux").Return(&sync.Mutex{})
		mockBT.On("GetLastAttestationCache").Return(mockLastAttestationCache)
		mockBT.On("GetValidatorBalanceHistory",
			[]uint64{mockActivationEpochVal[0].ValidatorIndex},
			mockActivationEpochVal[0].ActivationEpoch,
			mockActivationEpochVal[0].ActivationEpoch).Return(mockValidatorBalance, nil)

		// mock Client calls
		mockClient.On("GetChainHead").Return(mockChainHead, nil)
		mockClient.On("GetBlockBySlot", mockNonFinalSlots[0].Slot).Return(mockBlock, nil)
		mockClient.On("GetBlockHeader", mockNonFinalSlots[0].Slot).Return(&constypes.StandardBeaconHeaderResponse{}, nil)
		mockClient.On("GetEpochAssignments", utils.EpochOfSlot(mockBlock.Slot)+1).Return(mockBlock.EpochAssignments, nil)
		mockClient.On("GetBalancesForEpoch", int64(mockActivationEpochVal[0].ActivationEpoch)).Return(mockBalances, nil)
		mockClient.On("GetValidatorParticipation", utils.EpochOfSlot(mockNonFinalSlots[0].Slot)-1).Return(mockParticipationStats, nil)

		// mock Cache calls
		mockCache.On("GetLatestEpoch").Return(utils.EpochOfSlot(mockChainHead.HeadSlot)-1, nil)
		mockCache.On("GetLatestSlot").Return(mockChainHead.HeadSlot-1, nil)
		mockCache.On("GetLatestFinalizedEpoch").Return(mockParticipationStats.Epoch-1, nil)
		mockCache.On("GetLatestProposedSlot").Return(mockBlock.Slot-1, nil)
		mockCache.On("SetLatestEpoch", utils.EpochOfSlot(mockChainHead.HeadSlot)).Return(nil)
		mockCache.On("SetLatestSlot", mockChainHead.HeadSlot).Return(nil)
		mockCache.On("SetLatestProposedSlot", mockBlock.Slot).Return(nil)
		valMapping, err := compressValidatorMapping(mockCachedValidator)
		if err != nil {
			t.Errorf("error compressing validator mapping: %v", err)
		}
		mockCache.On("SetValidatorMapping", valMapping, time.Duration(0)).Return(nil)

		exporter := &slotExporter{
			Client:   mockClient,
			cache:    mockCache,
			db:       mockDB,
			bt:       mockBT,
			firstRun: false,
		}

		err = exporter.OnHead(nil)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		mockDB.AssertExpectations(t)
		mockBT.AssertExpectations(t)
		mockClient.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("non-head epoch", func(t *testing.T) {
		mockChainHead := &types.ChainHead{
			HeadSlot:       111,
			HeadEpoch:      10,
			FinalizedSlot:  111,
			FinalizedEpoch: 10,
		}
		mockDBLastSlot := uint64(110)
		mockNonFinalSlots := []*edb.NonFinalizedSlotsRow{
			{
				Slot:      111,
				BlockRoot: []byte{1, 2, 3, 4, 5},
				Finalized: false,
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
					Status:                     "pending",
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
		mockParticipationStats := &types.ValidatorParticipation{
			Epoch:                   10,
			GlobalParticipationRate: 0.5,
			VotedEther:              1,
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
		mockDB.On("SaveEpoch", utils.EpochOfSlot(mockBlock.Slot), mockBlock.Validators, mockTx).Return(nil)
		mockDB.On("GetAllNonFinalizedSlots").Return(mockNonFinalSlots, nil)
		mockDB.On("UpdateEpochStatus", mockParticipationStats, mockTx).Return(nil)
		mockDB.On("SetSlotFinalizationAndStatus", mockNonFinalSlots[0].Slot, mockNonFinalSlots[0].Slot <= mockChainHead.FinalizedSlot, "3", mockTx).Return(nil)

		// mock BT calls
		mockBT.On("SaveAttestationDuties", "mock.Anything").Return(nil)
		mockBT.On("SaveSyncCommitteeDuties", "mock.Anything").Return(nil)
		mockBT.On("SaveValidatorBalances", utils.EpochOfSlot(mockBlock.Slot), mockBlock.Validators).Return(nil)

		// mock Client calls
		mockClient.On("GetChainHead").Return(mockChainHead, nil)
		mockClient.On("GetBlockBySlot", mockChainHead.HeadSlot).Return(mockBlock, nil)
		mockClient.On("GetBlockHeader", mockNonFinalSlots[0].Slot).Return(&constypes.StandardBeaconHeaderResponse{}, nil)
		mockClient.On("GetValidatorParticipation", utils.EpochOfSlot(mockBlock.Slot)-1).Return(mockParticipationStats, nil)

		// mock Cache calls
		mockCache.On("GetLatestEpoch").Return(utils.EpochOfSlot(mockChainHead.HeadSlot)-1, nil)
		mockCache.On("GetLatestSlot").Return(mockChainHead.HeadSlot-1, nil)
		mockCache.On("GetLatestFinalizedEpoch").Return(mockParticipationStats.Epoch-1, nil)
		mockCache.On("GetLatestProposedSlot").Return(mockBlock.Slot-1, nil)
		mockCache.On("SetLatestEpoch", utils.EpochOfSlot(mockChainHead.HeadSlot)).Return(nil)
		mockCache.On("SetLatestSlot", mockChainHead.HeadSlot).Return(nil)
		mockCache.On("SetLatestProposedSlot", mockBlock.Slot).Return(nil)

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

func compressValidatorMapping(mapping *types.RedisCachedValidatorsMapping) ([]byte, error) {
	var serialized bytes.Buffer
	enc := gob.NewEncoder(&serialized)
	if err := enc.Encode(mapping); err != nil {
		return nil, fmt.Errorf("error serializing validator mapping: %w", err)
	}

	var compressed bytes.Buffer
	w, err := pgzip.NewWriterLevel(&compressed, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("error creating pgzip writer: %w", err)
	}
	err = w.SetConcurrency(500_000, 10)
	if err != nil {
		return nil, fmt.Errorf("error setting concurrency: %w", err)
	}
	if _, err := w.Write(serialized.Bytes()); err != nil {
		return nil, fmt.Errorf("error compressing data: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("error closing pgzip writer: %w", err)
	}

	return compressed.Bytes(), nil
}
