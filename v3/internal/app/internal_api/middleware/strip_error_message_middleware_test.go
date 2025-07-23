package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStripErrorMessageInterceptor(t *testing.T) {
	tests := []struct {
		name         string
		inputError   error
		expectedCode codes.Code
		expectedMsg  string
	}{
		{
			name:         "internal error stripped",
			inputError:   status.Errorf(codes.Internal, "failed to connect to DB"),
			expectedCode: codes.Internal,
			expectedMsg:  "internal server error",
		},
		{
			name:         "not found error stripped",
			inputError:   status.Errorf(codes.NotFound, "record with ID X not found"),
			expectedCode: codes.NotFound,
			expectedMsg:  "resource not found",
		},
		{
			name:         "already exists error stripped",
			inputError:   status.Errorf(codes.AlreadyExists, "key already exists"),
			expectedCode: codes.AlreadyExists,
			expectedMsg:  "resource already exists",
		},
		{
			name:         "non-status error becomes internal",
			inputError:   context.DeadlineExceeded,
			expectedCode: codes.Internal,
			expectedMsg:  "internal server error. please try again later.",
		},
	}

	interceptor := StripErrorMessageMiddleware()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, tt.inputError
			}

			_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
				FullMethod: "/test.Service/Method",
			}, handler)

			st, ok := status.FromError(err)
			assert.True(t, ok, "expected error to be a gRPC status")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Equal(t, tt.expectedMsg, st.Message())
		})
	}
}
