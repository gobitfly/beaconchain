package modules

import (
	"testing"

	"github.com/gobitfly/beaconchain/pkg/commons/db/mocks"
	rpcmocks "github.com/gobitfly/beaconchain/pkg/commons/rpc/mocks"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/pkg/errors"
)

func TestProcessGenesisDeposits(t *testing.T) {
	tests := []struct {
		name                     string
		mockEpochResponse        uint64
		mockDepositCountResponse uint64
		mockValidatorStateError  error
		expectedShouldSleep      bool
		expectedError            bool
	}{
		{
			name:                     "valid export",
			mockEpochResponse:        1,
			mockDepositCountResponse: 10,
			expectedShouldSleep:      false,
		},
		{
			name:                    "GetValidatorState error",
			mockValidatorStateError: errors.New("error"),
			expectedShouldSleep:     true,
			expectedError:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRPCClient := new(rpcmocks.Client)
			mockRPCClient.On("GetValidatorState", uint64(0)).Return(genesisValidators, tt.mockValidatorStateError)

			mockConsDBClient := new(mocks.ConsensusDBI)
			mockConsDBClient.On("GetLatestEpoch").Return(tt.mockEpochResponse, nil)
			mockConsDBClient.On("GetDepositsCountForBlockSlot").Return(tt.mockDepositCountResponse, nil)
			mockConsDBClient.On("SaveBlockDeposits",
				genesisValidators.Data[0].Index,
				[]byte(genesisValidators.Data[0].Validator.Pubkey),
				[]byte(genesisValidators.Data[0].Validator.WithdrawalCredentials),
				genesisValidators.Data[0].Balance,
			).Return(nil)
			mockConsDBClient.On("UpdateBlockDepositsSignature").Return(nil)
			mockConsDBClient.On("UpdateBlockDepositCount", tt.mockDepositCountResponse).Return(nil)

			shouldSleep, err := processGenesisDeposits(mockRPCClient, mockConsDBClient)

			if err != nil {
				if tt.expectedError {
					return
				}
				t.Errorf("expected no error, got: %v", err)
			}
			if tt.expectedError {
				t.Error("expected error, got nil")
			}
			if shouldSleep != tt.expectedShouldSleep {
				t.Errorf("expected shouldSleep: %v, got: %v", tt.expectedShouldSleep, shouldSleep)
			}
		})
	}
}

var genesisValidators = &types.StandardValidatorsResponse{
	Data: []types.StandardValidator{
		{
			Index: 1,
			Validator: types.Validator{
				Pubkey:                []byte("0xabc"),
				WithdrawalCredentials: []byte("0xdef"),
			},
			Balance: 1000,
		},
	},
}
