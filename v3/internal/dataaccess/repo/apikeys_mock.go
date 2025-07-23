package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/stretchr/testify/mock"
)

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) CreateAPIKey(ctx context.Context, userID uint64, key apikey.APIKey) (apikey.APIKey, error) {
	args := m.Called(ctx, userID, key)
	return args.Get(0).(apikey.APIKey), args.Error(1)
}

func (m *MockAuthRepository) DeleteAPIKey(ctx context.Context, userID uint64, name string) error {
	args := m.Called(ctx, userID, name)
	return args.Error(0)
}

func (m *MockAuthRepository) DisableAPIKey(ctx context.Context, userID uint64, name string) (apikey.APIKey, error) {
	args := m.Called(ctx, userID, name)
	return args.Get(0).(apikey.APIKey), args.Error(1)
}

func (m *MockAuthRepository) EnableAPIKey(ctx context.Context, userID uint64, name string) (apikey.APIKey, error) {
	args := m.Called(ctx, userID, name)
	return args.Get(0).(apikey.APIKey), args.Error(1)
}

func (m *MockAuthRepository) GetAPIKeys(ctx context.Context, userID uint64, name *string) ([]apikey.APIKey, error) {
	args := m.Called(ctx, userID, name)
	return args.Get(0).([]apikey.APIKey), args.Error(1)
}
