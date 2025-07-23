package domain

import (
	"errors"
)

// Domain errors
var (
	ErrDuplicate = errors.New("duplicate key entry")
	ErrNotFound  = errors.New("not found")
)
