package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
)

type APIKeyRepository interface {
	APIKeyManagementRepository
	APIKeyAuthRepository
}

type APIKeyManagementRepository interface {
	CreateAPIKey(ctx context.Context, userID uint64, key apikey.APIKey) (apikey.APIKey, error)

	DeleteAPIKey(ctx context.Context, userID uint64, name string) error

	DisableAPIKey(ctx context.Context, userID uint64, name string) (apikey.APIKey, error)

	EnableAPIKey(ctx context.Context, userID uint64, name string) (apikey.APIKey, error)

	GetAPIKeys(ctx context.Context, userID uint64, name *string) ([]apikey.APIKey, error)
}

type APIKeyAuthRepository interface {
	GetAPIKey(ctx context.Context, apikey apikey.HashedKeyCredential) (apikey.APIKey, error)
	UpdateLastUsedAt(ctx context.Context, key apikey.HashedKeyCredential) error
}
