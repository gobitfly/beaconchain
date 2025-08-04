package dataaccess

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockUserRepository) GetUserById(ctx context.Context, id uint64) (*domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, email string, initialApiKey string, hashedPassword string) (*domain.User, error) {
	args := m.Called(ctx, email, initialApiKey, hashedPassword)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
