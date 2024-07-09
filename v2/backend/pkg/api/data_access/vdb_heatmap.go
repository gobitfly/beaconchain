package dataaccess

import (
	"context"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

// retrieve data for last hour
func (d *DataAccessService) GetValidatorDashboardEpochHeatmap(ctx context.Context, dashboardId t.VDBId, protocolModes []enums.ProtocolMode) (*t.VDBHeatmap, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardEpochHeatmap(ctx, dashboardId, protocolModes)
}

// allowed periods are: last_7d, last_30d, last_365d
func (d *DataAccessService) GetValidatorDashboardDailyHeatmap(ctx context.Context, dashboardId t.VDBId, period enums.TimePeriod, protocolModes []enums.ProtocolMode) (*t.VDBHeatmap, error) {
	// TODO @remoterami
	return d.dummy.GetValidatorDashboardDailyHeatmap(ctx, dashboardId, period, protocolModes)
}

func (d *DataAccessService) GetValidatorDashboardGroupEpochHeatmap(ctx context.Context, dashboardId t.VDBId, groupId uint64, epoch uint64, protocolModes []enums.ProtocolMode) (*t.VDBHeatmapTooltipData, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardGroupEpochHeatmap(ctx, dashboardId, groupId, epoch, protocolModes)
}

func (d *DataAccessService) GetValidatorDashboardGroupDailyHeatmap(ctx context.Context, dashboardId t.VDBId, groupId uint64, day time.Time, protocolModes []enums.ProtocolMode) (*t.VDBHeatmapTooltipData, error) {
	// TODO @remoterami
	return d.dummy.GetValidatorDashboardGroupDailyHeatmap(ctx, dashboardId, groupId, day, protocolModes)
}
