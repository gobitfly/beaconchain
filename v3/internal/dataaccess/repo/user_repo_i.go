package dataaccess

import (
	"context"
	"time"
)

type User struct {
	id                              uint64
	password                        string
	email                           string
	email_confirmed                 bool
	email_confirmation_hash         string
	email_confirmation_ts           time.Time
	password_reset_hash             string
	password_reset_ts               time.Time
	register_ts                     time.Time
	api_key                         string
	stripe_customer_id              string
	email_change_to_value           string
	user_group                      string
	stripe_email_pending            bool
	password_reset_not_allowed      bool
	notifications_do_not_disturb_ts time.Time
	isAdmin                         bool
}

type UserRepository interface {
	// Creates a new user
	CreateUser(ctx context.Context, email string, initialApiKey string, hashedPassword string) (*User, error)

	/**
	 * Creates an empty Validator Dashboard
	 */
	GetUserById(ctx context.Context, userId uint64) (*User, error)

	GetUserByApiKey(ctx context.Context, apikey string) (*User, error)

	// Returns nothing on success, or error if successfully deleted. Idempotent, if deleted when it DNE, no error should be returned.
	DeleteUser(ctx context.Context, userId uint64) error
}
