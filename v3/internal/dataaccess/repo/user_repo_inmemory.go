package dataaccess

import (
	"context"
	"errors"
)

type InMemoryUserRepository struct {
	userDatabase map[uint64]*User
	lastId       uint64
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		userDatabase: make(map[uint64]*User),
		lastId:       0,
	}
}

func (r *InMemoryUserRepository) CreateUser(ctx context.Context, email string, initialApiKey string, hashedPassword string) (*User, error) {
	r.lastId = r.lastId + 1
	newUser := User{id: r.lastId, isAdmin: false}
	r.userDatabase[newUser.id] = &newUser
	return &newUser, nil
}

/**
 * Creates an empty Validator Dashboard
 */
func (r *InMemoryUserRepository) GetUserById(ctx context.Context, userId uint64) (*User, error) {
	user := r.userDatabase[userId]
	if user == nil {
		return nil, errors.New("No user found with the provided ID")
	}
	return user, nil
}

func (r *InMemoryUserRepository) GetUserByApiKey(ctx context.Context, apikey string) (*User, error) {
	return nil, errors.New("Unimplemented")
}

/**
 * Modifies the attributes of a dashboard.
 * If an attribute is not included or is nil in the model, it should not be updated.
 * If an attribute is included but modification of that is not possible, an error should be returned
 *    and other attribute modifications should not take place.
 */
func (r *InMemoryUserRepository) ModifyUser(ctx context.Context, userId uint64, isAdmin bool) (*User, error) {
	user := r.userDatabase[userId]
	user.isAdmin = isAdmin
	return user, nil
}

// Returns nothing on success, or error if successfully deleted. Idempotent, if deleted when it DNE, no error should be returned.
func (r *InMemoryUserRepository) DeleteUser(ctx context.Context, userId uint64) error {
	// Do nothing, pretend it was deleted
	delete(r.userDatabase, userId)
	return nil
}
