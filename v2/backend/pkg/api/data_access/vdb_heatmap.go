package dataaccess

import (
	"context"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

// retrieve data for last hour
func (d *DataAccessService) GetValidatorDashboardEpochHeatmap(ctx context.Context, dashboardId t.VDBId, poolMode bool) (*t.VDBHeatmap, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardEpochHeatmap(ctx, dashboardId, poolMode)
}

// allowed periods are: last_7d, last_30d, last_365d
func (d *DataAccessService) GetValidatorDashboardDailyHeatmap(ctx context.Context, dashboardId t.VDBId, period enums.TimePeriod, poolMode bool) (*t.VDBHeatmap, error) {
	// TODO @remoterami
	return d.dummy.GetValidatorDashboardDailyHeatmap(ctx, dashboardId, period, poolMode)
}

func (d *DataAccessService) GetValidatorDashboardGroupEpochHeatmap(ctx context.Context, dashboardId t.VDBId, groupId uint64, epoch uint64, poolMode bool) (*t.VDBHeatmapTooltipData, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardGroupEpochHeatmap(ctx, dashboardId, groupId, epoch, poolMode)
}

func (d *DataAccessService) GetValidatorDashboardGroupDailyHeatmap(ctx context.Context, dashboardId t.VDBId, groupId uint64, day time.Time, poolMode bool) (*t.VDBHeatmapTooltipData, error) {
	// TODO @remoterami
	return d.dummy.GetValidatorDashboardGroupDailyHeatmap(ctx, dashboardId, groupId, day, poolMode)
}
