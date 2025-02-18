package dataaccess

import "context"

type User struct {
	userId  int
	isAdmin bool
}

type UserRepository interface {
	// Creates a new user
	CreateUser(ctx context.Context) (*User, error)

	/**
	 * Creates an empty Validator Dashboard
	 */
	GetUserById(ctx context.Context, userId int) (*User, error)

	/**
	 * Modifies the attributes of a dashboard.
	 * If an attribute is not included or is nil in the model, it should not be updated.
	 * If an attribute is included but modification of that is not possible, an error should be returned
	 *    and other attribute modifications should not take place.
	 */
	ModifyUser(ctx context.Context, userId int, isAdmin bool) (*User, error)

	// Returns nothing on success, or error if successfully deleted. Idempotent, if deleted when it DNE, no error should be returned.
	DeleteUser(ctx context.Context, userId int) error
}
