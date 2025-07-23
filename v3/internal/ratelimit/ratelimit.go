package ratelimit

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	"github.com/gobitfly/beaconchain-backend/internal/subscription_products"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
)

// TODO: embed the Lua script for rate limiting
var scriptStr string

// GetRateLimitMiddleware returns a gRPC middleware that applies rate limiting based on the caller's tier and endpoint.
// The service must pass a function to get the rate limit settings for a specific endpoint and tier.
func GetRateLimitMiddleware(client redis.Scripter, getEndpointRatelimit func(fullMethod string, tier subscription_products.Tier) (*model.RateLimitSettings, error)) grpc.UnaryServerInterceptor {
	script := redis.NewScript(scriptStr)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		callerID := "caller123"                    // TODO: Replace with actual caller ID
		tier := subscription_products.TierHobbyist // TODO: Replace with actual tier
		globalRatelimit := subscription_products.SubscriptionPerksMap[tier].GlobalRateLimit
		endpointRatelimit, err := getEndpointRatelimit(info.FullMethod, tier)
		if err != nil {
			log.Error(fmt.Errorf("error getting rate limit options: %w", err))
			return nil, status.Errorf(codes.Internal, "internal error: couldn't get rate limit options")
		}
		if endpointRatelimit == nil { // no rate limit defined for this endpoint, fallback to global rate limit
			endpointRatelimit = globalRatelimit
		}
		err = limit(ctx, client, script,
			callerID,
			info.FullMethod,
			globalRatelimit,
			endpointRatelimit,
		)
		if err != nil {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %v", err)
		}
		return handler(ctx, req)
	}
}

// limit applies the rate limit for the given caller and endpoint using Redis.
func limit(
	ctx context.Context,
	client redis.Scripter,
	script *redis.Script,
	callerID string,
	endpointID string,
	globalRateLimit *model.RateLimitSettings,
	endpointRateLimit *model.RateLimitSettings,
) error {
	// allow all requests for now
	// TODO: Implement actual rate limiting logic
	return nil
}
