package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain/pkg/api/types"
)

type MachineRepository interface {
	GetUserMachineMetrics(context context.Context, userID uint64, limit uint64, offset uint64) (*types.MachineMetricsData, error)
}

func (d *DataAccessService) GetUserMachineMetrics(ctx context.Context, userID uint64, limit uint64, offset uint64) (*types.MachineMetricsData, error) {
	return d.dummy.GetUserMachineMetrics(ctx, userID, limit, offset)
}
