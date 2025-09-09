package sessionstorerepo

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to encode the legacy SCS session structure used by GetUserFromSessionID
func encodeSession(deadline time.Time, values map[string]interface{}) ([]byte, error) {
	aux := &struct {
		Deadline time.Time
		Values   map[string]interface{}
	}{
		Deadline: deadline,
		Values:   values,
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(aux); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestGetUserFromSessionID_Success_OrcaMapsToHobbyist(t *testing.T) {
	ctx := context.Background()
	client, mock := redismock.NewClientMock()
	defer func() { _ = client.Close() }()

	repo := &DBRepository{}
	repo.Initialize(client)

	sessionID := "sess-ok"
	values := map[string]interface{}{
		"user_id":       uint64(12345),
		"authenticated": true,
		"subscription":  "saphire",
	}
	payload, err := encodeSession(time.Now().Add(1*time.Hour), values)
	require.NoError(t, err)

	mock.ExpectGet("scs:session:" + sessionID).SetVal(string(payload))

	user, err := repo.GetUserFromSessionID(ctx, sessionID)
	require.NoError(t, err)
	assert.Equal(t, uint64(12345), user.ID)
	assert.Equal(t, domain.TierHobbyist, user.SubscriptionTier)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserFromSessionID_Expired(t *testing.T) {
	ctx := context.Background()
	client, mock := redismock.NewClientMock()
	defer func() { _ = client.Close() }()

	repo := &DBRepository{}
	repo.Initialize(client)

	sessionID := "sess-expired"
	values := map[string]interface{}{
		"user_id":       uint64(1),
		"authenticated": true,
		"subscription":  "saphire",
	}
	payload, err := encodeSession(time.Now().Add(-1*time.Minute), values)
	require.NoError(t, err)

	mock.ExpectGet("scs:session:" + sessionID).SetVal(string(payload))

	_, err = repo.GetUserFromSessionID(ctx, sessionID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session deadline exceeded")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserFromSessionID_RedisError(t *testing.T) {
	ctx := context.Background()
	client, mock := redismock.NewClientMock()
	defer func() { _ = client.Close() }()

	repo := &DBRepository{}
	repo.Initialize(client)

	sessionID := "sess-missing"

	mock.ExpectGet("scs:session:" + sessionID).SetErr(errors.New("redis get error"))

	_, err := repo.GetUserFromSessionID(ctx, sessionID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis get error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserFromSessionID_GobDecodeError(t *testing.T) {
	ctx := context.Background()
	client, mock := redismock.NewClientMock()
	defer func() { _ = client.Close() }()

	repo := &DBRepository{}
	repo.Initialize(client)

	sessionID := "sess-bad-gob"

	// Provide invalid gob bytes
	mock.ExpectGet("scs:session:" + sessionID).SetVal("not-gob-bytes")

	_, err := repo.GetUserFromSessionID(ctx, sessionID)
	require.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserFromSessionID_NotAuthenticated(t *testing.T) {
	ctx := context.Background()
	client, mock := redismock.NewClientMock()
	defer func() { _ = client.Close() }()

	repo := &DBRepository{}
	repo.Initialize(client)

	sessionID := "sess-not-auth"
	values := map[string]interface{}{
		"user_id":       uint64(999),
		"authenticated": false,
		"subscription":  "something",
	}
	payload, err := encodeSession(time.Now().Add(10*time.Minute), values)
	require.NoError(t, err)

	mock.ExpectGet("scs:session:" + sessionID).SetVal(string(payload))

	_, err = repo.GetUserFromSessionID(ctx, sessionID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not authenticated")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserFromSessionID_DefaultSubscriptionMapsToFree(t *testing.T) {
	ctx := context.Background()
	client, mock := redismock.NewClientMock()
	defer func() { _ = client.Close() }()

	repo := &DBRepository{}
	repo.Initialize(client)

	sessionID := "sess-free"
	values := map[string]interface{}{
		"user_id":       uint64(777),
		"authenticated": true,
		"subscription":  "unknown-tier",
	}
	payload, err := encodeSession(time.Now().Add(10*time.Minute), values)
	require.NoError(t, err)

	mock.ExpectGet("scs:session:" + sessionID).SetVal(string(payload))

	user, err := repo.GetUserFromSessionID(ctx, sessionID)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, uint64(777), user.ID)
	assert.Equal(t, domain.TierFree, user.SubscriptionTier)
	assert.NoError(t, mock.ExpectationsWereMet())
}
