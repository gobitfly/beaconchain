package common

import "net/http"

const GenericErrMsg = "internal server error. please try again later."

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

func Code(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Status
	}
	return http.StatusInternalServerError
}
