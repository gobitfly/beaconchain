package app

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
)

// (GET /ping)
func (ApiService) GetPing(ctx context.Context, request model.GetPingRequestObject) (model.GetPingResponseObject, error) {
	return model.GetPing200JSONResponse(model.Pong{}), nil
}
