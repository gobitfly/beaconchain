package dataaccess

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
)

type DBValidatorDashboardRepository struct {
	roConnection *sqlx.DB
	rwConnection *sqlx.DB
}

func NewDBValidatorDashboardRepository(roConnection *sqlx.DB, rwConnection *sqlx.DB) *DBValidatorDashboardRepository {
	return &DBValidatorDashboardRepository{
		roConnection: roConnection,
		rwConnection: rwConnection,
	}
}

/**
 * Creates an empty Validator Dashboard
 */
func (r *DBValidatorDashboardRepository) CreateValidatorDashboard(ctx context.Context, userId uint64) (*ValidatorDashboard, error) {
	return nil, errors.New("Unimplemented")
}

/**
 * Returns basic info of a dashboard for a particular user.
 * userId must be provided
 * If no dashboardId is provided, return all dashboards for that user
 */
func (r *DBValidatorDashboardRepository) GetValidatorDashboardsByUserId(ctx context.Context, userId uint64) (*[]ValidatorDashboard, error) {
	return nil, errors.New("Unimplemented")

}

func (r *DBValidatorDashboardRepository) GetValidatorDashboardByDashboardId(ctx context.Context, dashboardId uint64) (*ValidatorDashboard, error) {
	return nil, errors.New("Unimplemented")

}

/**
 * Modifies the attributes of a dashboard.
 * If an attribute is not included or is nil in the model, it should not be updated.
 * If an attribute is included but modification of that is not possible, an error should be returned
 *    and other attribute modifications should not take place.
 */
func (r *DBValidatorDashboardRepository) ModifyValidatorDashboard(ctx context.Context, dashboardId uint64, isPublic bool) (*ValidatorDashboard, error) {
	return nil, errors.New("Unimplemented")

}

// Returns nothing on success, or error if successfully deleted. Idempotent, if deleted when it DNE, no error should be returned.
func (r *DBValidatorDashboardRepository) DeleteValidatorDashboard(ctx context.Context, dashboardId uint64) error {
	return errors.New("Unimplemented")
}
