package userrepo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	model "github.com/gobitfly/beaconchain-backend/domain/gen"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestCachedUserRepository_GetUserById(t *testing.T) {
	ctx := context.Background()
	user := domain.User{ID: 42}
	redisKey := fmt.Sprintf("%s%d", cacheUserPrefix, user.ID)

	protoUser := &model.SerializableUser{Id: user.ID}
	protoBytes, _ := proto.Marshal(protoUser)

	tests := []struct {
		name           string
		cacheHit       bool
		cacheErr       error
		fallbackUser   domain.User
		fallbackErr    error
		expectSetCache bool
		expectError    bool
	}{
		{
			name:     "cache hit",
			cacheHit: true,
		},
		{
			name:           "cache miss, fallback success",
			cacheHit:       false,
			fallbackUser:   user,
			expectSetCache: true,
		},
		{
			name:        "cache miss, fallback error",
			cacheHit:    false,
			fallbackErr: domain.ErrNotFound,
			expectError: true,
		},
		{
			name:        "redis error",
			cacheErr:    redis.ErrClosed,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient, redisMock := redismock.NewClientMock()
			userRepo := new(MockRepository)

			switch {
			case tt.cacheHit:
				redisMock.ExpectGet(redisKey).SetVal(string(protoBytes))
			case tt.cacheErr != nil:
				redisMock.ExpectGet(redisKey).SetErr(tt.cacheErr)
			default:
				redisMock.ExpectGet(redisKey).RedisNil()

				if tt.fallbackUser != (domain.User{}) || tt.fallbackErr != nil {
					userRepo.On("Get", ctx, user.ID).Return(tt.fallbackUser, tt.fallbackErr)

					if tt.fallbackErr == nil && tt.expectSetCache {
						redisMock.ExpectSet(redisKey, protoBytes, time.Minute).SetVal("OK")
					}
				}
			}

			repo := &CachedRepository{
				redis:    redisClient,
				userRepo: userRepo,
			}

			result, err := repo.Get(ctx, user.ID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, domain.User{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, user, result)
			}

			assert.NoError(t, redisMock.ExpectationsWereMet())
			userRepo.AssertExpectations(t)
		})
	}
}
