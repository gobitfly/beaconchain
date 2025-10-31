package integration_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gobitfly/beaconchain-backend/test/testUtils"
	"github.com/stretchr/testify/assert"
)

func TestValidatorsQueueEndpoint(t *testing.T) {
	if !testUtils.HasTags("e2e", "smoke") {
		t.Skip("❌ Skipping test: requires tag 'e2e' or 'smoke'")
	}

	if ok, reason := testUtils.IsValidURL(testUtils.GetExternalHTTPUrl()); !ok {
		t.Skipf("⚠️ Skipping test due to invalid BASE_URL: %s", reason)
	}

	assert := assert.New(t)

	url := testUtils.GetExternalHTTPUrl() + "/api/v1/validators/queue"
	t.Logf("➡️ Requesting: %s", url)

	resp, err := http.Get(url) //nolint:gosec // trusted internal URL
	if err != nil {
		t.Fatalf("❌ Request failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("❌ Failed to close response body: %v", err)
		}
	}()

	t.Logf("✅ Status code: %d", resp.StatusCode)
	assert.Equal(http.StatusOK, resp.StatusCode, "Expected status code 200")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(err, "Failed to read response body")

	t.Logf("📦 Raw response: %s", string(body))

	var result struct {
		Status string `json:"status"`
		Data   struct {
			BeaconchainEntering        int64 `json:"beaconchain_entering"`
			BeaconchainExiting         int64 `json:"beaconchain_exiting"`
			ValidatorsCount            int64 `json:"validatorscount"`
			BeaconchainEnteringBalance int64 `json:"beaconchain_entering_balance"`
			BeaconchainExitingBalance  int64 `json:"beaconchain_exiting_balance"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &result)
	assert.NoError(err, "Failed to parse JSON response")

	assert.Equal("OK", result.Status, "Expected status to be OK")
	assert.GreaterOrEqual(result.Data.BeaconchainEntering, int64(0), "beaconchain_entering should be >= 0")
	assert.GreaterOrEqual(result.Data.BeaconchainExiting, int64(0), "beaconchain_exiting should be >= 0")
	assert.Greater(result.Data.ValidatorsCount, int64(0), "validatorscount should be > 0")
	assert.GreaterOrEqual(result.Data.BeaconchainEnteringBalance, int64(0), "beaconchain_entering_balance should be >= 0")
	assert.GreaterOrEqual(result.Data.BeaconchainExitingBalance, int64(0), "beaconchain_exiting_balance should be >= 0")
}
