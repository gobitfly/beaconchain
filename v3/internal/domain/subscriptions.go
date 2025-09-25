package domain

type Tier string

const (
	TierFree     Tier = "FREE"
	TierHobbyist Tier = "HOBBYIST"
	TierBusiness Tier = "BUSINESS"
	TierScale    Tier = "SCALE"
)

type RateLimitSettings struct {
	SteadyRate     float32
	BucketCapacity int
}
