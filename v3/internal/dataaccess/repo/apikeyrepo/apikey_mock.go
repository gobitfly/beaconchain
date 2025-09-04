package apikeyrepo

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/auth/apikey"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, userID uint64, key apikey.APIKey) (apikey.APIKey, error) {
	args := m.Called(ctx, userID, key)
	return args.Get(0).(apikey.APIKey), args.Error(1)
}

func (m *MockRepository) Delete(ctx context.Context, userID uint64, name string) error {
	args := m.Called(ctx, userID, name)
	return args.Error(0)
}

func (m *MockRepository) Disable(ctx context.Context, userID uint64, name string) (apikey.APIKey, error) {
	args := m.Called(ctx, userID, name)
	return args.Get(0).(apikey.APIKey), args.Error(1)
}

func (m *MockRepository) Enable(ctx context.Context, userID uint64, name string) (apikey.APIKey, error) {
	args := m.Called(ctx, userID, name)
	return args.Get(0).(apikey.APIKey), args.Error(1)
}

func (m *MockRepository) GetAll(ctx context.Context, userID uint64, name *string) ([]apikey.APIKey, error) {
	args := m.Called(ctx, userID, name)
	return args.Get(0).([]apikey.APIKey), args.Error(1)
}

func (m *MockRepository) UpdateLastUsedAt(ctx context.Context, key apikey.HashedKeyCredential) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockRepository) Get(ctx context.Context, key apikey.HashedKeyCredential) (apikey.APIKey, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(apikey.APIKey), args.Error(1)
}
