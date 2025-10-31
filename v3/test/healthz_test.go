package integration_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gobitfly/beaconchain-backend/test/testUtils"
	"github.com/stretchr/testify/assert"
)

func TestHealthzEndpoint(t *testing.T) {
	if !testUtils.HasTags("smoke", "e2e") {
		t.Skip("❌ Skipping test: requires tag 'e2e' or 'smoke'")
	}

	if ok, reason := testUtils.IsValidURL(testUtils.GetExternalHTTPUrl()); !ok {
		t.Skipf("⚠️ Skipping test due to invalid BASE_URL: %s", reason)
	}

	baseURL := testUtils.GetExternalHTTPUrl()
	url := baseURL + "/api/healthz"
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

	assert := assert.New(t)

	t.Logf("✅ Status: %d", resp.StatusCode)
	assert.Equal(http.StatusOK, resp.StatusCode, "Expected status code 200")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(err, "Failed to read response body")

	bodyStr := strings.ReplaceAll(string(body), "\r\n", "\n")
	t.Logf("📦 Body: %s", bodyStr)

	assert.NotEmpty(bodyStr, "Response body should not be empty")
	assert.Contains(bodyStr, "monitoring_api", "Expected 'monitoring_api' in response")

	expectedModules := []string{
		"module monitoring_api: OK",
		"module monitoring_cl_data: OK",
		"module monitoring_services: OK",
		"module monitoring_el_data: OK",
		"module monitoring_app: OK",
		"module monitoring_redis: OK",
	}

	for _, expected := range expectedModules {
		assert.Contains(bodyStr, expected, "Expected module output not found: "+expected)
	}
}
