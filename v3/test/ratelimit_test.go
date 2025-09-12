package integration

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	ctx, client := setupExternalAPIClient(t)
	// give a timeout to avoid hanging tests
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	const numRequests = 100
	var wg sync.WaitGroup
	results := make(chan *http.Response, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := client.GetPing(ctx)
			assert.Nil(t, resp.Body.Close())
			if err != nil {
				results <- &http.Response{StatusCode: http.StatusInternalServerError}
				return
			}
			results <- resp
		}()
	}

	wg.Wait()
	close(results)

	var (
		successSeen  bool
		rateLimitHit bool
	)

	for resp := range results {
		if resp.StatusCode == http.StatusOK {
			successSeen = true
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitHit = true
			continue
		}

	}

	assert.True(t, successSeen, "expected at least one successful request before rate limit is hit")
	assert.True(t, rateLimitHit, "expected at least one rate limited request")
}
