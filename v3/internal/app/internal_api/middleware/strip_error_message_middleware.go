package middleware

import (
	"context"
	"errors"

	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func StripErrorMessageMiddleware() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		return nil, sanitizeErrorMessage(err)
	}
}

func sanitizeErrorMessage(err error) error {
	var external *common.InternalUserFacingError
	if errors.As(err, &external) {
		return external.GRPCStatus().Err()
	}

	// Otherwise, sanitize the error
	st, ok := status.FromError(err)
	if st.Code() == codes.Internal {
		log.Infof("internal error: %v", err)
	}
	if !ok {
		return status.Error(codes.Internal, common.GenericErrMsg)
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
		return common.GenericErrMsg
	}
}
