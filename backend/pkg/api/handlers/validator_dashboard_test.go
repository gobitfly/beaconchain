package handlers

import (
	"context"
	"testing"
	"time"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/stretchr/testify/assert"
)

// ------------------------------------------------------------

type validatorDashboardDataAccessStub struct {
	dataaccess.DummyService
	overridePremiumPerks *types.PremiumPerks
	overrideGroupCount   *uint64
}

func (d *validatorDashboardDataAccessStub) GetUserInfo(ctx context.Context, id uint64) (*types.UserInfo, error) {
	premiumPerks := types.PremiumPerks{
		ValidatorGroupsPerDashboard: 1,
		ChartHistorySeconds: types.ChartHistorySeconds{
			Epoch:  100,
			Daily:  1000,
			Hourly: 10000,
			Weekly: 100000,
		},
	}

	if d.overridePremiumPerks != nil {
		premiumPerks = *d.overridePremiumPerks
	}

	return &types.UserInfo{
		PremiumPerks: premiumPerks,
	}, nil
}

func (d *validatorDashboardDataAccessStub) GetValidatorDashboardGroupCount(ctx context.Context, dashboardId types.VDBIdPrimary) (uint64, error) {
	var count uint64
	if d.overrideGroupCount != nil {
		count = *d.overrideGroupCount
	}
	return count, nil
}

func (*validatorDashboardDataAccessStub) GetLatestExportedChartTs(ctx context.Context, aggregation enums.ChartAggregation) (uint64, error) {
	return 1000000000, nil // 2001-09-09 01:46:40
}

type stubOption func(*validatorDashboardDataAccessStub)

func withPremiumPerks(perks types.PremiumPerks) stubOption {
	return func(d *validatorDashboardDataAccessStub) {
		d.overridePremiumPerks = &perks
	}
}
func withGroupCount(count uint64) stubOption {
	return func(d *validatorDashboardDataAccessStub) {
		d.overrideGroupCount = &count
	}
}
func validatorDashboardTestSetup(options ...stubOption) (context.Context, *HandlerService) {
	d := &validatorDashboardDataAccessStub{}
	for _, option := range options {
		option(d)
	}
	return handlerTestSetup(d)
}

// ------------------------------------------------------------
// POST /validator-dashboards/{dashboard_id}/groups

func TestInputPostValidatorDashboardGroupsValidate(t *testing.T) {
	params := make(map[string]string)
	t.Run("success", func(t *testing.T) {
		var i inputPostValidatorDashboardGroups
		params["dashboard_id"] = "1"
		body := stringAsBody(`{"name":"test"}`)
		err := i.Validate(params, body)
		assert.NoError(t, err)
		assert.Equal(t, types.VDBIdPrimary(1), i.dashboardId)
		assert.Equal(t, "test", i.name)
	})
	t.Run("empty name", func(t *testing.T) {
		var i inputPostValidatorDashboardGroups
		params["dashboard_id"] = "1"
		body := stringAsBody(`{"name":""}`)
		err := i.Validate(params, body)
		assert.Error(t, err)
	})
}
func TestPostValidatorDashboardGroups(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, h := validatorDashboardTestSetup()
		input := inputPostValidatorDashboardGroups{
			dashboardId: 0,
			name:        "test",
		}
		_, err := h.PostValidatorDashboardGroups(ctx, input)
		assert.NoError(t, err)
	})
	t.Run("group count reached", func(t *testing.T) {
		ctx, h := validatorDashboardTestSetup(withGroupCount(1))
		input := inputPostValidatorDashboardGroups{
			dashboardId: 0,
			name:        "test",
		}
		_, err := h.PostValidatorDashboardGroups(ctx, input)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errConflict)
	})
}

// ------------------------------------------------------------
// GET /validator-dashboards/{dashboard_id}/groups/{group_id}/summary

func TestInputGetValidatorDashboardGroupSummaryValidate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var i inputGetValidatorDashboardGroupSummary
		params := map[string]string{
			"dashboard_id": "1",
			"group_id":     "1",
			"period":       "all_time",
		}
		err := i.Validate(params, nil)
		assert.NoError(t, err)
		assert.Equal(t, types.VDBIdPrimary(1), i.dashboardIdParam)
		assert.Equal(t, int64(1), i.groupId)
	})
	t.Run("empty dashboard_id", func(t *testing.T) {
		var i inputGetValidatorDashboardGroupSummary
		params := map[string]string{
			"dashboard_id": "",
			"group_id":     "1",
			"period":       "all_time",
		}
		err := i.Validate(params, nil)
		assert.Error(t, err)
	})
	t.Run("empty group_id", func(t *testing.T) {
		var i inputGetValidatorDashboardGroupSummary
		params := map[string]string{
			"dashboard_id": "1",
			"group_id":     "",
			"period":       "all_time",
		}
		err := i.Validate(params, nil)
		assert.Error(t, err)
	})
	t.Run("empty period", func(t *testing.T) {
		var i inputGetValidatorDashboardGroupSummary
		params := map[string]string{
			"dashboard_id": "1",
			"group_id":     "1",
			"period":       "",
		}
		err := i.Validate(params, nil)
		assert.Error(t, err)
	})
}

// ------------------------------------------------------------
// GET /validator-dashboards/{dashboard_id}/summary-chart

func TestInputGetValidatorDashboardSummaryChartValidate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var i inputGetValidatorDashboardSummaryChart
		params := map[string]string{
			"dashboard_id": "1",
		}
		err := i.Validate(params, nil)
		assert.NoError(t, err)
		assert.Equal(t, types.VDBIdPrimary(1), i.dashboardId)
		assert.Nil(t, i.afterTs)
		assert.Nil(t, i.beforeTs)
	})
	t.Run("empty dashboard_id", func(t *testing.T) {
		var i inputGetValidatorDashboardSummaryChart
		params := map[string]string{
			"dashboard_id": "",
		}
		err := i.Validate(params, nil)
		assert.Error(t, err)
	})
	t.Run("only set after ts", func(t *testing.T) {
		var i inputGetValidatorDashboardSummaryChart
		params := map[string]string{
			"dashboard_id": "1",
			"after_ts":     "100",
		}
		err := i.Validate(params, nil)
		assert.NoError(t, err)
		assert.NotNil(t, i.afterTs)
		assert.Nil(t, i.beforeTs)
		assert.Equal(t, uint64(100), *i.afterTs)
	})
	t.Run("only set before ts", func(t *testing.T) {
		var i inputGetValidatorDashboardSummaryChart
		params := map[string]string{
			"dashboard_id": "1",
			"before_ts":    "100",
		}
		err := i.Validate(params, nil)
		assert.NoError(t, err)
		assert.NotNil(t, i.beforeTs)
		assert.Nil(t, i.afterTs)
		assert.Equal(t, uint64(100), *i.beforeTs)
	})
	t.Run("set both ts", func(t *testing.T) {
		var i inputGetValidatorDashboardSummaryChart
		params := map[string]string{
			"dashboard_id": "1",
			"after_ts":     "100",
			"before_ts":    "200",
		}
		err := i.Validate(params, nil)
		assert.NoError(t, err)
		assert.NotNil(t, i.afterTs)
		assert.NotNil(t, i.beforeTs)
		assert.Equal(t, uint64(100), *i.afterTs)
		assert.Equal(t, uint64(200), *i.beforeTs)
	})
	t.Run("after ts >= before ts", func(t *testing.T) {
		var i inputGetValidatorDashboardSummaryChart
		params := map[string]string{
			"dashboard_id": "1",
			"after_ts":     "100",
			"before_ts":    "100",
		}
		err := i.Validate(params, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "after_ts must be less than before_ts")
	})
}

func ptr[T any](v T) *T {
	return &v
}
func TestResolveAndValidateTimestamps_Success(t *testing.T) {
	var chartSeconds uint64 = 1000 // -> min timestamp = lastExportedTs - chartSeconds
	duration := time.Second        // -> max interval = 200s
	tests := []struct {
		name             string
		latestExportedTs uint64
		givenAfterTs     *uint64
		givenBeforeTs    *uint64
		wantAfterTs      uint64
		wantBeforeTs     uint64
	}{
		// no timestams are provided, should resolve to beforeTs = latestExportedTs and afterTs = latestExportedTs - maxAllowedInterval
		{
			name:             "no timestamps",
			latestExportedTs: 1000000000,
			givenAfterTs:     nil,
			givenBeforeTs:    nil,
			wantAfterTs:      999999800,
			wantBeforeTs:     1000000000,
		},
		// no timestamps are provided and latestExportedTs is low, should resolve to beforeTs = latestExportedTs and afterTs = 0
		{
			name:             "no timestamps - low latest ts",
			latestExportedTs: 100,
			givenAfterTs:     nil,
			givenBeforeTs:    nil,
			wantAfterTs:      0,
			wantBeforeTs:     100,
		},
		// afterTs is provided, beforeTs should be afterTs + maxAllowedInterval
		{
			name:             "high after ts",
			latestExportedTs: 1000000000,
			givenAfterTs:     ptr(uint64(1000000000)),
			givenBeforeTs:    nil,
			wantAfterTs:      1000000000,
			wantBeforeTs:     1000000200,
		},
		// afterTs is provided and lowest possible
		{
			name:             "low after ts",
			latestExportedTs: 1000000000,
			givenAfterTs:     ptr(uint64(999999000)),
			givenBeforeTs:    nil,
			wantAfterTs:      999999000,
			wantBeforeTs:     999999200,
		},
		// beforeTs is provided, afterTs should be beforeTs - maxAllowedInterval
		{
			name:             "high before ts",
			latestExportedTs: 1000000000,
			givenAfterTs:     nil,
			givenBeforeTs:    ptr(uint64(999999800)),
			wantAfterTs:      999999600,
			wantBeforeTs:     999999800,
		},
		// beforeTs is exactly minAllowedTs + maxAllowedInterval, afterTs should be minAllowedTs
		{
			name:             "low before ts - exact",
			latestExportedTs: 1000000000,
			givenAfterTs:     nil,
			givenBeforeTs:    ptr(uint64(999999200)),
			wantAfterTs:      999999000,
			wantBeforeTs:     999999200,
		},
		// beforeTs is provided and close to minAllowedTs, afterTs should be minAllowedTs
		{
			name:             "low before ts",
			latestExportedTs: 1000000000,
			givenAfterTs:     nil,
			givenBeforeTs:    ptr(uint64(999999050)),
			wantAfterTs:      999999000,
			wantBeforeTs:     999999050,
		},
		// both timestamps are provided
		{
			name:             "both timestamps",
			latestExportedTs: 1000000000,
			givenAfterTs:     ptr(uint64(999999950)),
			givenBeforeTs:    ptr(uint64(1000000000)),
			wantAfterTs:      999999950,
			wantBeforeTs:     1000000000,
		},
		// both timestamps are provided, high edge case
		{
			name:             "both timestamps - high edge",
			latestExportedTs: 1000000000,
			givenAfterTs:     ptr(uint64(1000000000)),
			givenBeforeTs:    ptr(uint64(1000000200)),
			wantAfterTs:      1000000000,
			wantBeforeTs:     1000000200,
		},
		// both timestamps are provided, low edge case
		{
			name:             "both timestamps - low edge",
			latestExportedTs: 1000000000,
			givenAfterTs:     ptr(uint64(999999000)),
			givenBeforeTs:    ptr(uint64(999999200)),
			wantAfterTs:      999999000,
			wantBeforeTs:     999999200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAfterTs, gotBeforeTs, err := resolveAndValidateTimestamps(tt.givenAfterTs, tt.givenBeforeTs, chartSeconds, duration, tt.latestExportedTs)
			assert.NoError(t, err, "Expected no error, got %v", err)
			assert.Equal(t, tt.wantAfterTs, gotAfterTs, "Expected afterTs to be %d, got %d", tt.wantAfterTs, gotAfterTs)
			assert.Equal(t, tt.wantBeforeTs, gotBeforeTs, "Expected beforeTs to be %d, got %d", tt.wantBeforeTs, gotBeforeTs)
		})
	}
}

func TestResolveAndValidateTimestamps_Failure(t *testing.T) {
	var chartSeconds uint64 = 1000
	duration := time.Second // -> max interval = 200s
	tests := []struct {
		name             string
		latestExportedTs uint64
		givenAfterTs     *uint64
		givenBeforeTs    *uint64
		errMsg           string
	}{
		{
			name:             "after ts below min allowed",
			latestExportedTs: 1000000000,
			givenAfterTs:     ptr(uint64(999998999)),
			givenBeforeTs:    nil,
			errMsg:           "`after_ts` must be greater or equal to 999999000",
		},
		{
			name:             "before ts below min allowed",
			latestExportedTs: 1000000000,
			givenAfterTs:     nil,
			givenBeforeTs:    ptr(uint64(999998999)),
			errMsg:           "`before_ts` must be greater or equal to 999999000",
		},
		{
			name:             "both timestamps - too high interval",
			latestExportedTs: 1000000000,
			givenAfterTs:     ptr(uint64(999999000)),
			givenBeforeTs:    ptr(uint64(999999201)),
			errMsg:           "difference between `before_ts` and `after_ts` must be smaller or equal to 200",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := resolveAndValidateTimestamps(tt.givenAfterTs, tt.givenBeforeTs, chartSeconds, duration, tt.latestExportedTs)
			assert.Error(t, err, "Expected error, got %v", err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestGetValidatorDashboardSummaryChart_Success(t *testing.T) {
	perks := types.PremiumPerks{
		ChartHistorySeconds: types.ChartHistorySeconds{
			Epoch: 100,
		},
	}
	ctx, h := validatorDashboardTestSetup(withPremiumPerks(perks))
	input := inputGetValidatorDashboardSummaryChart{
		dashboardId: types.VDBIdPrimary(1),
		aggregation: enums.ChartAggregations.Epoch,
	}
	_, err := h.GetValidatorDashboardSummaryChart(ctx, input)
	assert.NoError(t, err)
}

func TestGetValidatorDashboardSummaryChart_Failure(t *testing.T) {
	aggregations := enums.ChartAggregations
	tests := []struct {
		name         string
		chartSeconds types.ChartHistorySeconds
		aggregation  enums.ChartAggregation
	}{
		{
			name: "epoch",
			chartSeconds: types.ChartHistorySeconds{
				Epoch: 0,
			},
			aggregation: aggregations.Epoch,
		},
		{
			name: "hourly",
			chartSeconds: types.ChartHistorySeconds{
				Hourly: 0,
			},
			aggregation: aggregations.Hourly,
		},
		{
			name: "daily",
			chartSeconds: types.ChartHistorySeconds{
				Daily: 0,
			},
			aggregation: aggregations.Daily,
		},
		{
			name: "weekly",
			chartSeconds: types.ChartHistorySeconds{
				Weekly: 0,
			},
			aggregation: aggregations.Weekly,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perks := types.PremiumPerks{
				ChartHistorySeconds: tt.chartSeconds,
			}
			ctx, h := validatorDashboardTestSetup(withPremiumPerks(perks))
			input := inputGetValidatorDashboardSummaryChart{
				dashboardId: types.VDBIdPrimary(1),
				aggregation: tt.aggregation,
			}
			_, err := h.GetValidatorDashboardSummaryChart(ctx, input)
			assert.Error(t, err)
			assert.ErrorIs(t, err, errForbidden)
		})
	}
}
