package dataaccess

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	api_types "github.com/gobitfly/beaconchain/pkg/api/types"
	common_types "github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func setupConfig() {
	cfg := common_types.Config{
		Chain: common_types.Chain{
			GenesisTimestamp: 1606824000,
			ClConfig: common_types.ClChainConfig{
				EpochsPerSyncCommitteePeriod: 32,
				AltairForkEpoch:              0,
				SecondsPerSlot:               12,
				SlotsPerEpoch:                16,
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
			name: "Success",
			mockRows: sqlmock.NewRows([]string{"validatorindex", "period"}).
				AddRow(1, currentSyncPeriod).
				AddRow(2, currentSyncPeriod).
				AddRow(3, currentSyncPeriod+1).
				AddRow(4, currentSyncPeriod+1),
			expectedCurrent:  map[uint64]bool{1: true, 2: true},
			expectedUpcoming: map[uint64]bool{3: true, 4: true},
		},
		{
			name:          "Query error",
			mockError:     errors.New("db query failed"),
			expectedError: "error executing query: db query failed",
		},
		{
			name:             "Empty rows",
			mockRows:         sqlmock.NewRows([]string{"validatorindex", "period"}),
			expectedCurrent:  map[uint64]bool{},
			expectedUpcoming: map[uint64]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := buildCurrentAndUpcomingSyncCommitteesQuery(latestEpoch)
			query, _, err := ds.Prepared(true).ToSQL()
			assert.NoError(t, err)

			if tt.mockError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(tt.mockRows)
			}

			ctx := context.Background()
			current, upcoming, err := dataAccessService.getCurrentAndUpcomingSyncCommittees(ctx, latestEpoch)

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
			name:        "Success",
			indicies:    []uint64{1, 2, 3},
			epochStart:  100,
			latestEpoch: 200,
			mockRows: sqlmock.NewRows([]string{"validatorindex"}).
				AddRow(1).
				AddRow(2).
				AddRow(2),
			expectedMap: map[uint64]uint64{1: 1, 2: 2},
		},
		{
			name:        "Empty rows",
			mockRows:    sqlmock.NewRows([]string{"validatorindex"}),
			expectedMap: map[uint64]uint64{},
		},
		{
			name:          "Query error",
			indicies:      []uint64{1, 2, 3},
			epochStart:    100,
			latestEpoch:   200,
			mockError:     errors.New("db query failed"),
			expectedError: "error executing query: db query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := buildPastSyncCommitteesQuery(tt.indicies, tt.epochStart, tt.latestEpoch)
			query, _, err := ds.Prepared(true).ToSQL()
			assert.NoError(t, err)

			if tt.mockError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(tt.mockRows)
			}

			ctx := context.Background()
			result, err := dataAccessService.getPastSyncCommittees(ctx, tt.indicies, tt.epochStart, tt.latestEpoch)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}
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
			name:        "Success",
			aggregation: enums.IntervalDaily,
			mockRows: sqlmock.NewRows([]string{"max"}).
				AddRow(time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)),
			expectedTs: uint64(time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC).Unix()),
		},
		{
			name:          "Empty rows",
			aggregation:   enums.IntervalDaily,
			mockRows:      sqlmock.NewRows([]string{"max"}),
			expectedError: "error executing query: sql: no rows in result set",
		},
		{
			name:          "Query error",
			aggregation:   enums.IntervalDaily,
			mockError:     fmt.Errorf("db query failed"),
			expectedError: "error executing query: db query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view, dateColumn, err := getViewAndDateColumn(tt.aggregation)
			assert.NoError(t, err)

			ds := buildLatestExportedChartQuery(view, dateColumn)
			query, _, err := ds.Prepared(true).ToSQL()
			assert.NoError(t, err)

			if tt.mockError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(tt.mockRows)
			}

			ctx := context.Background()
			ts, err := dataAccessService.GetLatestExportedChartTs(ctx, enums.IntervalDaily)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}

			assert.Equal(t, tt.expectedTs, ts)
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
			name:   "Success",
			period: enums.Last24h,
			mockRows: sqlmock.NewRows([]string{"epoch_start"}).
				AddRow(12345),
			expectedVal: 12345,
		},
		{
			name:          "Empty rows",
			period:        enums.Last24h,
			mockRows:      sqlmock.NewRows([]string{"epoch_start"}),
			expectedError: "error executing query: sql: no rows in result set",
		},
		{
			name:          "Query error",
			period:        enums.Last24h,
			mockError:     fmt.Errorf("db query failed"),
			expectedError: "error executing query: db query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, _, err := getTablesForPeriod(tt.period)
			assert.NoError(t, err)

			ds := buildEpochStartQuery(table)
			query, _, err := ds.Prepared(true).ToSQL()
			assert.NoError(t, err)

			if tt.mockError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(tt.mockRows)
			}

			ctx := context.Background()
			val, err := dataAccessService.getEpochStart(ctx, tt.period)

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

func TestGetLastScheduledBlockAndSyncDate(t *testing.T) {
	dataAccessService, mock := setupTestDataAccess(t)
	defer dataAccessService.Close()

	dashboardId := api_types.VDBId{Id: 123, Validators: nil}
	groupId := int64(456)
	now := time.Now().Unix() // timestamp

	tests := []struct {
		name                  string
		mockRows              *sqlmock.Rows
		mockError             error
		period                enums.TimePeriod
		expectedLastScheduled time.Time
		expectedLastSync      time.Time
		expectedError         string
	}{
		{
			name: "Success",
			mockRows: sqlmock.NewRows([]string{"last_scheduled_block_epoch", "last_scheduled_sync_epoch"}).
				AddRow(now, now),
			period:                enums.AllTime,
			expectedLastScheduled: utils.EpochToTime(uint64(now)),
			expectedLastSync:      utils.EpochToTime(uint64(now)),
		},
		{
			name:          "Query error",
			period:        enums.AllTime,
			mockError:     errors.New("query failed"),
			expectedError: "error executing query: query failed",
		},
		{
			name:          "Empty rows",
			period:        enums.AllTime,
			mockRows:      sqlmock.NewRows([]string{"last_scheduled_block_epoch", "last_scheduled_sync_epoch"}),
			expectedError: "error executing query: sql: no rows in result set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clickhouseTotalTable, _, err := getTablesForPeriod(tt.period)
			assert.NoError(t, err)

			ds := buildLastScheduledBlockAndSyncDateQuery(clickhouseTotalTable, dashboardId, groupId)
			query, _, err := ds.Prepared(true).ToSQL()
			assert.NoError(t, err)

			if tt.mockError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(tt.mockRows)
			}

			ctx := context.Background()
			lastScheduledTime, lastSyncTime, err := dataAccessService.getLastScheduledBlockAndSyncDate(ctx, dashboardId, groupId)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}
			assert.Equal(t, tt.expectedLastScheduled, lastScheduledTime)
			assert.Equal(t, tt.expectedLastSync, lastSyncTime)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetMinMaxEpochs(t *testing.T) {
	dataAccessService, mock := setupTestDataAccess(t)
	defer dataAccessService.Close()

	dashboardId := api_types.VDBId{Id: 123, Validators: nil}
	groupId := int64(456)

	tests := []struct {
		name        string
		mockRows    *sqlmock.Rows
		mockError   error
		period      enums.TimePeriod
		expectedMin uint64
		expectedMax uint64
		expectedErr string
	}{
		{
			name: "Success",
			mockRows: sqlmock.NewRows([]string{"min_epoch_start", "max_epoch_end"}).
				AddRow(100, 200),
			period:      enums.Last24h,
			expectedMin: 100,
			expectedMax: 200,
		},
		{
			name:        "Query error",
			mockError:   errors.New("query failed"),
			period:      enums.Last24h,
			expectedErr: "error executing query: query failed",
		},
		{
			name:        "Empty rows",
			mockRows:    sqlmock.NewRows([]string{"min_epoch_start", "max_epoch_end"}),
			period:      enums.Last24h,
			expectedErr: "error executing query: sql: no rows in result set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clickhouseTable, _, err := getTablesForPeriod(tt.period)
			assert.NoError(t, err)

			ds := buildMinMaxEpochsQuery(dashboardId, groupId, clickhouseTable)
			query, _, err := ds.Prepared(true).ToSQL()
			assert.NoError(t, err)

			if tt.mockError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(tt.mockRows)
			}

			minEpochRes, maxEpochRes, err := dataAccessService.getMinMaxEpochs(context.Background(), dashboardId, groupId, tt.period)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedErr)
			}

			assert.Equal(t, tt.expectedMin, minEpochRes)
			assert.Equal(t, tt.expectedMax, maxEpochRes)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetMissedELRewards(t *testing.T) {
	dataAccessService, mock := setupTestDataAccess(t)
	defer dataAccessService.Close()

	dashboardId := api_types.VDBId{Id: 123, Validators: nil}
	groupId := int64(456)
	slots := utils.Config.Chain.ClConfig.SlotsPerEpoch / 2
	epochStart := uint64(1000)
	epochEnd := uint64(2000)

	// Mock EL rewards data
	expectedMissedRewards := 1234.56

	tests := []struct {
		name            string
		mockRows        *sqlmock.Rows
		mockError       error
		expectedRewards float64
		expectedError   string
	}{
		{
			name: "Success",
			mockRows: sqlmock.NewRows([]string{"total_missed_rewards"}).
				AddRow(expectedMissedRewards),
			expectedRewards: expectedMissedRewards,
		},
		{
			name:          "Query error",
			mockError:     errors.New("query failed"),
			expectedError: "error executing query: query failed",
		},
		{
			name:          "Empty rows",
			mockRows:      sqlmock.NewRows([]string{"total_missed_rewards"}),
			expectedError: "error executing query: sql: no rows in result set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := buildMissedELRewardsQuery(dashboardId, groupId, slots, epochStart, epochEnd)
			query, _, err := ds.Prepared(true).ToSQL()
			assert.NoError(t, err)

			if tt.mockError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(tt.mockRows)
			}

			missedRewards, err := dataAccessService.getMissedELRewards(context.Background(), dashboardId, groupId, epochStart, epochEnd)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedError)
			}

			assert.Equal(t, tt.expectedRewards, missedRewards)
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
