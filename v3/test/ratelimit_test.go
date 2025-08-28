package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/gobitfly/beaconchain-backend/api/gen/client/client/external_service"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	ctx, client := setupExternalAPIClient(t)
	// give a timeout to avoid hanging tests
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	in := &external_service.ExternalServiceExecutionBlockParams{BlockNumber: "1", Context: ctx}

	const numRequests = 100
	var wg sync.WaitGroup
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.ExternalServiceExecutionBlock(in, nil)
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
		if apiErr, ok := err.(*external_service.ExternalServiceExecutionBlockDefault); ok {
			if apiErr.Code() == 429 {
				rateLimitHit = true
				continue
			}
		}
	}

	assert.True(t, successSeen, "expected at least one successful request before rate limit is hit")
	assert.True(t, rateLimitHit, "expected at least one rate limited request")
}
