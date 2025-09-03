package dataaccess

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

func TestUpdateAndGetAPIKeyLastUsedMeta(t *testing.T) {
	ctx := context.Background()
	key := apikey.HashedKeyCredential{1, 2, 3}
	metaKey := fmt.Sprintf("%s%s", cacheAPIKeyMetaPrefix, key.String())

	type updateTestCase struct {
		name            string
		lastUsed        *time.Time
		lastUsedFlush   *time.Time
		expectDoNothing bool
	}

	now := time.Now()
	tests := []updateTestCase{
		{"both lu and luf", &now, &now, false},
		{"only lu", &now, nil, false},
		{"only luf", nil, &now, false},
		{"nothing", nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient, redisMock := redismock.NewClientMock()
			repo := &CachedAPIKeyRepository{redis: redisClient}

			if !tt.expectDoNothing {
				redisMock.ExpectTxPipeline()
				updates := make(map[string]interface{})
				if tt.lastUsed != nil {
					updates[cacheMetaLastUsedField] = tt.lastUsed.Unix()
				}
				if tt.lastUsedFlush != nil {
					updates[cacheMetaLastUsedFlushed] = tt.lastUsedFlush.Unix()
				}

				// HSet in pipeline: need a return value
				redisMock.ExpectHSet(metaKey, updates).SetVal(int64(len(updates))) // return number of fields set

				// Expire in pipeline: need a return value
				redisMock.ExpectExpire(metaKey, cacheMetaTTL).SetVal(true)

				redisMock.ExpectTxPipelineExec()
			}

			err := repo.updateAPIKeyLastUsedMeta(ctx, key, tt.lastUsed, tt.lastUsedFlush)
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

	luTS := time.Now().Unix()
	lufTS := time.Now().Add(-time.Minute).Unix()

	getTests := []getTestCase{
		{
			name: "both fields present",
			redisVal: map[string]string{
				cacheMetaLastUsedField:   strconv.FormatInt(luTS, 10),
				cacheMetaLastUsedFlushed: strconv.FormatInt(lufTS, 10),
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
				cacheMetaLastUsedField:   "abc",
				cacheMetaLastUsedFlushed: "123",
			},
			expectLU:  nil,
			expectLUF: func() *int64 { v := int64(123); return &v }(),
			expectErr: nil,
		},
	}

	for _, tt := range getTests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient, redisMock := redismock.NewClientMock()
			repo := &CachedAPIKeyRepository{redis: redisClient}

			if tt.redisErr != nil {
				redisMock.ExpectHGetAll(metaKey).SetErr(tt.redisErr)
			} else {
				redisMock.ExpectHGetAll(metaKey).SetVal(tt.redisVal)
			}

			lu, luf, err := repo.getAPIKeyLastUsedMeta(ctx, key)
			if tt.expectErr != nil {
				assert.Error(t, err)
				if tt.expectErr == domain.ErrNotFound {
					assert.ErrorIs(t, err, domain.ErrNotFound)
				}
			} else {
				assert.NoError(t, err)
				if tt.expectLU != nil {
					assert.Equal(t, *tt.expectLU, lu.Unix())
				} else {
					assert.Nil(t, lu)
				}
				if tt.expectLUF != nil {
					assert.Equal(t, *tt.expectLUF, luf.Unix())
				} else {
					assert.Nil(t, luf)
				}
			}

			assert.NoError(t, redisMock.ExpectationsWereMet())
		})
	}
}
