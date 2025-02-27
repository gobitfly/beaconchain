package modules

import (
	"context"
	"testing"
	"time"

	dbmocks "github.com/gobitfly/beaconchain/pkg/commons/db/mocks"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc/mocks"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
)

func TestGenesisDepositsExporter_Export(t *testing.T) {
	tests := []struct {
		name                     string
		mockEpochResponse        uint64
		mockDepositCountResponse uint64
	}{
		{
			name:                     "no records in db",
			mockEpochResponse:        1,
			mockDepositCountResponse: 0,
		},
		{
			name:                     "records exist in db",
			mockEpochResponse:        1,
			mockDepositCountResponse: 1,
		},
	}

	mockConsDBClient := new(dbmocks.ConsensusDBI)
	mockClient := new(mocks.ValidatorClient)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	exporter := genesisDepositsExporter{
		client: mockClient,
		//db:     mockConsDBClient,
		offset: 0,
		ctx:    ctx,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConsDBClient.On("GetLatestEpoch").Return(tt.mockEpochResponse, nil)
			mockConsDBClient.On("GetDepositsCountForBlockSlot").Return(tt.mockDepositCountResponse, nil)

			if tt.mockDepositCountResponse == 0 {
				// mock if no records in db
				mockClient.On("GetValidatorState", uint64(0)).Return(genesisValidators, nil)
				mockConsDBClient.On("SaveBlockDeposits",
					genesisValidators.Data[0].Index,
					[]byte(genesisValidators.Data[0].Validator.Pubkey),
					[]byte(genesisValidators.Data[0].Validator.WithdrawalCredentials),
					genesisValidators.Data[0].Balance,
				).Return(nil)
				mockConsDBClient.On("UpdateBlockDepositsSignature").Return(nil)
				mockConsDBClient.On("UpdateBlockDepositCount", 1).Return(nil)
			}

			exporter.Export()

			mockConsDBClient.AssertCalled(t, "GetLatestEpoch")
			mockConsDBClient.AssertCalled(t, "GetDepositsCountForBlockSlot")

			if tt.mockDepositCountResponse == 0 {
				// assert mock calls if no records in db
				mockClient.AssertCalled(t, "GetValidatorState", uint64(0))
				mockConsDBClient.AssertCalled(t, "SaveBlockDeposits",
					genesisValidators.Data[0].Index,
					[]byte(genesisValidators.Data[0].Validator.Pubkey),
					[]byte(genesisValidators.Data[0].Validator.WithdrawalCredentials),
					genesisValidators.Data[0].Balance,
				)
				mockConsDBClient.AssertCalled(t, "UpdateBlockDepositsSignature")
				mockConsDBClient.AssertCalled(t, "UpdateBlockDepositCount", 1)
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
