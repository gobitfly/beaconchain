package modules

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func TestProcessSyncCommitteesCount(t *testing.T) {
	tests := []struct {
		name          string
		mockRows      *sqlmock.Rows
		mockError     error
		mockRows2     *sqlmock.Rows
		mockError2    error
		mockRows3     *sqlmock.Rows
		mockError3    error
		mockRows4     *sqlmock.Rows
		mockError4    error
		expectedError bool
	}{
		{
			name: "valid sync committees count",
			mockRows: sqlmock.NewRows([]string{"period"}).
				AddRow(1).AddRow(2),
			mockError:  nil,
			mockRows2:  sqlmock.NewRows([]string{"epoch"}).AddRow(100).AddRow(101),
			mockError2: nil,
			mockRows3: sqlmock.NewRows([]string{"period"}).
				AddRow(3).
				AddRow(4),
			mockError3: nil,
			mockRows4: sqlmock.NewRows([]string{"count_so_far"}).
				AddRow(2).
				AddRow(4),
			mockError4:    nil,
			expectedError: false,
		},
		{
			name:       "GetSyncCommitteesCountPerValidator error",
			mockRows:   nil,
			mockError:  errors.New("database error"),
			mockRows2:  sqlmock.NewRows([]string{"epoch"}).AddRow(100).AddRow(101),
			mockError2: nil,
			mockRows3: sqlmock.NewRows([]string{"period"}).
				AddRow(3).
				AddRow(4),
			mockError3: nil,
			mockRows4: sqlmock.NewRows([]string{"count_so_far"}).
				AddRow(2).
				AddRow(4),
			mockError4:    nil,
			expectedError: true,
		},
		{
			name: "GetTotalPeriodSyncCommitteesCountPerValidator error",
			mockRows: sqlmock.NewRows([]string{"period"}).
				AddRow(1).AddRow(2),
			mockError:  nil,
			mockRows2:  sqlmock.NewRows([]string{"epoch"}).AddRow(100).AddRow(101),
			mockError2: nil,
			mockRows3:  nil,
			mockError3: errors.New("database error"),
			mockRows4: sqlmock.NewRows([]string{"count_so_far"}).
				AddRow(2).
				AddRow(4),
			mockError4:    nil,
			expectedError: true,
		},
		{
			name: "GetCountSoFarSyncCommitteesCountPerValidator error",
			mockRows: sqlmock.NewRows([]string{"period"}).
				AddRow(1).AddRow(2),
			mockError:  nil,
			mockRows2:  sqlmock.NewRows([]string{"epoch"}).AddRow(100).AddRow(101),
			mockError2: nil,
			mockRows3: sqlmock.NewRows([]string{"period"}).
				AddRow(3).
				AddRow(4),
			mockError3:    nil,
			mockRows4:     nil,
			mockError4:    errors.New("database error"),
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

			utils.Config = &types.Config{
				Chain: types.Chain{
					ClConfig: types.ClChainConfig{
						AltairForkEpoch:              0,
						EpochsPerSyncCommitteePeriod: 32,
					},
				},
			}

			// mock GetSyncCommitteesCountPerValidator query
			query := `SELECT COUNT\(\*\) FROM sync_committees_count_per_validator`
			if tt.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(query).WillReturnRows(tt.mockRows)
			}

			// mock GetLatestFinalizedEpoch query
			query2 := `SELECT epoch FROM epochs WHERE finalized ORDER BY epoch DESC LIMIT 1`
			if tt.mockError2 != nil {
				mock.ExpectQuery(query2).WillReturnError(tt.mockError2)
			} else {
				mock.ExpectQuery(query2).WillReturnRows(tt.mockRows2)
			}

			// mock GetTotalPeriodSyncCommitteesCountPerValidator query
			query3 := `SELECT MAX\(period\) FROM sync_committees_count_per_validator`
			if tt.mockError3 != nil {
				mock.ExpectQuery(query3).WillReturnError(tt.mockError3)
			} else {
				mock.ExpectQuery(query3).WillReturnRows(tt.mockRows3)
			}

			// mock GetCountSoFarSyncCommitteesCountPerValidator query
			query4 := `SELECT count_so_far FROM sync_committees_count_per_validator WHERE period = \$1`
			if tt.mockError4 != nil {
				mock.ExpectQuery(query4).WillReturnError(tt.mockError4)
			} else {
				mock.ExpectQuery(query4).WillReturnRows(tt.mockRows4)
			}

			err = processSyncCommitteesCount()

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

func TestGetEpochPeriodAndCountFromDB(t *testing.T) {
	tests := []struct {
		name                string
		mockRows            *sqlmock.Rows
		mockError           error
		mockRows2           *sqlmock.Rows
		mockError2          error
		expectedFirstPeriod uint64
		expectedCountSoFar  float64
		expectedError       bool
	}{
		{
			name: "valid epoch and period count",
			mockRows: sqlmock.NewRows([]string{"period"}).
				AddRow(3).
				AddRow(4),
			mockError: nil,
			mockRows2: sqlmock.NewRows([]string{"count_so_far"}).
				AddRow(2).
				AddRow(4),
			mockError2:          nil,
			expectedFirstPeriod: 4,
			expectedCountSoFar:  2,
			expectedError:       false,
		},
		{
			name:      "GetTotalPeriodSyncCommitteesCountPerValidator error",
			mockRows:  nil,
			mockError: errors.New("database error"),
			mockRows2: sqlmock.NewRows([]string{"count_so_far"}).
				AddRow(2).
				AddRow(4),
			mockError2:    nil,
			expectedError: true,
		},
		{
			name: "GetCountSoFarSyncCommitteesCountPerValidator error",
			mockRows: sqlmock.NewRows([]string{"period"}).
				AddRow(3).
				AddRow(4),
			mockError:     nil,
			mockRows2:     nil,
			mockError2:    errors.New("database error"),
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

			utils.Config = &types.Config{
				Chain: types.Chain{
					ClConfig: types.ClChainConfig{
						AltairForkEpoch:              0,
						EpochsPerSyncCommitteePeriod: 32,
					},
				},
			}

			// mock GetTotalPeriodSyncCommitteesCountPerValidator query
			query := `SELECT MAX\(period\) FROM sync_committees_count_per_validator`
			if tt.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(query).WillReturnRows(tt.mockRows)
			}

			// mock GetCountSoFarSyncCommitteesCountPerValidator query
			query2 := `SELECT count_so_far FROM sync_committees_count_per_validator WHERE period = \$1`
			if tt.mockError2 != nil {
				mock.ExpectQuery(query2).WillReturnError(tt.mockError2)
			} else {
				mock.ExpectQuery(query2).WillReturnRows(tt.mockRows2)
			}

			firstPeriod, countSoFar, err := getEpochPeriodAndCountFromDB(1, 1)

			if !tt.expectedError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if firstPeriod != tt.expectedFirstPeriod {
					t.Errorf("expected first period: %v, got: %v", tt.expectedFirstPeriod, firstPeriod)
				}
				if countSoFar != tt.expectedCountSoFar {
					t.Errorf("expected count so far: %v, got: %v", tt.expectedCountSoFar, countSoFar)
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

func TestExportSyncCommitteesCount(t *testing.T) {
	tests := []struct {
		name          string
		firstPeriod   uint64
		currentPeriod uint64
		countSoFar    float64
		mockRows      *sqlmock.Rows
		mockError     error
		mockSaveError error
		expectedError bool
	}{
		{
			name:          "valid export sync committees count",
			firstPeriod:   1,
			currentPeriod: 1,
			countSoFar:    10.0,
			mockRows:      sqlmock.NewRows([]string{"validatorscount"}).AddRow(1000),
			mockError:     nil,
			mockSaveError: nil,
			expectedError: false,
		},
		{
			name:          "calculateCountForPeriod error",
			firstPeriod:   1,
			currentPeriod: 1,
			countSoFar:    10.0,
			mockRows:      nil,
			mockError:     errors.New("database error"),
			mockSaveError: nil,
			expectedError: true,
		},
		{
			name:          "SaveSyncCommitteesCount error",
			firstPeriod:   1,
			currentPeriod: 1,
			countSoFar:    10.0,
			mockRows:      sqlmock.NewRows([]string{"validatorscount"}).AddRow(1000),
			mockError:     nil,
			mockSaveError: errors.New("database error"),
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

			utils.Config = &types.Config{
				Chain: types.Chain{
					ClConfig: types.ClChainConfig{
						SyncCommitteeSize: 100,
					},
				},
			}

			// mock GetEpochValidatorsCount query
			query := `SELECT validatorscount FROM epochs WHERE epoch = \$1`
			if tt.mockError != nil {
				mock.ExpectQuery(query).WithArgs(utils.FirstEpochOfSyncPeriod(tt.firstPeriod - 1)).WillReturnError(tt.mockError)
			} else if tt.firstPeriod != 0 {
				mock.ExpectQuery(query).WithArgs(utils.FirstEpochOfSyncPeriod(tt.firstPeriod - 1)).WillReturnRows(tt.mockRows)
			}

			if tt.mockError == nil {
				mock.ExpectBegin()

				// mock SaveSyncCommitteesCount query
				saveQuery := `INSERT INTO sync_committees_count_per_validator \(period, count_so_far\) VALUES \(\d+, \d+\.\d+\) ON CONFLICT \(period\) DO UPDATE SET period = excluded.period, count_so_far = excluded.count_so_far`
				if tt.mockSaveError != nil {
					mock.ExpectExec(saveQuery).WillReturnError(tt.mockSaveError)
					mock.ExpectRollback()
				} else {
					mock.ExpectExec(saveQuery).WillReturnResult(sqlmock.NewResult(1, 1))
					mock.ExpectCommit()
				}
			}

			err = exportSyncCommitteesCount(tt.firstPeriod, tt.currentPeriod, tt.countSoFar)

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

func TestCalculateCountForPeriod(t *testing.T) {
	tests := []struct {
		name          string
		period        uint64
		countSoFar    float64
		mockRows      *sqlmock.Rows
		mockError     error
		expectedCount float64
		expectedError bool
	}{
		{
			name:          "valid period and validators count",
			period:        1,
			countSoFar:    10.0,
			mockRows:      sqlmock.NewRows([]string{"validatorscount"}).AddRow(1000),
			mockError:     nil,
			expectedCount: 10.1,
			expectedError: false,
		},
		{
			name:          "GetEpochValidatorsCount error",
			period:        1,
			countSoFar:    10.0,
			mockRows:      nil,
			mockError:     errors.New("database error"),
			expectedCount: 0,
			expectedError: true,
		},
		{
			name:          "period == 0",
			period:        0,
			countSoFar:    10.0,
			mockRows:      nil,
			mockError:     nil,
			expectedCount: 0.0,
			expectedError: false,
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

			utils.Config = &types.Config{
				Chain: types.Chain{
					ClConfig: types.ClChainConfig{
						SyncCommitteeSize: 100,
					},
				},
			}

			// mock GetEpochValidatorsCount query
			query := `SELECT validatorscount FROM epochs WHERE epoch = \$1`
			if tt.mockError != nil {
				mock.ExpectQuery(query).WithArgs(utils.FirstEpochOfSyncPeriod(tt.period - 1)).WillReturnError(tt.mockError)
			} else if tt.period != 0 {
				mock.ExpectQuery(query).WithArgs(utils.FirstEpochOfSyncPeriod(tt.period - 1)).WillReturnRows(tt.mockRows)
			}

			count, err := calculateCountForPeriod(tt.period, tt.countSoFar)

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

			if count != tt.expectedCount {
				t.Errorf("expected count: %v, got: %v", tt.expectedCount, count)
			}
		})
	}
}
