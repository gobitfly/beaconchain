package modules

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func TestUpdatePubkeyTags(t *testing.T) {
	tests := []struct {
		name          string
		mockError     error
		expectedError bool
	}{
		{
			name:          "successful update",
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "database error",
			mockError:     errors.New("error"),
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

			// mock UpdatePubkeyTags query
			query := `INSERT INTO validator_tags (publickey, tag)
		SELECT publickey, FORMAT('pool:%s', sps.name) tag
		FROM eth1_deposits
		inner join stake_pools_stats as sps on ENCODE(from_address::bytea, 'hex')=sps.address
		WHERE sps.name NOT LIKE '%Rocketpool -%'
		ON CONFLICT (publickey, tag) DO NOTHING;`

			if tt.mockError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnError(tt.mockError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(query)).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			utils.Config = &types.Config{
				DeploymentType: "development",
			}

			err = updatePubkeyTagOnce()
			if !tt.expectedError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			}
		})
	}
}
