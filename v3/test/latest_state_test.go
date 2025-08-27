package integration_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gobitfly/beaconchain-backend/test/testUtils"
	"github.com/stretchr/testify/assert"
)

func TestLatestStateEndpoint(t *testing.T) {
	if !testUtils.HasTags("smoke", "e2e") {
		t.Skip("❌ Skipping test: requires tag 'smoke' or 'e2e'")
	}

	if ok, reason := testUtils.IsValidURL(testUtils.GetExternalHTTPUrl()); !ok {
		t.Skipf("⚠️ Skipping test due to invalid BASE_URL: %s", reason)
	}

	assert := assert.New(t)

	url := testUtils.GetExternalHTTPUrl() + "/api/v1/latestState"
	t.Logf("➡️  Requesting: %s", url)

	resp, err := http.Get(url) //nolint:gosec // trusted internal URL
	if err != nil {
		t.Fatalf("❌ Request failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("❌ Failed to close response body: %v", err)
		}
	}()

	t.Logf("✅ Status: %d", resp.StatusCode)
	assert.Equal(http.StatusOK, resp.StatusCode, "Expected status code 200")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(err, "Failed to read response body")

	t.Logf("📦 Raw Body: %s", string(body))

	var response struct {
		LastProposedSlot      int64 `json:"lastProposedSlot"`
		CurrentSlot           int64 `json:"currentSlot"`
		CurrentEpoch          int64 `json:"currentEpoch"`
		CurrentFinalizedEpoch int64 `json:"currentFinalizedEpoch"`
		FinalityDelay         int   `json:"finalityDelay"`
		Syncing               bool  `json:"syncing"`
		Rates                 struct {
			MainCurrencyTickerPrice float64 `json:"mainCurrencyTickerPrice"`
		} `json:"rates"`
	}

	err = json.Unmarshal(body, &response)
	assert.NoError(err, "Failed to parse JSON")

	assert.Greater(response.LastProposedSlot, int64(0), "lastProposedSlot should be > 0")
	assert.Equal(response.CurrentSlot, response.LastProposedSlot, "currentSlot should equal lastProposedSlot")
	assert.Greater(response.CurrentEpoch, int64(0), "currentEpoch should be > 0")
	assert.Greater(response.CurrentFinalizedEpoch, int64(0), "currentFinalizedEpoch should be > 0")
	assert.GreaterOrEqual(response.FinalityDelay, 0, "finalityDelay should be >= 0")

	t.Logf("🔄 Syncing: %v", response.Syncing)
	assert.Greater(response.Rates.MainCurrencyTickerPrice, float64(0), "mainCurrencyTickerPrice should be > 0")
}
