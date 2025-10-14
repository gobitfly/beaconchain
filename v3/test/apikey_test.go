package integration

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gobitfly/beaconchain-backend/api/external/client"
	inhouseclient "github.com/gobitfly/beaconchain-backend/api/inhouse/client"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/gobitfly/beaconchain-backend/internal/limits"
	"github.com/gobitfly/beaconchain-backend/test/testUtils"
	"github.com/stretchr/testify/assert"
)

type apiKeyMgmtTest struct {
	testRunID string
}

func newAPIKeyMgmt() *apiKeyMgmtTest {
	return &apiKeyMgmtTest{testRunID: "keymgmt"}
}

var apiKeyMgmt = newAPIKeyMgmt()

func TestAPIKeyLifecycle(t *testing.T) {
	var testAPIKeyLifecycle = apiKeyMgmt.newTestKey("lifecycle")
	var key *inhouseclient.CreateAPIKeyResponse

	apiKeyMgmt.withCleanState(t, func(ctx context.Context, client *inhouseclient.ClientWithResponses) {

		// Creation
		t.Run("create key", func(t *testing.T) {
			var err error
			key, err = client.CreateAPIKeyWithResponse(ctx, inhouseclient.CreateAPIKeyJSONRequestBody{Name: testAPIKeyLifecycle})
			assert.NoError(t, err)
		})

		t.Run("get after creation", func(t *testing.T) {
			got := apiKeyMgmt.mustGetKey(t, client, testAPIKeyLifecycle)
			assert.Equal(t, testAPIKeyLifecycle, got.JSON200.ApiKey.Name)
			assert.Nil(t, got.JSON200.ApiKey.LastUsedAt)
			assert.Nil(t, got.JSON200.ApiKey.DisabledAt)
			assert.NotNil(t, got.JSON200.ApiKey.CreatedAt)
		})

		// Usage

		var lastUsedAt time.Time

		t.Run("use key (no cache hit)", func(t *testing.T) {
			ctx, extClient := setupExternalAPIClientWithAuth(*key.JSON200.RawApiKey)
			resp, err := extClient.GetPing(ctx)
			assert.Nil(t, resp.Body.Close())
			assert.NoError(t, err)
		})

		t.Run("last used timestamp updated", func(t *testing.T) {
			got := apiKeyMgmt.mustGetKey(t, client, testAPIKeyLifecycle)
			assert.NotNil(t, got.JSON200.ApiKey.LastUsedAt)
			lastUsedAt = *got.JSON200.ApiKey.LastUsedAt
		})

		time.Sleep(1000 * time.Millisecond) // wait a bit to ensure next last used timestamp is different & not trip over ratelimit

		t.Run("use key (cache hit)", func(t *testing.T) {
			ctx, extClient := setupExternalAPIClientWithAuth(*key.JSON200.RawApiKey)
			resp, err := extClient.GetPing(ctx)
			assert.Nil(t, resp.Body.Close())
			assert.NoError(t, err)
		})

		t.Run("last used timestamp is updated and chronological", func(t *testing.T) {
			got := apiKeyMgmt.mustGetKey(t, client, testAPIKeyLifecycle)
			assert.NotNil(t, got.JSON200.ApiKey.LastUsedAt)
			assert.True(t, got.JSON200.ApiKey.LastUsedAt.After(lastUsedAt), "expected last used timestamp to be updated in chronological order (old: %v, new: %v)", lastUsedAt, got.JSON200.ApiKey.LastUsedAt)
		})

		// Disable

		t.Run("disable key", func(t *testing.T) {
			_, err := client.DisableAPIKeyWithResponse(ctx, testAPIKeyLifecycle)
			assert.NoError(t, err)

			got := apiKeyMgmt.mustGetKey(t, client, testAPIKeyLifecycle)
			assert.NotNil(t, got.JSON200.ApiKey.DisabledAt)
		})

		t.Run("disabled key cannot be used", func(t *testing.T) {
			ctx, extClient := setupExternalAPIClientWithAuth(*key.JSON200.RawApiKey)
			resp, err := extClient.GetPing(ctx)
			assert.Nil(t, resp.Body.Close())
			assert.Nil(t, err)
			assert.Equal(t, 401, resp.StatusCode)
		})

		t.Run("disabled key timestamp not changing after disabling again", func(t *testing.T) {
			key := apiKeyMgmt.mustGetKey(t, client, testAPIKeyLifecycle)
			assert.NotNil(t, key.JSON200.ApiKey.DisabledAt)

			keyDisabledAgain, err := client.DisableAPIKeyWithResponse(ctx, testAPIKeyLifecycle)
			assert.NoError(t, err)

			assert.Equal(t, key.JSON200.ApiKey.DisabledAt, keyDisabledAgain.JSON200.ApiKey.DisabledAt, "DisabledAt timestamp should not change when disabling an already disabled key")
		})

		// Enable
		t.Run("enable key", func(t *testing.T) {
			_, err := client.EnableAPIKeyWithResponse(ctx, testAPIKeyLifecycle)
			assert.NoError(t, err)

			got := apiKeyMgmt.mustGetKey(t, client, testAPIKeyLifecycle)
			assert.Nil(t, got.JSON200.ApiKey.DisabledAt)
		})

		t.Run("enabled key can be used again", func(t *testing.T) {
			ctx, extClient := setupExternalAPIClientWithAuth(*key.JSON200.RawApiKey)
			resp, err := extClient.GetPing(ctx)
			assert.Nil(t, resp.Body.Close())
			assert.NoError(t, err)
		})

		t.Run("enabling an enabled key should just return ok", func(t *testing.T) {
			_, err := client.EnableAPIKeyWithResponse(ctx, testAPIKeyLifecycle)
			assert.NoError(t, err)
		})

		// Deletion

		t.Run("delete key", func(t *testing.T) {
			res, err := client.DeleteAPIKeyWithResponse(ctx, testAPIKeyLifecycle)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, res.StatusCode())
		})

		t.Run("deleted key cannot be used", func(t *testing.T) {
			ctx, extClient := setupExternalAPIClientWithAuth(*key.JSON200.RawApiKey)
			resp, err := extClient.GetPing(ctx)
			assert.Nil(t, resp.Body.Close())
			assert.Nil(t, err)
			assert.Equal(t, 401, resp.StatusCode)
		})

	})
}

func TestAPIKeyCreationMaxLimit(t *testing.T) {
	var testAPIKeyMaxLimit = apiKeyMgmt.newTestKey("max-key")

	limiter := limits.NewLimiter()
	apiKeyMgmt.withCleanState(t, func(ctx context.Context, client *inhouseclient.ClientWithResponses) {
		tierMaxLimit, err := limiter.GetMaxAPIKeys(ctx, domain.User{ID: 1, SubscriptionTier: domain.TierBusiness})
		assert.NoError(t, err)

		var hasReachedLimit bool
		for i := 0; i < tierMaxLimit+1; i++ { // one above tier limit
			name := fmt.Sprintf("%s-%d", testAPIKeyMaxLimit, i)
			res, _ := client.CreateAPIKeyWithResponse(ctx, inhouseclient.CreateAPIKeyJSONRequestBody{Name: name})
			if res.StatusCode() == http.StatusForbidden {
				hasReachedLimit = true
				break
			}
		}
		assert.True(t, hasReachedLimit, "expected to eventually hit resource limit")
	})
}

func TestAPIKeyList(t *testing.T) {
	var testAPIKeyList = apiKeyMgmt.newTestKey("list-key")

	apiKeyMgmt.withCleanState(t, func(ctx context.Context, client *inhouseclient.ClientWithResponses) {
		// Create some keys
		keysToCreate := []string{fmt.Sprintf("%s-1", testAPIKeyList), fmt.Sprintf("%s-2", testAPIKeyList), fmt.Sprintf("%s-3", testAPIKeyList)}
		for _, name := range keysToCreate {
			_, err := client.CreateAPIKeyWithResponse(ctx, inhouseclient.CreateAPIKeyJSONRequestBody{Name: name})
			assert.NoError(t, err)
		}

		t.Run("list keys", func(t *testing.T) {
			resp, err := client.GetAPIKeysWithResponse(ctx)
			assert.NoError(t, err)
			foundKeys := make(map[string]bool)
			for _, k := range resp.JSON200.ApiKeys {
				foundKeys[k.Name] = true
			}
			for _, name := range keysToCreate {
				assert.True(t, foundKeys[name], "expected to find key %s in list", name)
			}
		})
	})
}

func TestAPIKeyInvalidUsages(t *testing.T) {
	var testAPIKeyInvalid = apiKeyMgmt.newTestKey("invalid")
	t.Run("invalid key cannot be used", func(t *testing.T) {
		ctx, extClient := setupExternalAPIClientWithAuth(testAPIKeyInvalid)
		resp, err := extClient.GetPing(ctx)
		assert.Nil(t, resp.Body.Close())
		assert.Nil(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("usage without key", func(t *testing.T) {
		cl, _ := client.NewClient(fmt.Sprintf("http://%s", testUtils.GetExternalHTTPUrl()))
		resp, err := cl.GetPing(context.Background())
		assert.Nil(t, resp.Body.Close())
		assert.Nil(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	})
}

// Todo: Make sure to add tests trying to manage keys as a different user once internal auth is in place

// --- Helpers ---

func (*apiKeyMgmtTest) mustGetKey(t *testing.T, client *inhouseclient.ClientWithResponses, name string) *inhouseclient.GetAPIKeyResponse {
	t.Helper()
	k, err := client.GetAPIKeyWithResponse(context.Background(), name)
	assert.NoError(t, err, "failed to get key %s", name)
	return k
}

func (*apiKeyMgmtTest) cleanState(t *testing.T, client *inhouseclient.ClientWithResponses, ctx context.Context) {
	keys, err := client.GetAPIKeysWithResponse(ctx)
	if err != nil {
		t.Fatal("failed to list API keys for cleanup:", err)
	}
	if keys.JSON200 == nil || keys.JSON200.ApiKeys == nil {
		return
	}
	for _, k := range keys.JSON200.ApiKeys {
		if !strings.HasPrefix(k.Name, apiKeyMgmt.testRunID) { // scoped to cleanup just the keys of this test suite
			continue
		}
		_, err := client.DeleteAPIKeyWithResponse(ctx, k.Name)
		if err != nil {
			t.Fatal("failed to delete API key:", err)
		}
	}
}

func (apiKeyMgmt *apiKeyMgmtTest) newTestKey(name string) string {
	return fmt.Sprintf("%s-%s", apiKeyMgmt.testRunID, name) // prefix key to know what to cleanup
}

func (apiKeyMgmt *apiKeyMgmtTest) withCleanState(t *testing.T, test func(ctx context.Context, client *inhouseclient.ClientWithResponses)) {
	t.Helper()
	ctx, client := setupInhouseAPIClient(t)

	apiKeyMgmt.cleanState(t, client, ctx)

	test(ctx, client)

	t.Cleanup(func() {
		apiKeyMgmt.cleanState(t, client, ctx)
	})
}
