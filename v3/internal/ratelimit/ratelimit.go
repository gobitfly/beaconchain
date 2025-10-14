package ratelimit

import (
	"context"
	_ "embed"
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
	"google.golang.org/grpc"

	"github.com/gobitfly/beaconchain-backend/internal/limits"
)

//go:embed ratelimit_script.lua
var scriptStr string

// GetRateLimitMiddleware returns a gRPC middleware that applies rate limiting based on the caller's tier and endpoint.
// The service must pass a function to get the rate limit settings for a specific endpoint and tier.
func GetRateLimitMiddleware(client redis.Scripter, getEndpointRatelimit func(fullMethod string, tier domain.Tier) (*limits.RateLimitSettings, error)) grpc.UnaryServerInterceptor {
	script := redis.NewScript(scriptStr)
	limiter := limits.NewLimiter()
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		user, ok := auth.UserFromContext(ctx)
		if !ok {
			return nil, common.NewAPIInternalError(http.StatusInternalServerError, "no user in context")
		}
		globalRatelimit, _ := limiter.GetRateLimit(context.Background(), user)
		endpointRatelimit, err := getEndpointRatelimit(info.FullMethod, user.SubscriptionTier)
		if err != nil {
			log.Error(fmt.Errorf("error getting rate limit options: %w", err))
			return nil, common.NewAPIInternalError(http.StatusInternalServerError, "internal error: couldn't get rate limit options")
		}
		if endpointRatelimit == nil { // no rate limit defined for this endpoint, fallback to global rate limit
			endpointRatelimit = globalRatelimit
		}
		isWithinRateLimit := isWithinRateLimit(ctx, client, script,
			time.Now(),
			strconv.FormatUint(user.ID, 10),
			info.FullMethod,
			globalRatelimit,
			endpointRatelimit,
		)
		if !isWithinRateLimit {
			return nil, common.NewAPIUserFacingError(http.StatusTooManyRequests, "rate limit exceeded") // already http since we removed grpc errors
		}
		return handler(ctx, req)
	}
}

// isWithinRateLimit checks if the caller is allowed to make a request based on the rate limit settings.
func isWithinRateLimit(
	ctx context.Context,
	client redis.Scripter,
	script *redis.Script,
	now time.Time,
	callerID string,
	endpointID string,
	globalRateLimit *limits.RateLimitSettings,
	endpointRateLimit *limits.RateLimitSettings,
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

func calcRefillInterval(steadyRate float32) time.Duration {
	if steadyRate <= 0 {
		return 0
	}
	return time.Duration(decimal.NewFromInt(time.Second.Nanoseconds()).Div(decimal.NewFromFloat32(steadyRate)).InexactFloat64())
}
