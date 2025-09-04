package limits

import (
	"context"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

var rateLimits = map[domain.Tier]*model.RateLimitSettings{
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

func (s *Limiter) GetRateLimit(ctx context.Context, user domain.User) (*model.RateLimitSettings, error) {
	return getLimitGeneric(user, rateLimits)
}
