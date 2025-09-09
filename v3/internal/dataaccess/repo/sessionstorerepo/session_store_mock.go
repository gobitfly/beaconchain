package sessionstorerepo

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Initialize(redisClient *redis.Client) {
	m.Called(redisClient)
}

func (m *MockRepository) GetUserFromSessionID(ctx context.Context, sessionId string) (domain.User, error) {
	args := m.Called(ctx, sessionId)
	return args.Get(0).(domain.User), args.Error(1)
}
