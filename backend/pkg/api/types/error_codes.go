package types

type ErrorCode string

const (
	ErrorEffectiveBalanceTooHigh ErrorCode = "MAX_EB_EXCEEDED"
	ErrorBadRequest              ErrorCode = "BAD_REQUEST"
	ErrorInternalServerError     ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrorNotFound                ErrorCode = "NOT_FOUND"
	ErrorUnauthorized            ErrorCode = "UNAUTHORIZED"
	ErrorForbidden               ErrorCode = "FORBIDDEN"
	ErrorConflict                ErrorCode = "CONFLICT"
	ErrorTooManyRequests         ErrorCode = "TOO_MANY_REQUESTS"
	ErrorGone                    ErrorCode = "GONE"
	ErrorServiceUnavailable      ErrorCode = "SERVICE_UNAVAILABLE"
)
