package limits

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

var maxAPIKeyLimits = map[domain.Tier]int{
	domain.TierFree:     1,
	domain.TierHobbyist: 2,
	domain.TierBusiness: 5,
	domain.TierScale:    10,
}

func (s *Limiter) GetMaxAPIKeys(ctx context.Context, user domain.User) (int, error) {
	return getLimitGeneric(user, maxAPIKeyLimits)
}
