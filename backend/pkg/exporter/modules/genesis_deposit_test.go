package modules

import (
	"testing"

	"github.com/gobitfly/beaconchain/pkg/commons/db/mocks"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/pkg/errors"
)

func TestGetLatestEpoch(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   uint64
		mockError      error
		expectedResult uint64
		expectedError  bool
	}{
		{
			name:           "latest epoch > 0",
			mockResponse:   1,
			expectedResult: 1,
		},
		{
			name:           "latest epoch = 0",
			mockResponse:   0,
			expectedResult: 0,
		},
		{
			name:           "database error",
			mockResponse:   0,
			mockError:      errors.New("error"),
			expectedResult: 0,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mocks.ConsensusDBI)
			mockClient.On("GetLatestEpoch").Return(tt.mockResponse, tt.mockError)

			result, err := getLatestEpoch(mockClient)

			if err != nil {
				if tt.expectedError {
					return
				}
				t.Errorf("expected no error, got: %v", err)
			}
			if tt.expectedError {
				t.Error("expected error, got nil")
			}
			if result != tt.expectedResult {
				t.Errorf("expected result: %v, got: %v", tt.expectedResult, result)
			}
		})
	}
}

func TestGetGenesisDepositCount(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   uint64
		mockError      error
		expectedResult uint64
		expectedError  bool
	}{
		{
			name:           "deposit count > 0",
			mockResponse:   10,
			expectedResult: 10,
		},
		{
			name:           "deposit epoch = 0",
			mockResponse:   0,
			expectedResult: 0,
		},
		{
			name:           "database error",
			mockResponse:   0,
			mockError:      errors.New("error"),
			expectedResult: 0,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mocks.ConsensusDBI)
			mockClient.On("GetDepositsCountForBlockSlot").Return(tt.mockResponse, tt.mockError)

			result, err := getGenesisDepositCount(mockClient)

			if err != nil {
				if tt.expectedError {
					return
				}
				t.Errorf("expected no error, got: %v", err)
			}
			if tt.expectedError {
				t.Error("expected error, got nil")
			}
			if result != tt.expectedResult {
				t.Errorf("expected result: %v, got: %v", tt.expectedResult, result)
			}
		})
	}
}

func TestExportGenesisDeposits(t *testing.T) {
	tests := []struct {
		name                            string
		genesisValidators               *types.StandardValidatorsResponse
		mockBlockDepositsError          error
		mockBlockDepositsSignatureError error
		mockBlockDepositsCountError     error
		expectedError                   bool
	}{
		{
			name:              "valid genesis validators export",
			genesisValidators: genesisValidators,
			expectedError:     false,
		},
		{
			name:                   "SaveBlockDeposits error",
			genesisValidators:      genesisValidators,
			mockBlockDepositsError: errors.New("error"),
			expectedError:          true,
		},
		{
			name:                            "UpdateBlockDepositsSignature error",
			genesisValidators:               genesisValidators,
			mockBlockDepositsSignatureError: errors.New("error"),
			expectedError:                   true,
		},
		{
			name:                        "UpdateBlockDepositCount error",
			genesisValidators:           genesisValidators,
			mockBlockDepositsCountError: errors.New("error"),
			expectedError:               true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mocks.ConsensusDBI)
			mockClient.On("SaveBlockDeposits",
				tt.genesisValidators.Data[0].Index,
				[]byte(tt.genesisValidators.Data[0].Validator.Pubkey),
				[]byte(tt.genesisValidators.Data[0].Validator.WithdrawalCredentials),
				tt.genesisValidators.Data[0].Balance,
			).Return(tt.mockBlockDepositsError)
			mockClient.On("UpdateBlockDepositsSignature").Return(tt.mockBlockDepositsSignatureError)
			mockClient.On("UpdateBlockDepositCount", 1).Return(tt.mockBlockDepositsCountError)

			err := exportGenesisDeposits(tt.genesisValidators, mockClient)

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
