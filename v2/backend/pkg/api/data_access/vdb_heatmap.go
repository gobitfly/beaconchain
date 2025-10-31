package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

// retrieve data for last hour
func (d *DataAccessService) GetValidatorDashboardHeatmap(ctx context.Context, dashboardId t.VDBId, protocolModes t.VDBProtocolModes, aggregation enums.ChartAggregation, afterTs uint64, beforeTs uint64) (*t.VDBHeatmap, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardHeatmap(ctx, dashboardId, protocolModes, aggregation, afterTs, beforeTs)
}

func (d *DataAccessService) GetValidatorDashboardGroupHeatmap(ctx context.Context, dashboardId t.VDBId, groupId uint64, protocolModes t.VDBProtocolModes, aggregation enums.ChartAggregation, timestamp uint64) (*t.VDBHeatmapTooltipData, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardGroupHeatmap(ctx, dashboardId, groupId, protocolModes, aggregation, timestamp)
}
