package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
)

type AuthRepository interface {
	CreateAPIKey(ctx context.Context, userID uint64, key apikey.APIKey) (apikey.APIKey, error)

	DeleteAPIKey(ctx context.Context, userID uint64, name string) error

	DisableAPIKey(ctx context.Context, userID uint64, name string) (apikey.APIKey, error)

	EnableAPIKey(ctx context.Context, userID uint64, name string) (apikey.APIKey, error)

	GetAPIKeys(ctx context.Context, userID uint64, name *string) ([]apikey.APIKey, error)
}
