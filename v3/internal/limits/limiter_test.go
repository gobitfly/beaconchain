package limits

import (
	"testing"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestGetLimitGeneric(t *testing.T) {
	tests := []struct {
		name        string
		userTier    domain.Tier
		defaults    map[domain.Tier]string
		expectErr   error
		expectValue string
	}{
		{
			name:        "known tier found",
			userTier:    domain.TierBusiness,
			defaults:    map[domain.Tier]string{domain.TierBusiness: "business-limit"},
			expectErr:   nil,
			expectValue: "business-limit",
		},
		{
			name:        "unknown tier falls back to free",
			userTier:    "UNKNOWN",
			defaults:    map[domain.Tier]string{domain.TierFree: "free-limit"},
			expectErr:   ErrFallbackToFreeTier,
			expectValue: "free-limit",
		},
		{
			name:        "no matching tier and no free fallback",
			userTier:    "UNKNOWN",
			defaults:    map[domain.Tier]string{},
			expectErr:   assert.AnError, // just check for some error
			expectValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &domain.User{SubscriptionTier: tt.userTier}

			val, err := getLimitGeneric(user, tt.defaults)
			switch tt.expectErr {
			case nil:
				assert.NoError(t, err)
			case ErrFallbackToFreeTier:
				assert.ErrorIs(t, err, ErrFallbackToFreeTier)
			default:
				assert.Error(t, err)
			}
			assert.Equal(t, tt.expectValue, val)
		})
	}
}
