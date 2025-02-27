package modules

import (
	"context"
	"testing"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db/mocks"
	rpcmocks "github.com/gobitfly/beaconchain/pkg/commons/rpc/mocks"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/pkg/errors"
)

func TestGenesisDepositsExporter_Export(t *testing.T) {
	tests := []struct {
		name                        string
		mockEpochResponse           uint64
		mockDepositCountResponse    uint64
		mockBlockDepositCountUpdate int
		mockValidatorStateError     error
	}{
		{
			name:                        "valid export",
			mockEpochResponse:           1,
			mockDepositCountResponse:    0,
			mockBlockDepositCountUpdate: 1,
		},
		{
			name:                        "GetValidatorState error",
			mockEpochResponse:           1,
			mockDepositCountResponse:    1,
			mockBlockDepositCountUpdate: 1,
			mockValidatorStateError:     errors.New("error"),
		},
	}

	mockConsDBClient := new(mocks.ConsensusDBI)
	mockRPCClient := new(rpcmocks.ValidatorClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	exporter := genesisDepositsExporter{
		client: mockRPCClient,
		db:     mockConsDBClient,
		offset: 0,
		ctx:    ctx,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock expected calls
			mockConsDBClient.On("GetLatestEpoch").Return(tt.mockEpochResponse, nil)
			mockConsDBClient.On("GetDepositsCountForBlockSlot").Return(tt.mockDepositCountResponse, nil)
			mockRPCClient.On("GetValidatorState", uint64(0)).Return(genesisValidators, nil)
			mockConsDBClient.On("SaveBlockDeposits",
				genesisValidators.Data[0].Index,
				[]byte(genesisValidators.Data[0].Validator.Pubkey),
				[]byte(genesisValidators.Data[0].Validator.WithdrawalCredentials),
				genesisValidators.Data[0].Balance,
			).Return(nil)
			mockConsDBClient.On("UpdateBlockDepositsSignature").Return(nil)
			mockConsDBClient.On("UpdateBlockDepositCount", tt.mockBlockDepositCountUpdate).Return(nil)

			exporter.Export()

			mockConsDBClient.AssertCalled(t, "GetLatestEpoch")
			mockConsDBClient.AssertCalled(t, "GetDepositsCountForBlockSlot")
			mockRPCClient.AssertCalled(t, "GetValidatorState", uint64(0))
			mockConsDBClient.AssertCalled(t, "SaveBlockDeposits",
				genesisValidators.Data[0].Index,
				[]byte(genesisValidators.Data[0].Validator.Pubkey),
				[]byte(genesisValidators.Data[0].Validator.WithdrawalCredentials),
				genesisValidators.Data[0].Balance,
			)
			mockConsDBClient.AssertCalled(t, "UpdateBlockDepositsSignature")
			mockConsDBClient.AssertCalled(t, "UpdateBlockDepositCount", tt.mockBlockDepositCountUpdate)
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
