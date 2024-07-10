package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

func (d *DataAccessService) GetValidatorDashboardRocketPool(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBRocketPoolColumn], search string, limit uint64) ([]t.VDBRocketPoolTableRow, *t.Paging, error) {
	// TODO @DATA-ACCESS
	return d.dummy.GetValidatorDashboardRocketPool(ctx, dashboardId, cursor, colSort, search, limit)
}

func (d *DataAccessService) GetValidatorDashboardTotalRocketPool(ctx context.Context, dashboardId t.VDBId, search string) (*t.VDBTotalRocketPoolData, error) {
	// TODO @DATA-ACCESS
	return d.dummy.GetValidatorDashboardTotalRocketPool(ctx, dashboardId, search)
}

func (d *DataAccessService) GetValidatorDashboardNodeRocketPool(ctx context.Context, dashboardId t.VDBId, node string) (*t.VDBNodeRocketPoolData, error) {
	// TODO @DATA-ACCESS
	return d.dummy.GetValidatorDashboardNodeRocketPool(ctx, dashboardId, node)
}
