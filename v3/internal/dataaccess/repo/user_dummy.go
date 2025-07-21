package dataaccess

import (
	"context"
)

type DummyUserRepository struct{}

func (r *DummyUserRepository) Ping() error {
	return nil
}

// TODO could return proper (semi-)random dummy data
var dummyUserResponse = User{id: 1, isAdmin: false}

func (r *DummyUserRepository) GetUserById(ctx context.Context, id uint64) (*User, error) {
	return &dummyUserResponse, nil
}

func (r *DummyUserRepository) GetUserByApiKey(ctx context.Context, apikey string) (*User, error) {
	return &dummyUserResponse, nil
}

func (r *DummyUserRepository) CreateUser(ctx context.Context, email string, initialApiKey string, hashedPassword string) (*User, error) {
	return &dummyUserResponse, nil
}

func (r *DummyUserRepository) DeleteUser(ctx context.Context, id uint64) error {
	return nil
}
