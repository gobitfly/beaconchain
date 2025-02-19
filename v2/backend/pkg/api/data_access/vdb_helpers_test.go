package dataaccess

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func setupConfig() {
	cfg := types.Config{
		Chain: types.Chain{
			ClConfig: types.ClChainConfig{
				EpochsPerSyncCommitteePeriod: 32,
				AltairForkEpoch:              0,
			},
		},
	}
	utils.Config = &cfg
}

func setupTestDataAccess(t *testing.T) (da *DataAccessService, mock sqlmock.Sqlmock) {
	setupConfig()

	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	dataAccess := &DataAccessService{
		clickhouseReader: sqlxDB,
		readerDb:         sqlxDB,
		alloyReader:      sqlxDB,
	}

	return dataAccess, mock
}

func assertTestError(t *testing.T, actual error, expected string) {
	if expected == "" {
		assert.NoError(t, actual)
	} else {
		assert.EqualError(t, actual, expected)
	}
}

func TestGetCurrentAndUpcomingSyncCommittees(t *testing.T) {
	dataAccessService, mock := setupTestDataAccess(t)
	defer dataAccessService.Close()

	latestEpoch := uint64(128)
	currentSyncPeriod := utils.SyncPeriodOfEpoch(latestEpoch)

	tests := []struct {
		name             string
		mockRows         *sqlmock.Rows
		mockError        error
		expectedCurrent  map[uint64]bool
		expectedUpcoming map[uint64]bool
		expectedError    string
	}{
		{
			name: "Valid data",
			mockRows: sqlmock.NewRows([]string{"validatorindex", "period"}).
				AddRow(1, currentSyncPeriod).
				AddRow(2, currentSyncPeriod).
				AddRow(3, currentSyncPeriod+1).
				AddRow(4, currentSyncPeriod+1),
			mockError:        nil,
			expectedCurrent:  map[uint64]bool{1: true, 2: true},
			expectedUpcoming: map[uint64]bool{3: true, 4: true},
			expectedError:    "",
		},
		{
			name:             "Database error",
			mockRows:         nil,
			mockError:        errors.New("db query failed"),
			expectedCurrent:  nil,
			expectedUpcoming: nil,
			expectedError:    "error executing query: db query failed",
		},
		{
			name:             "Empty result",
			mockRows:         sqlmock.NewRows([]string{"validatorindex", "period"}),
			mockError:        nil,
			expectedCurrent:  map[uint64]bool{},
			expectedUpcoming: map[uint64]bool{},
			expectedError:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `SELECT validatorindex, period FROM "sync_committees" WHERE period IN \(\$1, \$2\)`
			if tt.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(query).WillReturnRows(tt.mockRows)
			}

			ctx := context.Background()
			current, upcoming, err := dataAccessService.getCurrentAndUpcomingSyncCommittees(ctx, latestEpoch)

			assertTestError(t, err, tt.expectedError)
			assert.Equal(t, tt.expectedCurrent, current)
			assert.Equal(t, tt.expectedUpcoming, upcoming)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetPastSyncCommittees(t *testing.T) {
	dataAccessService, mock := setupTestDataAccess(t)
	defer dataAccessService.Close()

	tests := []struct {
		name          string
		indicies      []uint64
		epochStart    uint64
		latestEpoch   uint64
		mockRows      *sqlmock.Rows
		mockError     error
		expectedMap   map[uint64]uint64
		expectedError string
	}{
		{
			name:        "Valid data",
			indicies:    []uint64{1, 2, 3},
			epochStart:  100,
			latestEpoch: 200,
			mockRows: sqlmock.NewRows([]string{"validatorindex"}).
				AddRow(1).
				AddRow(2).
				AddRow(2),
			mockError:     nil,
			expectedMap:   map[uint64]uint64{1: 1, 2: 2},
			expectedError: "",
		},
		{
			name:          "Query error",
			indicies:      []uint64{1, 2, 3},
			epochStart:    100,
			latestEpoch:   200,
			mockRows:      nil,
			mockError:     errors.New("db query failed"),
			expectedMap:   nil,
			expectedError: "error executing query: db query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := "SELECT sc.validatorindex FROM sync_committees sc WHERE period >= $1 AND period < $2 AND validatorindex = ANY($3)"

			if tt.mockError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(utils.SyncPeriodOfEpoch(tt.epochStart), utils.SyncPeriodOfEpoch(tt.latestEpoch), pq.Array(tt.indicies)).
					WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(utils.SyncPeriodOfEpoch(tt.epochStart), utils.SyncPeriodOfEpoch(tt.latestEpoch), pq.Array(tt.indicies)).
					WillReturnRows(tt.mockRows)
			}

			ctx := context.Background()
			result, err := dataAccessService.getPastSyncCommittees(ctx, tt.indicies, tt.epochStart, tt.latestEpoch)

			assertTestError(t, err, tt.expectedError)
			assert.Equal(t, tt.expectedMap, result)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetLatestExportedChartTs(t *testing.T) {
	dataAccessService, mock := setupTestDataAccess(t)
	defer dataAccessService.Close()

	// Define the test cases
	tests := []struct {
		name          string
		aggregation   enums.ChartAggregation
		mockRows      *sqlmock.Rows
		mockError     error
		expectedTs    uint64
		expectedError string
	}{
		{
			name:        "Valid data",
			aggregation: enums.IntervalDaily,
			mockRows: sqlmock.NewRows([]string{"max"}).
				AddRow(time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)),
			mockError:     nil,
			expectedTs:    uint64(time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC).Unix()),
			expectedError: "",
		},
		{
			name:        "Invalid data type",
			aggregation: enums.IntervalDaily,
			mockRows: sqlmock.NewRows([]string{"max"}).
				AddRow("invalid_timestamp"),
			mockError:     nil,
			expectedTs:    0,
			expectedError: "error executing query: sql: Scan error on column index 0, name \"max\": unsupported Scan, storing driver.Value type string into type *time.Time",
		},
		{
			name:          "Empty result",
			aggregation:   enums.IntervalDaily,
			mockRows:      sqlmock.NewRows([]string{"max"}),
			mockError:     nil,
			expectedTs:    0,
			expectedError: "error executing query: sql: no rows in result set",
		},
		{
			name:          "Query error",
			aggregation:   enums.IntervalDaily,
			mockRows:      nil,
			mockError:     fmt.Errorf("db query failed"),
			expectedTs:    0,
			expectedError: "error executing query: db query failed",
		},
	}

	// Run the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view, dateColumn, err := getViewAndDateColumn(tt.aggregation)
			assert.NoError(t, err)

			// Mock the query
			query := fmt.Sprintf("SELECT MAX\\(\"%s\"\\) FROM \"%s\"", dateColumn, view)
			if tt.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(query).WillReturnRows(tt.mockRows)
			}

			// Run the function under test
			ctx := context.Background()
			ts, err := dataAccessService.GetLatestExportedChartTs(ctx, enums.IntervalDaily)

			assertTestError(t, err, tt.expectedError)
			assert.Equal(t, tt.expectedTs, ts)

			// Ensure all expectations are met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetEpochStart(t *testing.T) {
	dataAccessService, mock := setupTestDataAccess(t)
	defer dataAccessService.Close()

	tests := []struct {
		name          string
		period        enums.TimePeriod
		mockRows      *sqlmock.Rows
		mockError     error
		expectedVal   uint64
		expectedError string
	}{
		{
			name:   "Valid data",
			period: enums.Last24h,
			mockRows: sqlmock.NewRows([]string{"epoch_start"}).
				AddRow(12345),
			mockError:     nil,
			expectedVal:   12345,
			expectedError: "",
		},
		{
			name:   "Invalid data type",
			period: enums.Last24h,
			mockRows: sqlmock.NewRows([]string{"epoch_start"}).
				AddRow("invalid_value"),
			mockError:     nil,
			expectedVal:   0,
			expectedError: "error executing query: sql: Scan error on column index 0, name \"epoch_start\": converting driver.Value type string (\"invalid_value\") to a uint64: invalid syntax",
		},
		{
			name:          "Empty result",
			period:        enums.Last24h,
			mockRows:      sqlmock.NewRows([]string{"epoch_start"}),
			mockError:     nil,
			expectedVal:   0,
			expectedError: "error executing query: sql: no rows in result set",
		},
		{
			name:          "Query error",
			period:        enums.Last24h,
			mockRows:      nil,
			mockError:     fmt.Errorf("db query failed"),
			expectedVal:   0,
			expectedError: "error executing query: db query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, _, err := getTablesForPeriod(tt.period)
			assert.NoError(t, err)

			query := fmt.Sprintf("SELECT epoch_start FROM %s FINAL ORDER BY epoch_start ASC LIMIT \\$1", table)
			if tt.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(query).WithArgs(1).WillReturnRows(tt.mockRows)
			}

			ctx := context.Background()
			val, err := dataAccessService.getEpochStart(ctx, tt.period)
			assertTestError(t, err, tt.expectedError)
			assert.Equal(t, tt.expectedVal, val)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

type getTableTestCase[T any, R1 any, R2 any] struct {
	name      string
	input     T
	expected1 R1
	expected2 R2
	expectErr error
}

func runGetTableTest[T any, R1 any, R2 any](t *testing.T, cases []getTableTestCase[T, R1, R2], testFunc func(T) (R1, R2, error)) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result1, result2, err := testFunc(tc.input)

			assert.Equal(t, tc.expected1, result1)
			assert.Equal(t, tc.expected2, result2)
			if tc.expectErr != nil {
				assert.EqualError(t, err, tc.expectErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetTable(t *testing.T) {
	t.Run("GetTablesForPeriod", func(t *testing.T) {
		cases := []getTableTestCase[enums.TimePeriod, string, int]{
			{"Last 1 hour", enums.TimePeriods.Last1h, "validator_dashboard_data_rolling_1h", 1, nil},
			{"Last 24 hours", enums.TimePeriods.Last24h, "validator_dashboard_data_rolling_24h", 24, nil},
			{"Last 7 days", enums.TimePeriods.Last7d, "validator_dashboard_data_rolling_7d", 7 * 24, nil},
			{"Last 30 days", enums.TimePeriods.Last30d, "validator_dashboard_data_rolling_30d", 30 * 24, nil},
			{"All time", enums.TimePeriods.AllTime, "validator_dashboard_data_rolling_total", -1, nil},
			{"Invalid time period", enums.TimePeriod(999), "", 0, fmt.Errorf("not-implemented time period: %v", enums.TimePeriod(999))},
		}
		runGetTableTest(t, cases, getTablesForPeriod)
	})

	t.Run("GetTableAndDateColumn", func(t *testing.T) {
		cases := []getTableTestCase[enums.ChartAggregation, string, string]{
			{"Epoch aggregation", enums.IntervalEpoch, "validator_dashboard_data_epoch", "epoch_timestamp", nil},
			{"Hourly aggregation", enums.IntervalHourly, "validator_dashboard_data_hourly", "t", nil},
			{"Daily aggregation", enums.IntervalDaily, "validator_dashboard_data_daily", "t", nil},
			{"Weekly aggregation", enums.IntervalWeekly, "validator_dashboard_data_weekly", "t", nil},
			{"Invalid aggregation", enums.ChartAggregation(999), "", "", fmt.Errorf("unexpected aggregation type: %v", enums.ChartAggregation(999))},
		}
		runGetTableTest(t, cases, getTableAndDateColumn)
	})

	t.Run("GetViewAndDateColumn", func(t *testing.T) {
		cases := []getTableTestCase[enums.ChartAggregation, string, string]{
			{"Epoch aggregation", enums.IntervalEpoch, "view_validator_dashboard_data_epoch_max_ts", "t", nil},
			{"Hourly aggregation", enums.IntervalHourly, "view_validator_dashboard_data_hourly_max_ts", "t", nil},
			{"Daily aggregation", enums.IntervalDaily, "view_validator_dashboard_data_daily_max_ts", "t", nil},
			{"Weekly aggregation", enums.IntervalWeekly, "view_validator_dashboard_data_weekly_max_ts", "t", nil},
			{"Invalid aggregation", enums.ChartAggregation(999), "", "", fmt.Errorf("unexpected aggregation type: %v", enums.ChartAggregation(999))},
		}
		runGetTableTest(t, cases, getViewAndDateColumn)
	})
}
