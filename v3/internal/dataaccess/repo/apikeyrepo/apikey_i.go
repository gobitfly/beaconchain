package apikeyrepo

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
)

type Repository interface {
	ManagementRepository
	AuthRepository
}

type ManagementRepository interface {
	Create(ctx context.Context, userID uint64, key apikey.APIKey) (apikey.APIKey, error)

	Delete(ctx context.Context, userID uint64, name string) error

	Disable(ctx context.Context, userID uint64, name string) (apikey.APIKey, error)

	Enable(ctx context.Context, userID uint64, name string) (apikey.APIKey, error)

	GetAll(ctx context.Context, userID uint64, name *string) ([]apikey.APIKey, error)
}

type AuthRepository interface {
	Get(ctx context.Context, apikey apikey.HashedKeyCredential) (apikey.APIKey, error)
	UpdateLastUsedAt(ctx context.Context, key apikey.HashedKeyCredential) error
}
