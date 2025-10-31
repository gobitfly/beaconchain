package auth

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

type CtxKey string

const ctxUserKey CtxKey = "user"
const ctxHashedAPIKey CtxKey = "hashed_api_key" // #nosec G101

type Header string

const (
	APIKeyHeader Header = "apikey"
)

func SetUserInContext(ctx context.Context, user domain.User) context.Context {
	return context.WithValue(ctx, ctxUserKey, user)
}

// UserFromContext retrieves the user from context
func UserFromContext(ctx context.Context) (domain.User, bool) {
	user, ok := ctx.Value(ctxUserKey).(domain.User)
	return user, ok
}

// MustUserFromContext should be used only when a user context is always expected such as in all external routes.
// Use UserFromContext if the user may not be present (selective internal routes).
func MustUserFromContext(ctx context.Context) domain.User {
	user, ok := UserFromContext(ctx)
	if !ok {
		panic("user not found in context: middleware contract broken")
	}
	return user
}

func SetAPIKeyInContext(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, ctxHashedAPIKey, apiKey)
}

// APIKeyFromContext retrieves the hashed API key from context.
// It acts as an identifier for the api key used in the request.
func APIKeyFromContext(ctx context.Context) (string, bool) {
	apiKey, ok := ctx.Value(ctxHashedAPIKey).(string)
	return apiKey, ok
}

// MustAPIKeyFromContext should be used only when a API key context is always expected such as in all external routes.
// Not applicable for internal routes where API key is never present.
func MustAPIKeyFromContext(ctx context.Context) string {
	apiKey, ok := APIKeyFromContext(ctx)
	if !ok || apiKey == "" {
		panic("API key not found in context: middleware contract broken")
	}
	return apiKey
}
