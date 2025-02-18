package dataaccess

import "context"

type ValidatorDashboard struct {
	DashboardId int
	UserId      int
	IsPublic    bool
}

type ValidatorDashboardRepository interface {
	/**
	 * Creates an empty Validator Dashboard
	 */
	CreateValidatorDashboard(ctx context.Context, userId int) (*ValidatorDashboard, error)

	/**
	 * Returns basic info of a dashboard for a particular user.
	 * userId must be provided
	 * If no dashboardId is provided, return all dashboards for that user
	 */
	GetValidatorDashboardsByUserId(ctx context.Context, userId int) (*[]ValidatorDashboard, error)

	GetValidatorDashboardByDashboardId(ctx context.Context, dashboardId int) (*ValidatorDashboard, error)

	/**
	 * Modifies the attributes of a dashboard.
	 * If an attribute is not included or is nil in the model, it should not be updated.
	 * If an attribute is included but modification of that is not possible, an error should be returned
	 *    and other attribute modifications should not take place.
	 */
	ModifyValidatorDashboard(ctx context.Context, dashboardId int, isPublic bool) (*ValidatorDashboard, error)

	// Returns nothing on success, or error if successfully deleted. Idempotent, if deleted when it DNE, no error should be returned.
	DeleteValidatorDashboard(ctx context.Context, dashboardId int) error
}
