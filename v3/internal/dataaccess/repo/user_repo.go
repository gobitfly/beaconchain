package dataaccess

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type DBUserRepository struct {
	roConnection *sqlx.DB
	rwConnection *sqlx.DB
}

func NewDBUserRepository(roConnection *sqlx.DB, rwConnection *sqlx.DB) *DBUserRepository {
	return &DBUserRepository{
		roConnection: roConnection,
		rwConnection: rwConnection,
	}
}

func (r *DBUserRepository) GetUserById(ctx context.Context, id uint64) (*User, error) {
	user := User{}

	err := r.roConnection.GetContext(ctx, &user, "SELECT * FROM users WHERE id=$1 LIMIT 1", id)
	return &user, err
}

func (r *DBUserRepository) GetUserByApiKey(ctx context.Context, apiKey string) (*User, error) {
	user := User{}
	err := r.roConnection.GetContext(ctx, &user, `SELECT * FROM users WHERE id IN (SELECT user_id FROM api_keys WHERE api_key = $1) LIMIT 1`, apiKey)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // No error and nothing returned means the User was not found
	}
	return &user, nil
}

func (r *DBUserRepository) CreateUser(ctx context.Context, email string, initialApiKey string, hashedPassword string) (*User, error) {
	user := User{}
	err := r.rwConnection.GetContext(ctx, &user, `
    	INSERT INTO users (password, email, register_ts, api_key)
      		VALUES ($1, $2, NOW(), $3)
		RETURNING *`,
		hashedPassword, email, initialApiKey)

	return &user, err
}

func (r *DBUserRepository) DeleteUser(ctx context.Context, id uint64) error {
	_, err := r.rwConnection.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}
