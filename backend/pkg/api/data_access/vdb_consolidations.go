package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	t "github.com/gobitfly/beaconchain/pkg/api/types"
)

func (d *DataAccessService) GetValidatorDashboardConsolidations(ctx context.Context, dashboardId t.VDBId, cursor string, colSort t.Sort[enums.VDBConsolidationsColumn], search string, limit uint64) ([]t.VDBConsolidationsTableRow, *t.Paging, error) {
	return nil, nil, nil
}
