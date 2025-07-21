package dataaccess

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gobitfly/beaconchain-backend/internal/dataaccess/data_sources"
	"github.com/jmoiron/sqlx"
)

type DBUserRepository struct {
	roConnectionAdminDb *sqlx.DB
	rwConnectionAdminDb *sqlx.DB
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

func (r *DBUserRepository) GetUserById(ctx context.Context, id uint64) (*User, error) {
	user := User{}

	_ = r.roConnectionAdminDb.GetContext(ctx, &user, "SELECT * FROM users WHERE id=$1 LIMIT 1", id)
	return &user, nil
}

func (r *DBUserRepository) GetUserByApiKey(ctx context.Context, apikey string) (*User, error) {
	user := User{}
	err := r.roConnectionAdminDb.GetContext(ctx, &user, `SELECT * FROM users WHERE id IN (SELECT user_id FROM api_keys WHERE api_key = $1) LIMIT 1`, apikey)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // No error and nothing returned means the User was not found
	}
	return &user, nil
}

func (r *DBUserRepository) CreateUser(ctx context.Context, email string, initialApiKey string, hashedPassword string) (*User, error) {
	user := User{}

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
