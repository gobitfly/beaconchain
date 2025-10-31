package ratelimit

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	"github.com/shopspring/decimal"

	"github.com/gobitfly/beaconchain-backend/internal/limits"
)

//go:embed ratelimit_script.lua
var scriptStr string

// GetEndpointSettingsFunc extracts rate limit configuration from HTTP request.
type GetEndpointSettingsFunc func(request *http.Request, tier domain.Tier) (operationID string, settings *domain.RateLimitSettings)

// Middleware returns an HTTP middleware that applies rate limiting based on the caller's tier and endpoint.
// The service must pass a function to get the rate limit settings for a specific endpoint through the HTTP request.
// If `getEndpointSettings` returns a nil value for the settings, the global rate limit is applied to that endpoint.
func Middleware(client redis.Scripter, getEndpointSettings GetEndpointSettingsFunc, errorHandler func(w http.ResponseWriter, r *http.Request, err error)) func(next http.Handler) http.Handler {
	script := redis.NewScript(scriptStr)
	limiter := limits.NewLimiter()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			user, ok := auth.UserFromContext(ctx)
			if !ok {
				errorHandler(w, r, errors.New("user not found in context"))
				return
			}
			globalRateLimit, err := limiter.GetRateLimit(ctx, user)
			if err != nil {
				errorHandler(w, r, fmt.Errorf("failed to get global rate limit: %w", err))
				return
			}
			operationID, endpointRateLimit := getEndpointSettings(r, user.SubscriptionTier)
			if endpointRateLimit == nil {
				endpointRateLimit = &globalRateLimit
			}
			hasToken := isWithinRateLimit(ctx, client, script,
				time.Now(),
				strconv.FormatUint(user.ID, 10),
				operationID,
				globalRateLimit,
				*endpointRateLimit,
			)
			if !hasToken {
				errorHandler(w, r, common.NewAPIUserFacingError(http.StatusTooManyRequests, "rate limit exceeded"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// isWithinRateLimit attempts to get a token from the rate limit buckets.
// Returns true if the request is allowed, false if rate limit is exceeded.
// In case of an error (e.g., Redis outage), it logs the error and allows the request.
func isWithinRateLimit(
	ctx context.Context,
	client redis.Scripter,
	script *redis.Script,
	now time.Time,
	callerID string,
	endpointID string,
	globalRateLimit domain.RateLimitSettings,
	endpointRateLimit domain.RateLimitSettings,
) bool {
	globalRefillInterval := calcRefillInterval(globalRateLimit.SteadyRate)
	endpointRefillInterval := calcRefillInterval(endpointRateLimit.SteadyRate)
	val, err := script.Run(ctx, client, []string{getCallerKey(callerID)},
		now.UnixMilli(),
		globalRateLimit.BucketCapacity,
		globalRefillInterval.Milliseconds(),
		endpointRateLimit.BucketCapacity,
		endpointRefillInterval.Milliseconds(),
		endpointID,
	).Int() // returns 1 if allowed, 0 if not allowed
	if err != nil {
		log.Error(fmt.Errorf("error executing rate limit script: %w", err))
		return true // In case of a redis outage, we do not want to block the request
	}

	return val != 0
}

func getCallerKey(callerID string) string {
	return "ratelimit:" + callerID
}

// calcRefillInterval calculates the refill interval based on the steady rate (tokens per second).
func calcRefillInterval(steadyRate float32) time.Duration {
	if steadyRate <= 0 {
		return 0
	}
	return time.Duration(decimal.NewFromInt(time.Second.Nanoseconds()).Div(decimal.NewFromFloat32(steadyRate)).InexactFloat64())
}
