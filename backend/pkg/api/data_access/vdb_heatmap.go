package dataaccess

import t "github.com/gobitfly/beaconchain/pkg/api/types"

func (d *DataAccessService) GetValidatorDashboardHeatmap(dashboardId t.VDBId) (*t.VDBHeatmap, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardHeatmap(dashboardId)
}

func (d *DataAccessService) GetValidatorDashboardGroupHeatmap(dashboardId t.VDBId, groupId uint64, epoch uint64) (*t.VDBHeatmapTooltipData, error) {
	// WORKING Rami
	return d.dummy.GetValidatorDashboardGroupHeatmap(dashboardId, groupId, epoch)
}
