package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain/pkg/api/types"
)

type ProtocolRepository interface {
	// Rocket Pool
	GetRocketPoolOverview(context.Context) (*types.RocketPoolData, error)

	// Lido, ...
}

func (d *DataAccessService) GetRocketPoolOverview(ctx context.Context) (*types.RocketPoolData, error) {
	return d.dummy.GetRocketPoolOverview(ctx)
}
