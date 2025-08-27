package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/gobitfly/beaconchain-backend/test/testUtils"
	"github.com/stretchr/testify/assert"
)

func TestE2EValidatorLifecycle(t *testing.T) {
	if !testUtils.HasTags("v2-beta-gnosis", "v2-beta-mainnet") {
		t.Skip("❌ Skipping test: requires tag 'v2-beta-gnosis' or 'v2-beta-mainnet'")
	}

	if ok, reason := testUtils.IsValidURL(testUtils.GetExternalHTTPUrl()); !ok {
		t.Skipf("⚠️ Skipping test due to invalid BASE_URL: %s", reason)
	}

	apiKey := os.Getenv("API_KEY_ORCA_TEST")
	if apiKey == "" {
		t.Skip("⚠️ Skipping test: API_KEY_ORCA_TEST is not set in environment variables")
	}

	assert := assert.New(t)
	baseURL := testUtils.GetExternalHTTPUrl()
	dashboardID := testUtils.GetDashboardID()
	groupID := 0
	client := &http.Client{}
	expectedIndices := []int{21, 3455, 654, 23, 4, 5, 6, 87, 465, 243546, 3343}
	t.Logf("🔑 API Key (first 5 chars): %s...", apiKey[:5])

	addValidators := func() {
		addURL := fmt.Sprintf("%s/api/v2/validator-dashboards/%d/validators?api_key=%s", baseURL, dashboardID, apiKey)
		t.Logf("➕ POST to: %s", addURL)

		requestBody := map[string]any{
			"deposit_address":       "",
			"graffiti":              "",
			"group_id":              groupID,
			"validators":            expectedIndices,
			"withdrawal_credential": "",
		}

		jsonBody, err := json.Marshal(requestBody)
		assert.NoError(err)

		addReq, err := http.NewRequest("POST", addURL, bytes.NewBuffer(jsonBody))
		assert.NoError(err)
		addReq.Header.Set("Content-Type", "application/json")
		addReq.Header.Set("Accept", "application/json")

		addResp, err := client.Do(addReq)
		if err != nil {
			t.Fatalf("❌ Request failed: %v", err)
		}
		defer func() {
			if err := addResp.Body.Close(); err != nil {
				t.Logf("❌ Failed to close addResp body: %v", err)
			}
		}()

		addRespBody, err := io.ReadAll(addResp.Body)
		assert.NoError(err)

		t.Logf("📦 Add response: %s", string(addRespBody))
		assert.Equal(http.StatusCreated, addResp.StatusCode)

		var addResponse struct {
			Data []struct {
				Index   int `json:"index"`
				GroupID int `json:"group_id"`
			} `json:"data"`
		}
		err = json.Unmarshal(addRespBody, &addResponse)
		assert.NoError(err)

		actualIndices := make(map[int]bool)
		for _, item := range addResponse.Data {
			assert.Equal(groupID, item.GroupID)
			actualIndices[item.Index] = true
		}
		for _, expected := range expectedIndices {
			assert.True(actualIndices[expected], fmt.Sprintf("❌ Expected validator index %d to be in response", expected))
		}
	}

	deleteAll := func() {
		deleteURL := fmt.Sprintf("%s/api/v2/validator-dashboards/%d/groups/%d/validators?api_key=%s", baseURL, dashboardID, groupID, apiKey)
		t.Logf("🗑️ DELETE ALL: %s", deleteURL)

		delReq, err := http.NewRequest("DELETE", deleteURL, nil)
		assert.NoError(err)
		delReq.Header.Set("Accept", "application/json")

		delResp, err := client.Do(delReq)
		if err != nil {
			t.Fatalf("❌ Request failed: %v", err)
		}
		defer func() {
			if err := delResp.Body.Close(); err != nil {
				t.Logf("❌ Failed to close delResp body: %v", err)
			}
		}()

		t.Logf("📦 Delete response status: %d", delResp.StatusCode)
		assert.Equal(http.StatusNoContent, delResp.StatusCode)
	}

	// Run lifecycle: Add → Delete All → Re-add → Bulk Delete → Final Delete
	addValidators()
	deleteAll()
	addValidators()

	// Bulk delete two validators
	bulkDeleteURL := fmt.Sprintf("%s/api/v2/validator-dashboards/%d/validators/bulk-deletions?api_key=%s", baseURL, dashboardID, apiKey)
	t.Logf("🗑️ BULK DELETE: %s", bulkDeleteURL)

	bulkBody := map[string]any{
		"validators": []int{87, 465},
	}
	jsonBulk, err := json.Marshal(bulkBody)
	assert.NoError(err)

	bulkReq, err := http.NewRequest("POST", bulkDeleteURL, bytes.NewBuffer(jsonBulk))
	assert.NoError(err)
	bulkReq.Header.Set("Content-Type", "application/json")
	bulkReq.Header.Set("Accept", "application/json")

	bulkResp, err := client.Do(bulkReq)
	if err != nil {
		t.Fatalf("❌ Request failed: %v", err)
	}
	defer func() {
		if err := bulkResp.Body.Close(); err != nil {
			t.Logf("❌ Failed to close bulkResp body: %v", err)
		}
	}()

	bulkRespBody, err := io.ReadAll(bulkResp.Body)
	assert.NoError(err)

	t.Logf("📦 Bulk delete response: %s", string(bulkRespBody))
	assert.Equal(http.StatusNoContent, bulkResp.StatusCode)

	deleteAll()

	t.Log("✅ E2E validator lifecycle complete: add → delete all → re-add → bulk delete → final delete all")
}
