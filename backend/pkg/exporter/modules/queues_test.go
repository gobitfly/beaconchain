package modules

import (
	"context"
	"database/sql"
	"encoding/hex"
	"testing"
	"time"

	dbmocks "github.com/gobitfly/beaconchain/pkg/commons/db2/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
)

func TestQueueExport(t *testing.T) {
	tests := []struct {
		name                        string
		mockStateID                 string
		mockChainHead               *types.ChainHead
		mockValidatorState          *constypes.StandardValidatorsResponse
		mockPendingDepositsResponse *constypes.StandardBeaconPendingDepositsResponse
		mockExpectedDBResult        []types.PendingDeposit
	}{
		{
			name:          "topup on empty queue",
			mockStateID:   "head",
			mockChainHead: &types.ChainHead{HeadEpoch: 10},
			mockValidatorState: &constypes.StandardValidatorsResponse{
				Data: []constypes.StandardValidator{
					{
						Index:   1,
						Balance: 1000000000,
						Status:  constypes.ActiveOngoing,
						Validator: constypes.Validator{

							Pubkey:                     decodeHex("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
							WithdrawalCredentials:      decodeHex("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
							EffectiveBalance:           1000000000,
							Slashed:                    false,
							ActivationEligibilityEpoch: 1,
							ActivationEpoch:            1,
							ExitEpoch:                  1,
							WithdrawableEpoch:          1,
						},
					},
				},
			},
			mockPendingDepositsResponse: &constypes.StandardBeaconPendingDepositsResponse{
				Data: []constypes.StandardBeaconPendingDepositsData{
					{
						Pubkey:                decodeHex("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
						WithdrawalCredentials: decodeHex("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
						Amount:                1000000000,
						Signature:             decodeHex("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
						Slot:                  1,
					},
				},
			},
			mockExpectedDBResult: []types.PendingDeposit{
				{
					ID:                    0,
					Pubkey:                decodeHex("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					WithdrawalCredentials: decodeHex("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					Amount:                1000000000,
					Signature:             decodeHex("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					Slot:                  1,
					ValidatorIndex: sql.NullInt64{
						Int64: 1,
						Valid: true,
					},
					QueuedBalanceAhead: 0,
					EstClearEpoch:      11,
				},
			},
		},
		{
			name:          "deposit on non-empty queue",
			mockStateID:   "head",
			mockChainHead: &types.ChainHead{HeadEpoch: 10},
			mockValidatorState: &constypes.StandardValidatorsResponse{
				Data: []constypes.StandardValidator{
					{
						Index:   1,
						Balance: 1000000000,
						Status:  constypes.ActiveOngoing,
						Validator: constypes.Validator{
							Pubkey:                     decodeHex("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
							WithdrawalCredentials:      decodeHex("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
							EffectiveBalance:           1000000000,
							Slashed:                    false,
							ActivationEligibilityEpoch: 1,
							ActivationEpoch:            1,
							ExitEpoch:                  99,
							WithdrawableEpoch:          99,
						},
					},
				},
			},
			mockPendingDepositsResponse: &constypes.StandardBeaconPendingDepositsResponse{
				Data: []constypes.StandardBeaconPendingDepositsData{
					{
						Pubkey:                decodeHex("2234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
						WithdrawalCredentials: decodeHex("2234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
						Amount:                32e9,
						Signature:             decodeHex("2234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
						Slot:                  1,
					},
					{
						Pubkey:                decodeHex("3234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
						WithdrawalCredentials: decodeHex("3234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
						Amount:                32e9,
						Signature:             decodeHex("3234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
						Slot:                  1,
					},
					{
						Pubkey:                decodeHex("4234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
						WithdrawalCredentials: decodeHex("4234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
						Amount:                320e9,
						Signature:             decodeHex("4234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
						Slot:                  1,
					},
				},
			},
			mockExpectedDBResult: []types.PendingDeposit{
				{
					ID:                    0,
					Pubkey:                decodeHex("2234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					WithdrawalCredentials: decodeHex("2234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					Amount:                32e9,
					Signature:             decodeHex("2234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					Slot:                  1,
					ValidatorIndex:        sql.NullInt64{},
					QueuedBalanceAhead:    0,
					EstClearEpoch:         11,
				},
				{
					ID:                    1,
					Pubkey:                decodeHex("3234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					WithdrawalCredentials: decodeHex("3234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					Amount:                32e9,
					Signature:             decodeHex("3234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					Slot:                  1,
					ValidatorIndex:        sql.NullInt64{},
					QueuedBalanceAhead:    32e9,
					EstClearEpoch:         11,
				},
				{
					ID:                    2,
					Pubkey:                decodeHex("4234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					WithdrawalCredentials: decodeHex("4234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					Amount:                320e9,
					Signature:             decodeHex("4234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					Slot:                  1,
					ValidatorIndex:        sql.NullInt64{},
					QueuedBalanceAhead:    64e9,
					EstClearEpoch:         13,
				},
			},
		},
	}

	utils.Config = &types.Config{
		Chain: types.Chain{
			ClConfig: types.ClChainConfig{
				AltairForkEpoch:                     1,
				ElectraForkEpoch:                    1,
				EpochsPerSyncCommitteePeriod:        32,
				SlotsPerEpoch:                       10,
				DepositChainID:                      1,
				SecondsPerSlot:                      12,
				ChurnLimitQuotient:                  65536,
				MinPerEpochChurnLimit:               4,
				EffectiveBalanceIncrement:           1000000000,
				MaxPerEpochActivationExitChurnLimit: 256000000000,
				MinActivationBalance:                32000000000,
				MaxEffectiveBalance:                 32000000000,
				MaxPendingDepositsPerEpoch:          16,
				MinPerEpochChurnLimitElectra:        128e9,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConsDBClient := new(dbmocks.ConsensusRepository)
			mockClient := new(mocks.QueueClient)

			mockClient.On("GetChainHead").Return(tt.mockChainHead, nil)
			mockClient.On("GetValidatorState", tt.mockChainHead.HeadEpoch).Return(tt.mockValidatorState, nil)
			mockClient.On("GetPendingDeposits", tt.mockStateID).Return(tt.mockPendingDepositsResponse, nil)
			mockConsDBClient.On("SavePendingDepositsQueue", tt.mockExpectedDBResult).Return(nil)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			exporter := pendingQueueExporter{
				lc:    mockClient,
				db:    mockConsDBClient,
				ctx:   ctx,
				delay: 100 * time.Millisecond,
			}

			exporter.Export()

			mockClient.AssertCalled(t, "GetChainHead")
			mockClient.AssertCalled(t, "GetValidatorState", tt.mockChainHead.HeadEpoch)
			mockClient.AssertCalled(t, "GetPendingDeposits", tt.mockStateID)

			mockConsDBClient.AssertCalled(t, "SavePendingDepositsQueue", tt.mockExpectedDBResult)

			cancel()
		})
	}
}

func decodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}
