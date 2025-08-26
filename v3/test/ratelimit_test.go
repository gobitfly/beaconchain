package integration

import (
	"testing"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRateLimit(t *testing.T) {
	ctx, client := setupExternalApiClient(t)
	in := &model.ExecutionBlockRequest{
		BlockNumber: "1",
	}
	hasOneSuccess := false
	for range 10000 {
		_, err := client.ExecutionBlock(ctx, in)
		if err == nil {
			hasOneSuccess = true
			continue
		}
		assert.Equal(t, status.Code(err), codes.ResourceExhausted)
		assert.True(t, hasOneSuccess, "expected at least one successful request before rate limit is hit")
		break
	}
}
