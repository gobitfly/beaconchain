package userrepo

import (
	"context"

	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRepository) Get(ctx context.Context, id uint64) (domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockRepository) Create(ctx context.Context, email string, initialApiKey string, hashedPassword string) (domain.User, error) {
	args := m.Called(ctx, email, initialApiKey, hashedPassword)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockRepository) Delete(ctx context.Context, id uint64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
