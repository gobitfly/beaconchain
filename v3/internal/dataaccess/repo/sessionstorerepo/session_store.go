package sessionstorerepo

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
)

const (
	LegacySessionPrefix     = "scs:session:"
	UserIdSessionKey        = "user_id"
	AuthenticatedSessionKey = "authenticated"
	SubscriptionSessionKey  = "subscription"
)

type UserV1Notification int

func init() {
	gob.RegisterName("github.com/gobitfly/eth2-beaconchain-explorer/types.UserV1Notification", UserV1Notification(0))
}

type DBRepository struct {
	redis *redis.Client
}

func (s *DBRepository) Initialize(redisClient *redis.Client) {
	s.redis = redisClient
}

// GetUserFromSessionID retrieves the current user using a given session ID
// As session generation is done using the v1 code, this function needs to
// ensure compatibility with the SCS library
// Session data are stored as a GOB encoded object that contains a deadline
// and a key / value map for data storage
func (s *DBRepository) GetUserFromSessionID(ctx context.Context, sessionId string) (domain.User, error) {
	serializedUser, err := s.redis.Get(ctx, LegacySessionPrefix+sessionId).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	deadline, values, err := s.decodeSessionData(serializedUser)
	if err != nil {
		return domain.User{}, err
	}

	if deadline.Before(time.Now()) {
		return domain.User{}, errors.New("session deadline exceeded")
	}
	user := domain.User{}
	user.ID = values[UserIdSessionKey].(uint64)

	authenticated := values[AuthenticatedSessionKey].(bool)
	if !authenticated {
		return domain.User{}, errors.New("user not authenticated")
	}

	subscription := values[SubscriptionSessionKey].(string)
	user.SubscriptionTier = s.mapLegacySubscription(subscription)

	// TODO: should unlock admin functionality?
	// userGroup := values["user_group"].(string)
	return user, nil
}

func (s *DBRepository) decodeSessionData(data []byte) (time.Time, map[string]interface{}, error) {
	aux := &struct {
		Deadline time.Time
		Values   map[string]interface{}
	}{}

	r := bytes.NewReader(data)
	if err := gob.NewDecoder(r).Decode(&aux); err != nil {
		return time.Time{}, nil, err
	}

	return aux.Deadline, aux.Values, nil
}

func (s *DBRepository) mapLegacySubscription(legacySubscriptionName string) domain.Tier {
	// TODO: add mapping to match old subscriptions to new ones
	switch legacySubscriptionName {
	case "saphire":
		return domain.TierHobbyist
	case "emerald":
		return domain.TierHobbyist
	case "diamond":
		return domain.TierBusiness
	default:
		return domain.TierFree
	}
}
