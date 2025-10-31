package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain/pkg/api/types"
)

type RatelimitRepository interface {
	GetApiWeights(ctx context.Context) ([]types.ApiWeightItem, error)
	// TODO @patrick: move queries from commons/ratelimit/ratelimit.go to here
}

func (d *DataAccessService) GetApiWeights(ctx context.Context) ([]types.ApiWeightItem, error) {
	var result []types.ApiWeightItem
	err := d.userReader.SelectContext(ctx, &result, `
		SELECT bucket, endpoint, method, weight
		FROM api_weights
		WHERE valid_from <= NOW()
	`)
	return result, err
}
