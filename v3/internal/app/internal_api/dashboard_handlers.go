package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	dataaccess "github.com/gobitfly/beaconchain-backend/internal/dataaccess/repo"
)

func (service *ApiService) GetValidatorDashboard(ctx context.Context, in *model.GetValidatorDashboardRequest) (*model.GetValidatorDashboardResponse, error) {
	if len(in.DashboardId) == 0 {
		return nil, errors.New("dashboard identifier must be specified")
	}

	// If dashboardId is specified, then return a specific dashboard for the user who sent the request. Ensure that the requestor only ever sees their own dashboards.
	dashboardId, err := strconv.ParseUint(in.DashboardId, 10, 64)
	if err != nil {
		return nil, errors.New("invalid dashboard identifier")
	}

	// Get the user
	_, err = service.userRepository.GetUserById(ctx, 1234 /*TODO: Get userId from request headers via session / JWT Token*/)
	if err != nil {
		return nil, errors.New("internal user id")
	}

	// In a proper implementation we would also want to make sure that this dashboard is actually owned by the requesting user.
	dbDashboard, err := service.dashboardRepository.GetValidatorDashboardByDashboardId(ctx, dashboardId)
	if err != nil {
		return nil, err
	}
	// Request was valid and handled properly by the service, but nothing was found
	if dbDashboard == nil {
		return nil, nil
	}

	modelDashboard := transformDbDashboardToModel(*dbDashboard)
	return &model.GetValidatorDashboardResponse{
		Dashboard: &modelDashboard,
	}, nil
}

func (service *ApiService) CreateValidatorDashboard(ctx context.Context, in *model.CreateValidatorDashboardRequest) (*model.CreateValidatorDashboardResponse, error) {

	dummyUserId := uint64(1337) // For testing so that I dont have to wire up correct UserId handling for this PoC

	dbDashboard, err := service.dashboardRepository.CreateValidatorDashboard(ctx, dummyUserId)
	if err != nil {
		return nil, errors.New("internal Service Error")
	}

	modelDashboard := transformDbDashboardToModel(*dbDashboard)
	return &model.CreateValidatorDashboardResponse{
		Dashboard: &modelDashboard,
	}, nil
}

// transformDbDashboardToModel
// Converts the abstract database representation of a ValidatorDashboard to the proto model representation.
func transformDbDashboardToModel(dbDashboard dataaccess.ValidatorDashboard) model.ValidatorDashboard {
	dashboardId := fmt.Sprintf("%d", dbDashboard.DashboardId)
	return model.ValidatorDashboard{
		DashboardId: dashboardId,
		IsPublic:    dbDashboard.IsPublic,
	}
}
