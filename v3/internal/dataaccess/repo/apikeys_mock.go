package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/stretchr/testify/mock"
)

type MockAPIKeyRepository struct {
	mock.Mock
}

func (m *MockAPIKeyRepository) CreateAPIKey(ctx context.Context, userID uint64, key apikey.APIKey) (apikey.APIKey, error) {
	args := m.Called(ctx, userID, key)
	return args.Get(0).(apikey.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) DeleteAPIKey(ctx context.Context, userID uint64, name string) error {
	args := m.Called(ctx, userID, name)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) DisableAPIKey(ctx context.Context, userID uint64, name string) (apikey.APIKey, error) {
	args := m.Called(ctx, userID, name)
	return args.Get(0).(apikey.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) EnableAPIKey(ctx context.Context, userID uint64, name string) (apikey.APIKey, error) {
	args := m.Called(ctx, userID, name)
	return args.Get(0).(apikey.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) GetAPIKeys(ctx context.Context, userID uint64, name *string) ([]apikey.APIKey, error) {
	args := m.Called(ctx, userID, name)
	return args.Get(0).([]apikey.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) GetAPIKeyCount(ctx context.Context, userID uint64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAPIKeyRepository) UpdateLastUsedAt(ctx context.Context, key apikey.HashedKeyCredential) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) GetAPIKey(ctx context.Context, key apikey.HashedKeyCredential) (apikey.APIKey, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(apikey.APIKey), args.Error(1)
}
