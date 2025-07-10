package app

import (
	"context"

	model "github.com/gobitfly/beaconchain-api/api/gen/api_service/v1"
)

func (service *ApiService) ExecutionBlock(ctx context.Context, in *model.ExecutionBlockRequest) (*model.ExecutionBlockResponse, error) {
	data := model.BlockSummary{BlockNumber: "123", BlockHash: "fab"}
	return &model.ExecutionBlockResponse{Status: "ok", Data: &data}, nil
}
