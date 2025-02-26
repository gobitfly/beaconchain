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
		name                            string
		mockEpochResponse               uint64
		mockValidatorStateError         error
		mockLatestEpochError            error
		mockDepositCountError           error
		mockBlockDepositsError          error
		mockBlockDepositsSignatureError error
		mockBlockDepositsCountError     error
		expectedError                   bool
	}{
		{
			name:              "valid export",
			mockEpochResponse: 1,
		},
		{
			name:                    "GetValidatorState error",
			mockValidatorStateError: errors.New("error"),
			expectedError:           true,
		},
		{
			name:                 "GetLatestEpoch error",
			mockLatestEpochError: errors.New("error"),
			expectedError:        true,
		},
		{
			name:                  "GetDepositsCountForBlockSlot error",
			mockDepositCountError: errors.New("error"),
			expectedError:         true,
		},
		{
			name:                   "SaveBlockDeposits error",
			mockBlockDepositsError: errors.New("error"),
			expectedError:          true,
		},
		{
			name:                            "UpdateBlockDepositsSignature error",
			mockBlockDepositsSignatureError: errors.New("error"),
			expectedError:                   true,
		},
		{
			name:                        "UpdateBlockDepositCount error",
			mockBlockDepositsCountError: errors.New("error"),
			expectedError:               true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRPCClient := new(rpcmocks.Client)
			mockRPCClient.On("GetValidatorState", uint64(0)).Return(genesisValidators, tt.mockValidatorStateError)

			mockConsDBClient := new(mocks.ConsensusDBI)
			mockConsDBClient.On("GetLatestEpoch").Return(tt.mockEpochResponse, tt.mockLatestEpochError)
			mockConsDBClient.On("GetDepositsCountForBlockSlot").Return(uint64(10), tt.mockDepositCountError)
			mockConsDBClient.On("SaveBlockDeposits",
				genesisValidators.Data[0].Index,
				[]byte(genesisValidators.Data[0].Validator.Pubkey),
				[]byte(genesisValidators.Data[0].Validator.WithdrawalCredentials),
				genesisValidators.Data[0].Balance,
			).Return(tt.mockBlockDepositsError)
			mockConsDBClient.On("UpdateBlockDepositsSignature").Return(tt.mockBlockDepositsSignatureError)
			mockConsDBClient.On("UpdateBlockDepositCount", 1).Return(tt.mockBlockDepositsCountError)

			err := processGenesisDeposits(mockRPCClient, mockConsDBClient)

			if err != nil {
				if tt.expectedError {
					return
				}
				t.Errorf("expected no error, got: %v", err)
			}
			if tt.expectedError {
				t.Error("expected error, got nil")
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
