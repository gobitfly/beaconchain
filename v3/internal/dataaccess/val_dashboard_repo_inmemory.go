package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain-api/internal/log"
)

type InMemoryValidatorDashboardRepository struct {
	dashboardDatabase map[int]*ValidatorDashboard // DashboardId => ValidatorDashboard
	lastId            int
}

func NewInMemoryValidatorDashboardRepository() *InMemoryValidatorDashboardRepository {
	log.Info("Successfully initialized repo")
	return &InMemoryValidatorDashboardRepository{
		dashboardDatabase: make(map[int]*ValidatorDashboard), // DashboardId => Dashboard
		lastId:            0,
	}
}

/**
 * Creates an empty Validator Dashboard
 */
func (r *InMemoryValidatorDashboardRepository) CreateValidatorDashboard(ctx context.Context, userId int) (*ValidatorDashboard, error) {

	// Create a new dashboard
	r.lastId = r.lastId + 1
	newDashboard := ValidatorDashboard{
		DashboardId: r.lastId,
		UserId:      userId,
		IsPublic:    false}

	// Insert the new dashboards in to the slice of user dashboards
	r.dashboardDatabase[newDashboard.DashboardId] = &newDashboard

	return &newDashboard, nil
}

/**
 * Returns basic info of a dashboard for a particular user.
 * userId must be provided
 * If no dashboardId is provided, return all dashboards for that user
 * If the user does not exist, return empty dashboard list
 * If the dashboardId does not exist, skip over it (returning empty dashboard list if none found)
 */
func (r *InMemoryValidatorDashboardRepository) GetValidatorDashboardsByUserId(ctx context.Context, userId int) (*[]ValidatorDashboard, error) {

	foundDashboards := make([]ValidatorDashboard, 0)
	for _, dashboard := range r.dashboardDatabase {
		if dashboard.UserId == userId {
			foundDashboards = append(foundDashboards, *dashboard)
		}
	}

	return &foundDashboards, nil
}

func (r *InMemoryValidatorDashboardRepository) GetValidatorDashboardByDashboardId(ctx context.Context, dashboardId int) (*ValidatorDashboard, error) {
	log.Info("got this far really")
	return r.dashboardDatabase[dashboardId], nil
}

/**
 * Modifies the attributes of a dashboard.
 * If an attribute is not included or is nil in the model, it should not be updated.
 * If an attribute is included but modification of that is not possible, an error should be returned
 *    and other attribute modifications should not take place.
 */
func (r *InMemoryValidatorDashboardRepository) ModifyValidatorDashboard(ctx context.Context, dashboardId int, isPublic bool) (*ValidatorDashboard, error) {
	dashboard := r.dashboardDatabase[dashboardId]
	dashboard.IsPublic = isPublic
	return dashboard, nil
}

// Returns nothing on success, or error if successfully deleted. Idempotent, if deleted when it DNE, no error should be returned.
func (r *InMemoryValidatorDashboardRepository) DeleteValidatorDashboard(ctx context.Context, dashboardId int) error {
	// Do nothing, pretend it was deleted
	delete(r.dashboardDatabase, dashboardId)
	return nil
}
