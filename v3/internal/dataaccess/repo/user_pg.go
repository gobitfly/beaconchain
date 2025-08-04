package dataaccess

import (
	"context"

	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/data_sources"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type DBUserRepository struct {
	roConnectionAdminDb *sqlx.DB
	rwConnectionAdminDb *sqlx.DB
}

type dbUser struct {
	ID uint64 `db:"id"`
}

func (r *DBUserRepository) Initialize(roConnectionAdminDb data_sources.AdminRoConnection, rwConnectionAdminDb data_sources.AdminRwConnection) {
	r.roConnectionAdminDb = roConnectionAdminDb
	r.rwConnectionAdminDb = rwConnectionAdminDb
}

func (r *DBUserRepository) Ping() error {
	if r.roConnectionAdminDb == nil {
		return fmt.Errorf("read connection not initialized")
	}
	if err := r.roConnectionAdminDb.Ping(); err != nil {
		return fmt.Errorf("read connection ping failed: %w", err)
	}

	if r.rwConnectionAdminDb == nil {
		return fmt.Errorf("write connection not initialized")
	}
	if err := r.rwConnectionAdminDb.Ping(); err != nil {
		return fmt.Errorf("write connection ping failed: %w", err)
	}
	return nil
}

func (r *DBUserRepository) GetUserById(ctx context.Context, id uint64) (*domain.User, error) {
	return queryUser(ctx, r.roConnectionAdminDb, id)
}

func (r *DBUserRepository) GetUserByAPIKey(ctx context.Context, key apikey.HashedKeyCredential) (*domain.User, error) {
	subQuery := goqu.Dialect("postgres").From("api_keys_v2").
		Select("user_id").
		Where(
			goqu.And(
				goqu.C("api_key").Eq(key.Bytes()),
				goqu.C("deleted_at").IsNull(),
				goqu.C("disabled_at").IsNull(),
			),
		).
		Limit(1)

	return queryUser(ctx, r.roConnectionAdminDb, subQuery)
}

func queryUser[T uint64 | *goqu.SelectDataset](ctx context.Context, db *sqlx.DB, id T) (*domain.User, error) {
	// todo: authz
	ds := goqu.Dialect("postgres").From("users").
		Select("id").
		Where(goqu.C("id").Eq(id)).
		Limit(1)

	user, err := runQuery[dbUser](ctx, db, ds)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query user")
	}
	return &domain.User{
		ID: user.ID,
	}, nil
}

func (r *DBUserRepository) CreateUser(ctx context.Context, email string, initialApiKey string, hashedPassword string) (*domain.User, error) {
	user := domain.User{}

	err := r.rwConnectionAdminDb.GetContext(ctx, &user, `
	    	INSERT INTO users (password, email, register_ts, api_key)
	      		VALUES ($1, $2, NOW(), $3)
			RETURNING *`,
		hashedPassword, email, initialApiKey)

	return &user, err
}

func (r *DBUserRepository) DeleteUser(ctx context.Context, id uint64) error {
	_, err := r.rwConnectionAdminDb.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}
