package common

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const GenericErrMsg = "internal server error. please try again later."

// ==== Internal API Errors ====

// InternalUserFacingError is a trusted, user-facing error type
// use status.Errorf instead if you want an error that should not be exposed to the user and catched in the middleware.
type InternalUserFacingError struct {
	Message string
	Code    codes.Code
}

func (e *InternalUserFacingError) Error() string {
	return e.Message
}

func (e *InternalUserFacingError) GRPCStatus() *status.Status {
	return status.New(e.Code, e.Message)
}

func NewInternalUserFacingError(code codes.Code, message string) error {
	return &InternalUserFacingError{
		Message: message,
		Code:    code,
	}
}

// ===== External API Errors =====

// ErrorVisibility indicates whether an error is safe to show to users.
type ErrorVisibility string

const (
	ErrorVisibilityUser     ErrorVisibility = "user"     // safe to show, user facing
	ErrorVisibilityInternal ErrorVisibility = "internal" // log only, replace with generic for users
)

type APIError struct {
	Status     int                    `json:"status"`
	Message    string                 `json:"message"`
	Visibility ErrorVisibility        `json:"-"`
	Extras     map[string]interface{} `json:"extras,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewAPIUserFacingError(status int, msg string) *APIError {
	return &APIError{Status: status, Message: msg, Visibility: ErrorVisibilityUser, Extras: nil}
}

func NewAPIInternalError(status int, msg string) *APIError {
	return &APIError{Status: status, Message: msg, Visibility: ErrorVisibilityInternal, Extras: nil}
}

func NewAPIError(status int, vis ErrorVisibility, msg string) *APIError {
	return &APIError{Status: status, Message: msg, Visibility: vis, Extras: nil}
}

func NewAPIErrorWithExtras(status int, vis ErrorVisibility, msg string, extras map[string]interface{}) *APIError {
	return &APIError{Status: status, Message: msg, Visibility: vis, Extras: extras}
}
