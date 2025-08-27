package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRateLimit(t *testing.T) {
	ctx, client := setupExternalAPIClient(t)
	in := &model.ExecutionBlockRequest{BlockNumber: "1"}

	const numRequests = 100
	var wg sync.WaitGroup
	results := make(chan error, numRequests)

	// give a timeout to avoid hanging tests
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.ExecutionBlock(ctx, in)
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	var (
		successSeen  bool
		rateLimitHit bool
	)

	for err := range results {
		if err == nil {
			successSeen = true
			continue
		}
		if status.Code(err) == codes.ResourceExhausted {
			rateLimitHit = true
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	assert.True(t, successSeen, "expected at least one successful request before rate limit is hit")
	assert.True(t, rateLimitHit, "expected at least one rate limited request")
}
