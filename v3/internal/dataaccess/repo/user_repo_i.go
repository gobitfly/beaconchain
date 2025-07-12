package dataaccess

import (
	"context"
)

type User struct {
	// id      uint64
	// isAdmin bool
}

type UserRepository interface {
	// CreateUser
	// Creates a new user
	CreateUser(ctx context.Context, email string, initialApiKey string, hashedPassword string) (*User, error)

	// GetUserById
	// Creates an empty Validator Dashboard
	GetUserById(ctx context.Context, userId uint64) (*User, error)

	GetUserByApiKey(ctx context.Context, apikey string) (*User, error)

	// DeleteUser
	// Returns nothing on success, or error if successfully deleted. Idempotent, if deleted when it DNE, no error should be returned.
	DeleteUser(ctx context.Context, userId uint64) error
}
