package app

import (
	"context"
	"errors"
	"strconv"

	model "github.com/gobitfly/beaconchain-api/api/gen"
	dataaccess "github.com/gobitfly/beaconchain-api/internal/dataaccess/repo"
)

func (service *ApiService) GetValidatorDashboard(ctx context.Context, in *model.GetValidatorDashboardRequest) (*model.GetValidatorDashboardResponse, error) {
	if len(in.DashboardId) <= 0 {
		return nil, errors.New("dashboard identifier must be specified")
	}

	// If dashboardId is specified, then return a specific dashboard for the user who sent the request. Ensure that the requestor only ever sees their own dashboards.
	dashboardId, err := strconv.Atoi(in.DashboardId)
	if err != nil {
		return nil, errors.New("invalid dashboard identifier")
	}

	// Get the user
	service.userRepository.GetUserById(ctx, 1234 /*TODO: Get userId from request headers via session/API Key*/)

	// In a proper implementation we would also want to make sure that this dashboard is actually owned by the requesting user.
	dbDashboard, err := service.dashboardRepository.GetValidatorDashboardByDashboardId(ctx, uint64(dashboardId))
	if err != nil {
		return nil, errors.New("Internal Service Error")
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

	dummyUserId := 1337 // For testing so that I dont have to wire up correct UserId handling for this PoC

	dbDashboard, err := service.dashboardRepository.CreateValidatorDashboard(ctx, uint64(dummyUserId))
	if err != nil {
		return nil, errors.New("Internal Service Error")
	}

	modelDashboard := transformDbDashboardToModel(*dbDashboard)
	return &model.CreateValidatorDashboardResponse{
		Dashboard: &modelDashboard,
	}, nil
}

/**
 * Converts the abstract database representation of a ValidatorDashboard to the proto model representation.
 */
func transformDbDashboardToModel(dbDashboard dataaccess.ValidatorDashboard) model.ValidatorDashboard {
	dashboardId := strconv.Itoa(int(dbDashboard.DashboardId))
	return model.ValidatorDashboard{
		DashboardId: dashboardId,
		IsPublic:    dbDashboard.IsPublic,
	}
}
