package dataaccess

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type DBUserRepository struct {
	roConnectionAdminDb *sqlx.DB
	rwConnectionAdminDb *sqlx.DB
}

func NewDBUserRepository(roConnectionAdminDb *sqlx.DB, rwConnectionAdminDb *sqlx.DB) *DBUserRepository {
	return &DBUserRepository{
		roConnectionAdminDb: roConnectionAdminDb,
		rwConnectionAdminDb: rwConnectionAdminDb,
	}
}

func (r *DBUserRepository) GetUserById(ctx context.Context, id uint64) (*User, error) {
	user := User{}

	err := r.roConnectionAdminDb.GetContext(ctx, &user, "SELECT * FROM users WHERE id=$1 LIMIT 1", id)
	return &user, err
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
