package apikeyrepo

import (
	"context"
	"fmt"
	"strconv"
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
			apiKeyRepo := new(MockRepository)

			switch {
			case tt.cacheHit:
				redisMock.ExpectGet(redisKey).SetVal(string(protoBytes))
			case tt.cacheErr != nil:
				redisMock.ExpectGet(redisKey).SetErr(tt.cacheErr)
			default:
				redisMock.ExpectGet(redisKey).RedisNil()

				if tt.fallbackErr != nil || tt.fallbackAPIKey.ID != nil {
					apiKeyRepo.On("Get", ctx, hashedKey).Return(tt.fallbackAPIKey, tt.fallbackErr)

					if tt.fallbackErr == nil && tt.expectSetCache {
						redisMock.ExpectSet(redisKey, protoBytes, time.Minute).SetVal("OK")
					}
				}
			}

			repo := &CachedRepository{
				redis:      redisClient,
				apikeyRepo: apiKeyRepo,
			}

			result, err := repo.Get(ctx, hashedKey)

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

func TestUpdateAndGetAPIKeyLastUsedCache(t *testing.T) {
	ctx := context.Background()
	key := apikey.HashedKeyCredential{1, 2, 3}
	metaKey := fmt.Sprintf("%s%s", cacheAPIKeyLastUsedPrefix, key.String())

	type updateTestCase struct {
		name            string
		lastUsed        time.Time
		lastUsedFlush   *time.Time
		expectDoNothing bool
	}

	now := time.Now()
	tests := []updateTestCase{
		{"both lu and luf", now, &now, false},
		{"only lu", now, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient, redisMock := redismock.NewClientMock()
			repo := &CachedRepository{redis: redisClient}

			if !tt.expectDoNothing {
				redisMock.ExpectTxPipeline()
				updates := make(map[string]interface{})
				updates[cacheLastUsedField] = tt.lastUsed.UnixMilli()

				if tt.lastUsedFlush != nil {
					updates[cacheLastUsedFlushed] = tt.lastUsedFlush.UnixMilli()
				}

				// HSet in pipeline: need a return value
				redisMock.ExpectHSet(metaKey, updates).SetVal(int64(len(updates))) // return number of fields set

				// Expire in pipeline: need a return value
				redisMock.ExpectExpire(metaKey, cacheLastUsedTTL).SetVal(true)

				redisMock.ExpectTxPipelineExec()
			}

			err := repo.updateAPIKeyLastUsedCache(ctx, key, tt.lastUsed, tt.lastUsedFlush)
			assert.NoError(t, err)
			assert.NoError(t, redisMock.ExpectationsWereMet())
		})
	}

	type getTestCase struct {
		name      string
		redisVal  map[string]string
		redisErr  error
		expectLU  *int64
		expectLUF *int64
		expectErr error
	}

	luTS := time.Now().UnixMilli()
	lufTS := time.Now().Add(-time.Minute).UnixMilli()

	getTests := []getTestCase{
		{
			name: "both fields present",
			redisVal: map[string]string{
				cacheLastUsedField:   strconv.FormatInt(luTS, 10),
				cacheLastUsedFlushed: strconv.FormatInt(lufTS, 10),
			},
			expectLU:  &luTS,
			expectLUF: &lufTS,
			expectErr: nil,
		},
		{
			name:      "not found",
			redisVal:  map[string]string{},
			expectLU:  nil,
			expectLUF: nil,
			expectErr: domain.ErrNotFound,
		},
		{
			name:      "redis error",
			redisErr:  fmt.Errorf("redis down"),
			expectLU:  nil,
			expectLUF: nil,
			expectErr: fmt.Errorf("redis down"),
		},
		{
			name: "invalid integer",
			redisVal: map[string]string{
				cacheLastUsedField:   "abc",
				cacheLastUsedFlushed: "123",
			},
			expectLU:  nil,
			expectLUF: func() *int64 { v := int64(123); return &v }(),
			expectErr: nil,
		},
	}

	for _, tt := range getTests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient, redisMock := redismock.NewClientMock()
			repo := &CachedRepository{redis: redisClient}

			if tt.redisErr != nil {
				redisMock.ExpectHGetAll(metaKey).SetErr(tt.redisErr)
			} else {
				redisMock.ExpectHGetAll(metaKey).SetVal(tt.redisVal)
			}

			lu, luf, err := repo.getAPIKeyLastUsedCache(ctx, key)
			if tt.expectErr != nil {
				assert.Error(t, err)
				if tt.expectErr == domain.ErrNotFound {
					assert.ErrorIs(t, err, domain.ErrNotFound)
				}
			} else {
				assert.NoError(t, err)
				if tt.expectLU != nil {
					assert.Equal(t, *tt.expectLU, lu.UnixMilli())
				} else {
					assert.Equal(t, lu, time.Time{})
				}
				if tt.expectLUF != nil {
					assert.Equal(t, *tt.expectLUF, luf.UnixMilli())
				} else {
					assert.Equal(t, luf, time.Time{})
				}
			}

			assert.NoError(t, redisMock.ExpectationsWereMet())
		})
	}
}
