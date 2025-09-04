package limits

import (
	"errors"
	"fmt"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

var ErrFallbackToFreeTier = errors.New("unknown subscription tier, falling back to free tier")

type Limiter struct{}

func NewLimiter() *Limiter {
	return &Limiter{}
}

// GetLimitGeneric is a generic function to get limits based on user tier with fallback to free tier.
// It takes a map of limits per tier and returns the limit for the user's tier.
// If the user's tier is unknown, it falls back to the free tier limit.
func getLimitGeneric[T any](
	user domain.User,
	defaults map[domain.Tier]T,
) (T, error) {
	// Fallback to default
	if limit, ok := defaults[user.SubscriptionTier]; ok {
		return limit, nil
	}
	// Unknown tier, fallback to free tier
	if limit, ok := defaults[domain.TierFree]; ok {
		return limit, ErrFallbackToFreeTier
	}

	var zero T
	return zero, fmt.Errorf("limit not found for tier: %s", user.SubscriptionTier)
}
