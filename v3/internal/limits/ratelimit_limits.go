package limits

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

type RateLimitSettings struct {
	SteadyRate     float32
	BucketCapacity int
}

var rateLimits = map[domain.Tier]*RateLimitSettings{
	domain.TierFree: {
		SteadyRate:     0.5,
		BucketCapacity: 3,
	},
	domain.TierHobbyist: {
		SteadyRate:     1.0,
		BucketCapacity: 5,
	},
	domain.TierBusiness: {
		SteadyRate:     2.0,
		BucketCapacity: 10,
	},
	domain.TierScale: {
		SteadyRate:     5.0,
		BucketCapacity: 20,
	},
}

func (s *Limiter) GetRateLimit(ctx context.Context, user domain.User) (*RateLimitSettings, error) {
	return getLimitGeneric(user, rateLimits)
}
