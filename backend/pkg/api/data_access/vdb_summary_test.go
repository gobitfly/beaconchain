package dataaccess

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGetCurrentAndUpcomingSyncCommittees(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	// Wrap the mock DB with sqlx
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	cfg := types.Config{
		Chain: types.Chain{
			ClConfig: types.ClChainConfig{
				EpochsPerSyncCommitteePeriod: 32,
				AltairForkEpoch:              0,
			},
		},
	}
	utils.Config = &cfg

	// Create the DataAccessService with the mock DB
	dataService := &DataAccessService{
		readerDb: sqlxDB,
	}

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
			mockError: nil,
			expectedCurrent: map[uint64]bool{
				1: true,
				2: true,
			},
			expectedUpcoming: map[uint64]bool{
				3: true,
				4: true,
			},
			expectedError: "",
		},
		{
			name:             "Database error",
			mockRows:         nil,
			mockError:        errors.New("db query failed"),
			expectedCurrent:  nil,
			expectedUpcoming: nil,
			expectedError:    "error retrieving sync committee current and next period data: db query failed",
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
			current, upcoming, err := dataService.getCurrentAndUpcomingSyncCommittees(ctx, latestEpoch)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}

			assert.Equal(t, tt.expectedCurrent, current)
			assert.Equal(t, tt.expectedUpcoming, upcoming)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetLatestExportedChartTs(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	dataService := &DataAccessService{
		clickhouseReader: sqlxDB,
	}

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
			expectedError: "error retrieving latest exported chart timestamp: sql: Scan error on column index 0, name \"max\": unsupported Scan, storing driver.Value type string into type *time.Time",
		},
		{
			name:          "Empty result",
			aggregation:   enums.IntervalDaily,
			mockRows:      sqlmock.NewRows([]string{"max"}),
			mockError:     nil,
			expectedTs:    0,
			expectedError: "error retrieving latest exported chart timestamp: sql: no rows in result set",
		},
		{
			name:          "Query error",
			aggregation:   enums.IntervalDaily,
			mockRows:      nil,
			mockError:     fmt.Errorf("db query failed"),
			expectedTs:    0,
			expectedError: "error retrieving latest exported chart timestamp: db query failed",
		},
	}

	// Run the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view, dateColumn, err := dataService.getViewAndDateColumn(tt.aggregation)
			assert.NoError(t, err)

			// Mock the query
			query := fmt.Sprintf("SELECT max\\(%s\\) FROM %s", dateColumn, view)
			if tt.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(query).WillReturnRows(tt.mockRows)
			}

			// Run the function under test
			ctx := context.Background()
			ts, err := dataService.GetLatestExportedChartTs(ctx, enums.IntervalDaily)

			// Validate the result
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}
			assert.Equal(t, tt.expectedTs, ts)

			// Ensure all expectations are met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetEpochStart(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	dataService := &DataAccessService{
		clickhouseReader: sqlxDB,
	}

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
			expectedError: "error retrieving cutoff epoch for past sync committees: sql: Scan error on column index 0, name \"epoch_start\": converting driver.Value type string (\"invalid_value\") to a uint64: invalid syntax",
		},
		{
			name:          "Empty result",
			period:        enums.Last24h,
			mockRows:      sqlmock.NewRows([]string{"epoch_start"}),
			mockError:     nil,
			expectedVal:   0,
			expectedError: "error retrieving cutoff epoch for past sync committees: sql: no rows in result set",
		},
		{
			name:          "Query error",
			period:        enums.Last24h,
			mockRows:      nil,
			mockError:     fmt.Errorf("db query failed"),
			expectedVal:   0,
			expectedError: "error retrieving cutoff epoch for past sync committees: db query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, _, err := dataService.getTablesForPeriod(tt.period)
			assert.NoError(t, err)

			query := fmt.Sprintf("SELECT epoch_start FROM %s FINAL ORDER BY epoch_start ASC LIMIT \\$1", table)
			if tt.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(query).WithArgs(1).WillReturnRows(tt.mockRows)
			}

			ctx := context.Background()
			val, err := dataService.getEpochStart(ctx, tt.period)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}
			assert.Equal(t, tt.expectedVal, val)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetTablesForPeriod(t *testing.T) {
	d := &DataAccessService{}

	tests := []struct {
		name          string
		input         enums.TimePeriod
		expectedTable string
		expectedHours int
		expectedError error
	}{
		{
			name:          "Last 1 hour",
			input:         enums.TimePeriods.Last1h,
			expectedTable: "validator_dashboard_data_rolling_1h",
			expectedHours: 1,
			expectedError: nil,
		},
		{
			name:          "Last 24 hours",
			input:         enums.TimePeriods.Last24h,
			expectedTable: "validator_dashboard_data_rolling_24h",
			expectedHours: 24,
			expectedError: nil,
		},
		{
			name:          "Last 7 days",
			input:         enums.TimePeriods.Last7d,
			expectedTable: "validator_dashboard_data_rolling_7d",
			expectedHours: 7 * 24,
			expectedError: nil,
		},
		{
			name:          "Last 30 days",
			input:         enums.TimePeriods.Last30d,
			expectedTable: "validator_dashboard_data_rolling_30d",
			expectedHours: 30 * 24,
			expectedError: nil,
		},
		{
			name:          "All time",
			input:         enums.TimePeriods.AllTime,
			expectedTable: "validator_dashboard_data_rolling_total",
			expectedHours: -1,
			expectedError: nil,
		},
		{
			name:          "Invalid time period",
			input:         enums.TimePeriod(999),
			expectedTable: "",
			expectedHours: 0,
			expectedError: fmt.Errorf("not-implemented time period: %v", enums.TimePeriod(999)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, hours, err := d.getTablesForPeriod(tt.input)

			assert.Equal(t, tt.expectedTable, table)
			assert.Equal(t, tt.expectedHours, hours)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetTableAndDateColumn(t *testing.T) {
	tests := []struct {
		name               string
		aggregation        enums.ChartAggregation
		expectedTable      string
		expectedDateColumn string
		expectError        error
	}{
		{
			name:               "Epoch aggregation",
			aggregation:        enums.IntervalEpoch,
			expectedTable:      "validator_dashboard_data_epoch",
			expectedDateColumn: "epoch_timestamp",
			expectError:        nil,
		},
		{
			name:               "Hourly aggregation",
			aggregation:        enums.IntervalHourly,
			expectedTable:      "validator_dashboard_data_hourly",
			expectedDateColumn: "t",
			expectError:        nil,
		},
		{
			name:               "Daily aggregation",
			aggregation:        enums.IntervalDaily,
			expectedTable:      "validator_dashboard_data_daily",
			expectedDateColumn: "t",
			expectError:        nil,
		},
		{
			name:               "Weekly aggregation",
			aggregation:        enums.IntervalWeekly,
			expectedTable:      "validator_dashboard_data_weekly",
			expectedDateColumn: "t",
			expectError:        nil,
		},
		{
			name:               "Invalid aggregation",
			aggregation:        enums.ChartAggregation(999), // An invalid aggregation value
			expectedTable:      "",
			expectedDateColumn: "",
			expectError:        fmt.Errorf("unexpected aggregation type: %v", enums.ChartAggregation(999)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DataAccessService{}
			table, dateColumn, err := d.getTableAndDateColumn(tt.aggregation)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unexpected aggregation type")
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedTable, table)
			assert.Equal(t, tt.expectedDateColumn, dateColumn)
		})
	}
}

func TestGetViewAndDateColumn(t *testing.T) {
	tests := []struct {
		name           string
		aggregation    enums.ChartAggregation
		expectedView   string
		expectedColumn string
		expectedError  error
	}{
		{
			name:           "Epoch aggregation",
			aggregation:    enums.IntervalEpoch,
			expectedView:   "view_validator_dashboard_data_epoch_max_ts",
			expectedColumn: "t",
			expectedError:  nil,
		},
		{
			name:           "Hourly aggregation",
			aggregation:    enums.IntervalHourly,
			expectedView:   "view_validator_dashboard_data_hourly_max_ts",
			expectedColumn: "t",
			expectedError:  nil,
		},
		{
			name:           "Daily aggregation",
			aggregation:    enums.IntervalDaily,
			expectedView:   "view_validator_dashboard_data_daily_max_ts",
			expectedColumn: "t",
			expectedError:  nil,
		},
		{
			name:           "Weekly aggregation",
			aggregation:    enums.IntervalWeekly,
			expectedView:   "view_validator_dashboard_data_weekly_max_ts",
			expectedColumn: "t",
			expectedError:  nil,
		},
		{
			name:           "Invalid aggregation",
			aggregation:    enums.ChartAggregation(999),
			expectedView:   "",
			expectedColumn: "",
			expectedError:  fmt.Errorf("unexpected aggregation type: %v", enums.ChartAggregation(999)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DataAccessService{}
			view, column, err := d.getViewAndDateColumn(tt.aggregation)

			assert.Equal(t, tt.expectedView, view)
			assert.Equal(t, tt.expectedColumn, column)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
