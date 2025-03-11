package modules

import (
	"testing"

	rpcmocks "github.com/gobitfly/beaconchain/pkg/commons/rpc/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	dbmocks "github.com/gobitfly/beaconchain/pkg/exporter/db/mocks"
	"github.com/jmoiron/sqlx"
)

func TestSlotExporterOnHead(t *testing.T) {
	tests := []struct {
		name              string
		mockChainHead     *types.ChainHead
		mockDBSlots       []uint64
		mockDBLastSlot    uint64
		mockNonFinalSlots []*edb.NonFinalizedSlotsRow
		mockBlock         *types.Block
		expectError       bool
	}{
		{
			name: "genesis slot",
			mockChainHead: &types.ChainHead{
				HeadSlot:       1,
				HeadEpoch:      1,
				FinalizedSlot:  1,
				FinalizedEpoch: 1,
			},
			mockDBSlots:       []uint64{},
			mockDBLastSlot:    0,
			mockNonFinalSlots: []*edb.NonFinalizedSlotsRow{},
			mockBlock: &types.Block{
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
			},
			expectError: false,
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(dbmocks.SlotExporterDBRepository)
			mockBT := new(dbmocks.SlotExporterBTRepository)
			mockClient := new(rpcmocks.SlotExporterClient)
			mockCache := new(dbmocks.SlotExporterCacheRepository)
			mockTx := new(sqlx.Tx)

			// mock DB calls
			mockDB.On("BeginTx").Return(mockTx, nil)
			mockDB.On("RollbackTx", mockTx).Return(nil)
			mockDB.On("CommitTx", mockTx).Return(nil)
			mockDB.On("GetAllSlots", mockTx).Return(tt.mockDBSlots, nil)
			mockDB.On("GetLastSlot", mockTx).Return(tt.mockDBLastSlot, nil)
			mockDB.On("SaveBlock", tt.mockBlock, false, mockTx).Return(nil)
			mockDB.On("SaveEpoch", utils.EpochOfSlot(tt.mockChainHead.HeadEpoch), tt.mockBlock.Validators, mockTx).Return(nil)
			mockDB.On("GetAllNonFinalizedSlots").Return(tt.mockNonFinalSlots, nil)

			// mock BT calls
			mockBT.On("SaveAttestationDuties", "mock.Anything").Return(nil)
			mockBT.On("SaveSyncCommitteeDuties", "mock.Anything").Return(nil)
			mockBT.On("SaveValidatorBalances", utils.EpochOfSlot(tt.mockChainHead.HeadEpoch), tt.mockBlock.Validators).Return(nil)

			// mock Client calls
			mockClient.On("GetChainHead").Return(tt.mockChainHead, nil)
			mockClient.On("GetBlockBySlot", tt.mockChainHead.HeadSlot).Return(tt.mockBlock, nil)

			// mock Cache calls
			mockCache.On("GetLatestEpoch").Return(tt.mockChainHead.HeadEpoch-1, nil)
			mockCache.On("GetLatestSlot").Return(tt.mockChainHead.HeadSlot-1, nil)
			mockCache.On("GetLatestFinalizedEpoch").Return(tt.mockChainHead.FinalizedEpoch-1, nil)
			mockCache.On("GetLatestProposedSlot").Return(tt.mockBlock.Slot-1, nil)
			mockCache.On("SetLatestSlot", tt.mockChainHead.HeadSlot).Return(nil)
			mockCache.On("SetLatestProposedSlot", tt.mockChainHead.HeadSlot).Return(nil)

			exporter := &slotExporter{
				Client:         mockClient,
				cache:          mockCache,
				db:             mockDB,
				bt:             mockBT,
				firstRun:       true,
				finalizedEpoch: 0,
				latestProposed: 0,
				latestEpoch:    0,
				latestSlot:     0,
			}

			err := exporter.OnHead(nil)

			if err != nil {
				if !tt.expectError {
					t.Errorf("expected no error, got %v", err)
				}
			}

			if err == nil {
				if tt.expectError {
					t.Errorf("expected error, got nil")
				}
			}

			mockDB.AssertExpectations(t)
			mockBT.AssertExpectations(t)
			mockClient.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}
