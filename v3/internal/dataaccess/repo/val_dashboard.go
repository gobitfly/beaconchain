package dataaccess

import (
	"context"
	"errors"
	"fmt"

	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/data_sources"
	"github.com/jmoiron/sqlx"
)

type DBValidatorDashboardRepository struct {
	roConnection *sqlx.DB
	rwConnection *sqlx.DB

	roChConnection *sqlx.DB
	rwChConnection *sqlx.DB

	redis *data_sources.RedisCache

	bigtable *data_sources.Bigtable
}

func (r *DBValidatorDashboardRepository) Initialize(roConnection data_sources.ChainRoConnection, rwConnection data_sources.ChainRwConnection, roChConnection data_sources.ClickhouseRoConnection, rwChConnection data_sources.ClickhouseRwConnection, redis *data_sources.RedisCache, bigtable *data_sources.Bigtable) {
	r.roConnection = roConnection
	r.rwConnection = rwConnection

	r.roChConnection = roChConnection
	r.rwChConnection = rwChConnection

	r.redis = redis

	r.bigtable = bigtable
}

func (r *DBValidatorDashboardRepository) Ping() error {
	dbs := []*sqlx.DB{
		r.roConnection,
		r.rwConnection,
		r.roChConnection,
		r.rwChConnection,
	}
	for _, db := range dbs {
		if db == nil {
			return fmt.Errorf("database connection not initialized")
		}
		if err := db.Ping(); err != nil {
			return fmt.Errorf("database connection ping failed: %w", err)
		}
	}

	if r.redis == nil || r.redis.RedisRemoteCache == nil {
		return fmt.Errorf("redis connection not initialized")
	}
	if err := r.redis.RedisRemoteCache.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis connection ping failed: %w", err)
	}

	// no built-in ping for bigtable, TODO write custom
	return nil
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
