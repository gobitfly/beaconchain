package dataaccess

import "context"

type ValidatorDashboard struct {
	DashboardId uint64
	UserId      uint64
	IsPublic    bool
}

type ValidatorDashboardRepository interface {
	// to indicate if the repository is ready
	Ping() error

	// CreateValidatorDashboard
	// Creates an empty Validator Dashboard
	CreateValidatorDashboard(ctx context.Context, userId uint64) (*ValidatorDashboard, error)

	// GetValidatorDashboardsByUserId
	// Returns basic info of a dashboard for a particular user.
	// userId must be provided
	// If no dashboardId is provided, return all dashboards for that user
	GetValidatorDashboardsByUserId(ctx context.Context, userId uint64) (*[]ValidatorDashboard, error)

	GetValidatorDashboardByDashboardId(ctx context.Context, dashboardId uint64) (*ValidatorDashboard, error)

	// ModifyValidatorDashboard
	// Modifies the attributes of a dashboard.
	// If an attribute is not included or is nil in the model, it should not be updated.
	// If an attribute is included but modification of that is not possible, an error should be returned
	// and other attribute modifications should not take place.
	ModifyValidatorDashboard(ctx context.Context, dashboardId uint64, isPublic bool) (*ValidatorDashboard, error)

	// DeleteValidatorDashboard
	// Returns nothing on success, or error if successfully deleted. Idempotent, if deleted when it DNE, no error should be returned.
	DeleteValidatorDashboard(ctx context.Context, dashboardId uint64) error
}
