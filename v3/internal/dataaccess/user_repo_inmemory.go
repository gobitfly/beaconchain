package dataaccess

import (
	"context"
	"errors"
)

type InMemoryUserRepository struct {
	userDatabase map[int]*User
	lastId       int
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		userDatabase: make(map[int]*User),
		lastId:       0,
	}
}

func (r *InMemoryUserRepository) CreateUser(ctx context.Context) (*User, error) {
	r.lastId = r.lastId + 1
	newUser := User{userId: r.lastId, isAdmin: false}
	r.userDatabase[newUser.userId] = &newUser
	return &newUser, nil
}

/**
 * Creates an empty Validator Dashboard
 */
func (r *InMemoryUserRepository) GetUserById(ctx context.Context, userId int) (*User, error) {
	user := r.userDatabase[userId]
	if user == nil {
		return nil, errors.New("No user found with the provided ID")
	}
	return user, nil
}

/**
 * Modifies the attributes of a dashboard.
 * If an attribute is not included or is nil in the model, it should not be updated.
 * If an attribute is included but modification of that is not possible, an error should be returned
 *    and other attribute modifications should not take place.
 */
func (r *InMemoryUserRepository) ModifyUser(ctx context.Context, userId int, isAdmin bool) (*User, error) {
	user := r.userDatabase[userId]
	user.isAdmin = isAdmin
	return user, nil
}

// Returns nothing on success, or error if successfully deleted. Idempotent, if deleted when it DNE, no error should be returned.
func (r *InMemoryUserRepository) DeleteUser(ctx context.Context, userId int) error {
	// Do nothing, pretend it was deleted
	delete(r.userDatabase, userId)
	return nil
}
