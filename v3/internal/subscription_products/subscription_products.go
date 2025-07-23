package subscription_products

import (
	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
)

type Tier string

const (
	TierFree     Tier = "FREE"
	TierHobbyist Tier = "HOBBYIST"
	TierBusiness Tier = "BUSINESS"
	TierScale    Tier = "SCALE"
)

type SubscriptionPerks struct {
	GlobalRateLimit *model.RateLimitSettings
}

// Source of truth for subscription perks
var SubscriptionPerksMap = map[Tier]SubscriptionPerks{
	TierFree: {
		GlobalRateLimit: &model.RateLimitSettings{
			SteadyRate:     0.5,
			BucketCapacity: 3,
		},
	},
	TierHobbyist: {
		GlobalRateLimit: &model.RateLimitSettings{
			SteadyRate:     1.0,
			BucketCapacity: 5,
		},
	},
	TierBusiness: {
		GlobalRateLimit: &model.RateLimitSettings{
			SteadyRate:     2.0,
			BucketCapacity: 10,
		},
	},
	TierScale: {
		GlobalRateLimit: &model.RateLimitSettings{
			SteadyRate:     5.0,
			BucketCapacity: 20,
		},
	},
}
