package dataaccess

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/go-redis/redis/v8"
	model "github.com/gobitfly/beaconchain-backend/domain/gen"
)

func TestCachedAPIKeyRepository_GetAPIKey(t *testing.T) {
	ctx := context.Background()

	userID := uint64(42)
	apiKeyID := uuid.New()
	hashedKey := apikey.HashedKeyCredential{1, 2, 3}
	redisKey := fmt.Sprintf("%s%s", cacheAPIKeyPrefix, hashedKey.String())

	expectedAPIKey := apikey.APIKey{
		ID:     &apiKeyID,
		UserID: userID,
	}

	protoKey := &model.SerializableAPIKey{
		UserId: userID,
		ApiKeyId: &model.UUID{
			Value: apiKeyID[:],
		},
	}
	protoBytes, _ := proto.Marshal(protoKey)

	tests := []struct {
		name           string
		cacheHit       bool
		cacheErr       error
		fallbackAPIKey apikey.APIKey
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
			fallbackAPIKey: expectedAPIKey,
			expectSetCache: true,
		},
		{
			name:        "cache miss, fallback error",
			cacheHit:    false,
			fallbackErr: domain.ErrNotFound,
			expectError: true,
		},
		{
			name:        "redis error on get",
			cacheErr:    redis.ErrClosed,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient, redisMock := redismock.NewClientMock()
			apiKeyRepo := new(MockAPIKeyRepository)

			switch {
			case tt.cacheHit:
				redisMock.ExpectGet(redisKey).SetVal(string(protoBytes))
			case tt.cacheErr != nil:
				redisMock.ExpectGet(redisKey).SetErr(tt.cacheErr)
			default:
				redisMock.ExpectGet(redisKey).RedisNil()

				if tt.fallbackErr != nil || tt.fallbackAPIKey.ID != nil {
					apiKeyRepo.On("GetAPIKey", ctx, hashedKey).Return(tt.fallbackAPIKey, tt.fallbackErr)

					if tt.fallbackErr == nil && tt.expectSetCache {
						redisMock.ExpectSet(redisKey, protoBytes, time.Minute).SetVal("OK")
					}
				}
			}

			repo := &CachedAPIKeyRepository{
				redis:      redisClient,
				apikeyRepo: apiKeyRepo,
			}

			result, err := repo.GetAPIKey(ctx, hashedKey)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, apikey.APIKey{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedAPIKey, result)
			}

			assert.NoError(t, redisMock.ExpectationsWereMet())
			apiKeyRepo.AssertExpectations(t)
		})
	}
}
