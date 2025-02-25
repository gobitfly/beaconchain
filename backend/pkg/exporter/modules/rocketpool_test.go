package modules

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	smartnodeCfg "github.com/rocket-pool/smartnode/shared/services/config"
	smartnodeNetwork "github.com/rocket-pool/smartnode/shared/types/config"
)

func TestSaveConfig(t *testing.T) {
	tests := []struct {
		name          string
		mockSaveError error
		expectedError bool
	}{
		{
			name:          "successful save",
			expectedError: false,
		},
		{
			name:          "SaveRocketPoolDAOProposals error",
			mockSaveError: errors.New("error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer dbMock.Close()

			sqlxDB := sqlx.NewDb(dbMock, "sqlmock")
			db.WriterDb = sqlxDB

			rpconfig := smartnodeCfg.NewSmartnodeConfig(&smartnodeCfg.RocketPoolConfig{
				RocketPoolDirectory: "/tmp/rocketpool",
			})
			rpconfig.Network.Value = smartnodeNetwork.Network_Mainnet

			rp := &RocketpoolExporter{
				OnchainConfig: &RocketPoolOnchainConfig{
					SmoothingPoolAddress: common.HexToAddress("0x001"),
				},
				RPConfig: rpconfig,
			}

			// mock SaveRocketPoolConfig query
			if tt.mockSaveError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(saveConfigsQ)).WillReturnError(tt.mockSaveError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(saveConfigsQ)).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err = rp.SaveConfigs()

			if !tt.expectedError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error got nil")
				}
			}

		})
	}
}

var (
	saveConfigsQ = `INSERT INTO rocketpool_onchain_configs (
			rocketpool_storage_address,
			smoothing_pool_address
		) VALUES ($1, $2) ON CONFLICT (rocketpool_storage_address) DO UPDATE SET
		 	smoothing_pool_address = excluded.smoothing_pool_address
	`
)
