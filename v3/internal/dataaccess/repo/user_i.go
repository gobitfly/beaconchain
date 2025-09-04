package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

type UserRepository interface {
	UserAuthRepository

	// to indicate if the repository is ready
	Ping() error

	// CreateUser
	// Creates a new user
	CreateUser(ctx context.Context, email string, initialApiKey string, hashedPassword string) (domain.User, error)

	// DeleteUser
	// Returns nothing on success, or error if successfully deleted. Idempotent, if deleted when it DNE, no error should be returned.
	DeleteUser(ctx context.Context, userId uint64) error
}

type UserAuthRepository interface {
	// GetUserById
	// Creates an empty Validator Dashboard
	GetUserById(ctx context.Context, userId uint64) (domain.User, error)
}
