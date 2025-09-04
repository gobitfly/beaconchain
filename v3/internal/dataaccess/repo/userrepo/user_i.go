package userrepo

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

type Repository interface {
	AuthRepository

	// to indicate if the repository is ready
	Ping() error

	// Create
	// Creates a new user
	Create(ctx context.Context, email string, initialApiKey string, hashedPassword string) (domain.User, error)

	// Delete
	// Returns nothing on success, or error if successfully deleted. Idempotent, if deleted when it DNE, no error should be returned.
	Delete(ctx context.Context, userId uint64) error
}

type AuthRepository interface {
	// Get
	// Creates an empty Validator Dashboard
	Get(ctx context.Context, userId uint64) (domain.User, error)
}
