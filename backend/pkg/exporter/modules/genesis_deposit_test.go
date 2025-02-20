package modules

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func TestIsBeaconChainStarted(t *testing.T) {
	tests := []struct {
		name           string
		mockRows       *sqlmock.Rows
		mockError      error
		expectedResult bool
	}{
		{
			name:           "beacon chain started",
			mockRows:       sqlmock.NewRows([]string{"epoch"}).AddRow(10),
			mockError:      nil,
			expectedResult: true,
		},
		{
			name:           "beacon chain not started",
			mockRows:       sqlmock.NewRows([]string{"epoch"}).AddRow(0),
			mockError:      nil,
			expectedResult: false,
		},
		{
			name:           "database error",
			mockRows:       nil,
			mockError:      errors.New("database error"),
			expectedResult: false,
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

			// mock GetLatestEpoch query
			query := `SELECT COALESCE\(MAX\(epoch\), 0\) FROM epochs`
			if tt.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(query).WillReturnRows(tt.mockRows)
			}

			result := isBeaconChainStarted()

			if result != tt.expectedResult {
				t.Errorf("expected result: %v, got: %v", tt.expectedResult, result)
			}
		})
	}
}

func TestAreGenesisDepositsExported(t *testing.T) {
	tests := []struct {
		name           string
		mockRows       *sqlmock.Rows
		mockError      error
		expectedResult bool
	}{
		{
			name:           "genesis deposits exported",
			mockRows:       sqlmock.NewRows([]string{"count"}).AddRow(10),
			mockError:      nil,
			expectedResult: true,
		},
		{
			name:           "genesis deposits not exported",
			mockRows:       sqlmock.NewRows([]string{"count"}).AddRow(0),
			mockError:      nil,
			expectedResult: false,
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

			// mock GetDepositsCountForBlockSlot query
			query := `SELECT COUNT\(\*\) FROM blocks_deposits WHERE block_slot=0`
			if tt.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(query).WillReturnRows(tt.mockRows)
			}

			result := areGenesisDepositsExported()

			if result != tt.expectedResult {
				t.Errorf("expected result: %v, got: %v", tt.expectedResult, result)
			}
		})
	}
}

func TestExportGenesisDeposits(t *testing.T) {
	tests := []struct {
		name                 string
		genesisValidators    *types.StandardValidatorsResponse
		mockSaveError        error
		mockUpdateSigError   error
		mockUpdateCountError error
		expectedError        bool
	}{
		{
			name: "valid genesis validators export",
			genesisValidators: &types.StandardValidatorsResponse{
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
			},
			mockSaveError:        nil,
			mockUpdateSigError:   nil,
			mockUpdateCountError: nil,
			expectedError:        false,
		},
		{
			name: "SaveBlockDeposits error",
			genesisValidators: &types.StandardValidatorsResponse{
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
			},
			mockSaveError:        errors.New("database error"),
			mockUpdateSigError:   nil,
			mockUpdateCountError: nil,
			expectedError:        true,
		},
		{
			name: "UpdateBlockDepositsSignature error",
			genesisValidators: &types.StandardValidatorsResponse{
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
			},
			mockSaveError:        nil,
			mockUpdateSigError:   errors.New("database error"),
			mockUpdateCountError: nil,
			expectedError:        true,
		},
		{
			name: "UpdateBlockDepositCount error",
			genesisValidators: &types.StandardValidatorsResponse{
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
			},
			mockSaveError:        nil,
			mockUpdateSigError:   nil,
			mockUpdateCountError: errors.New("database error"),
			expectedError:        true,
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

			// mock SaveBlockDeposits query
			saveDepositsQuery := `INSERT INTO blocks_deposits \(block_slot, block_root, block_index, publickey, withdrawalcredentials, amount, signature\) VALUES \(0, '\\x01', \$1, \$2, \$3, \$4, \$5\) ON CONFLICT DO NOTHING`
			if tt.mockSaveError == nil {
				for _, validator := range tt.genesisValidators.Data {
					mock.ExpectBegin()
					mock.ExpectExec(saveDepositsQuery).
						WithArgs(validator.Index, validator.Validator.Pubkey, validator.Validator.WithdrawalCredentials, validator.Balance, []byte{0x0}).
						WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectCommit()
				}
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(saveDepositsQuery).
					WithArgs(tt.genesisValidators.Data[0].Index, tt.genesisValidators.Data[0].Validator.Pubkey, tt.genesisValidators.Data[0].Validator.WithdrawalCredentials, tt.genesisValidators.Data[0].Balance, []byte{0x0}).
					WillReturnError(tt.mockSaveError)
				mock.ExpectRollback()
			}

			// mock UpdateBlockDepositsSignature query
			updateDepositsQuery := `UPDATE blocks_deposits
					SET signature = a\.signature
					FROM \(
						SELECT DISTINCT ON\(publickey\) publickey, signature
						FROM eth1_deposits
						WHERE valid_signature = true\) AS a
					WHERE block_slot = 0 AND blocks_deposits\.publickey = a\.publickey AND blocks_deposits\.signature = '\\x'`
			if tt.mockUpdateSigError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(updateDepositsQuery).
					WillReturnError(tt.mockUpdateSigError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(updateDepositsQuery).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			// mock UpdateBlockDepositCount query
			updateDepositCountQuery := `UPDATE blocks SET depositscount = \$1 WHERE slot = 0`
			if tt.mockUpdateCountError != nil {
				mock.ExpectBegin()
				mock.ExpectExec(updateDepositCountQuery).
					WithArgs(len(tt.genesisValidators.Data)).
					WillReturnError(tt.mockUpdateCountError)
				mock.ExpectRollback()
			} else {
				mock.ExpectBegin()
				mock.ExpectExec(updateDepositCountQuery).
					WithArgs(len(tt.genesisValidators.Data)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err = exportGenesisDeposits(tt.genesisValidators)

			if !tt.expectedError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error: %v, got nil", tt.expectedError)
				}
			}
		})
	}
}
