package common

import (
	"errors"

	"github.com/gobitfly/beaconchain-backend/internal/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const GenericErrMsg = "internal server error. please try again later."

// ExternalError is a trusted, user-facing error type
// use status.Errorf instead if you want an error that should not be exposed to the user and catched in the middleware.
type ExternalError struct {
	Message string
	Code    codes.Code
}

func (e *ExternalError) Error() string {
	return e.Message
}

func (e *ExternalError) GRPCStatus() *status.Status {
	return status.New(e.Code, e.Message)
}

func NewExternalError(code codes.Code, message string) error {
	return &ExternalError{
		Message: message,
		Code:    code,
	}
}

func SanitizeErrorMessage(err error) error {
	var external *ExternalError
	if errors.As(err, &external) {
		return external.GRPCStatus().Err()
	}

	// Otherwise, sanitize the error
	st, ok := status.FromError(err)
	if st.Code() == codes.Internal {
		log.Infof("internal error: %v", err)
	}
	if !ok {
		return status.Error(codes.Internal, GenericErrMsg)
	}

	return status.Error(st.Code(), genericMessageForCode(st.Code()))
}

func genericMessageForCode(code codes.Code) string {
	switch code {
	case codes.NotFound:
		return "resource not found"
	case codes.PermissionDenied:
		return "permission denied"
	case codes.AlreadyExists:
		return "resource already exists"
	case codes.InvalidArgument:
		return "invalid request parameters"
	case codes.Unauthenticated:
		return "unauthenticated request"
	default:
		return GenericErrMsg
	}
}
