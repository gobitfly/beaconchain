package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/api/gen/client/client"
	"github.com/gobitfly/beaconchain-backend/api/gen/client/client/external_service"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/gobitfly/beaconchain-backend/internal/limits"
	"github.com/gobitfly/beaconchain-backend/test/testUtils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	var key *model.CreateAPIKeyResponse

	apiKeyMgmt.withCleanState(t, func(ctx context.Context, client model.InternalServiceClient) {

		// Creation
		t.Run("create key", func(t *testing.T) {
			var err error
			key, err = client.CreateAPIKey(ctx, &model.CreateAPIKeyRequest{Name: testAPIKeyLifecycle})
			assert.NoError(t, err)
		})

		t.Run("get after creation", func(t *testing.T) {
			got := apiKeyMgmt.mustGetKey(t, client, testAPIKeyLifecycle)
			assert.Equal(t, testAPIKeyLifecycle, got.ApiKey.Name)
			assert.Nil(t, got.ApiKey.LastUsedAt)
			assert.Nil(t, got.ApiKey.DisabledAt)
			assert.NotNil(t, got.ApiKey.CreatedAt)
		})

		// Usage

		var lastUsedAt time.Time

		t.Run("use key (no cache hit)", func(t *testing.T) {
			_, extClient := setupExternalAPIClientWithAuth(key.RawApiKey)
			_, err := extClient.ExternalServiceExecutionBlock(&external_service.ExternalServiceExecutionBlockParams{BlockNumber: "1"}, nil)
			assert.NoError(t, err)
		})

		t.Run("last used timestamp updated", func(t *testing.T) {
			got := apiKeyMgmt.mustGetKey(t, client, testAPIKeyLifecycle)
			assert.NotNil(t, got.ApiKey.LastUsedAt)
			lastUsedAt = got.ApiKey.LastUsedAt.AsTime()
		})

		time.Sleep(1 * time.Second) // last used timestamps have second precision, so wait a bit to ensure next usage is chronologically after

		t.Run("use key (cache hit)", func(t *testing.T) {
			_, extClient := setupExternalAPIClientWithAuth(key.RawApiKey)
			_, err := extClient.ExternalServiceExecutionBlock(&external_service.ExternalServiceExecutionBlockParams{BlockNumber: "1"}, nil)
			assert.NoError(t, err)
		})

		t.Run("last used timestamp is updated and chronological", func(t *testing.T) {
			got := apiKeyMgmt.mustGetKey(t, client, testAPIKeyLifecycle)
			assert.NotNil(t, got.ApiKey.LastUsedAt)
			assert.True(t, got.ApiKey.LastUsedAt.AsTime().After(lastUsedAt), "expected last used timestamp to be updated in chronological order (old: %v, new: %v)", lastUsedAt, got.ApiKey.LastUsedAt.AsTime())
		})

		// Disable

		t.Run("disable key", func(t *testing.T) {
			_, err := client.DisableAPIKey(ctx, &model.DisableAPIKeyRequest{Name: testAPIKeyLifecycle})
			assert.NoError(t, err)

			got := apiKeyMgmt.mustGetKey(t, client, testAPIKeyLifecycle)
			assert.NotNil(t, got.ApiKey.DisabledAt)
		})

		t.Run("disabled key cannot be used", func(t *testing.T) {
			_, extClient := setupExternalAPIClientWithAuth(key.RawApiKey)
			_, err := extClient.ExternalServiceExecutionBlock(&external_service.ExternalServiceExecutionBlockParams{BlockNumber: "1"}, nil)
			assert.Error(t, err)
			if apiErr, ok := err.(*external_service.ExternalServiceExecutionBlockDefault); ok {
				assert.Equal(t, 401, apiErr.Code())
			} else {
				t.Fatalf("expected unauthorized error, got: %v", err)
			}
		})

		t.Run("disabled key timestamp not changing after disabling again", func(t *testing.T) {
			key := apiKeyMgmt.mustGetKey(t, client, testAPIKeyLifecycle)
			assert.NotNil(t, key.ApiKey.DisabledAt)

			keyDisabledAgain, err := client.DisableAPIKey(ctx, &model.DisableAPIKeyRequest{Name: testAPIKeyLifecycle})
			assert.NoError(t, err)

			assert.Equal(t, key.ApiKey.DisabledAt, keyDisabledAgain.ApiKey.DisabledAt)
		})

		// Enable
		t.Run("enable key", func(t *testing.T) {
			_, err := client.EnableAPIKey(ctx, &model.EnableAPIKeyRequest{Name: testAPIKeyLifecycle})
			assert.NoError(t, err)

			got := apiKeyMgmt.mustGetKey(t, client, testAPIKeyLifecycle)
			assert.Nil(t, got.ApiKey.DisabledAt)
		})

		t.Run("enabled key can be used again", func(t *testing.T) {
			_, extClient := setupExternalAPIClientWithAuth(key.RawApiKey)
			_, err := extClient.ExternalServiceExecutionBlock(&external_service.ExternalServiceExecutionBlockParams{BlockNumber: "1"}, nil)
			assert.NoError(t, err)
		})

		t.Run("enabling an enabled key should just return ok", func(t *testing.T) {
			_, err := client.EnableAPIKey(ctx, &model.EnableAPIKeyRequest{Name: testAPIKeyLifecycle})
			assert.NoError(t, err)
		})

		// Deletion

		t.Run("delete key", func(t *testing.T) {
			_, err := client.DeleteAPIKey(ctx, &model.DeleteAPIKeyRequest{Name: testAPIKeyLifecycle})
			assert.NoError(t, err)
			assert.Equal(t, codes.OK, status.Code(err))
		})

		t.Run("deleted key cannot be used", func(t *testing.T) {
			_, extClient := setupExternalAPIClientWithAuth(key.RawApiKey)
			_, err := extClient.ExternalServiceExecutionBlock(&external_service.ExternalServiceExecutionBlockParams{BlockNumber: "1"}, nil)
			assert.Error(t, err)
			if apiErr, ok := err.(*external_service.ExternalServiceExecutionBlockDefault); ok {
				assert.Equal(t, 401, apiErr.Code())
			} else {
				t.Fatalf("expected unauthorized error, got: %v", err)
			}
		})

	})
}

func TestAPIKeyCreationMaxLimit(t *testing.T) {
	var testAPIKeyMaxLimit = apiKeyMgmt.newTestKey("max-key")

	limiter := limits.NewLimiter()
	apiKeyMgmt.withCleanState(t, func(ctx context.Context, client model.InternalServiceClient) {
		tierMaxLimit, err := limiter.GetMaxAPIKeys(ctx, &domain.User{ID: 1, SubscriptionTier: domain.TierScale})
		assert.NoError(t, err)

		var hasReachedLimit bool
		for i := 0; i < tierMaxLimit+1; i++ { // one above tier limit
			name := fmt.Sprintf("%s-%d", testAPIKeyMaxLimit, i)
			_, err := client.CreateAPIKey(ctx, &model.CreateAPIKeyRequest{Name: name})
			if status.Code(err) == codes.ResourceExhausted {
				hasReachedLimit = true
				break
			}
		}
		assert.True(t, hasReachedLimit, "expected to eventually hit resource limit")
	})
}

func TestAPIKeyList(t *testing.T) {
	var testAPIKeyList = apiKeyMgmt.newTestKey("list-key")

	apiKeyMgmt.withCleanState(t, func(ctx context.Context, client model.InternalServiceClient) {
		// Create some keys
		keysToCreate := []string{fmt.Sprintf("%s-1", testAPIKeyList), fmt.Sprintf("%s-2", testAPIKeyList), fmt.Sprintf("%s-3", testAPIKeyList)}
		for _, name := range keysToCreate {
			_, err := client.CreateAPIKey(ctx, &model.CreateAPIKeyRequest{Name: name})
			assert.NoError(t, err)
		}

		t.Run("list keys", func(t *testing.T) {
			resp, err := client.GetAPIKeys(ctx, &model.GetAPIKeysRequest{})
			assert.NoError(t, err)
			foundKeys := make(map[string]bool)
			for _, k := range resp.ApiKeys {
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
		_, extClient := setupExternalAPIClientWithAuth(testAPIKeyInvalid)
		_, err := extClient.ExternalServiceExecutionBlock(&external_service.ExternalServiceExecutionBlockParams{BlockNumber: "1"}, nil)
		assert.Error(t, err)
		if apiErr, ok := err.(*external_service.ExternalServiceExecutionBlockDefault); ok {
			assert.Equal(t, 401, apiErr.Code())
		} else {
			t.Fatalf("expected unauthorized error, got: %v", err)
		}
	})

	t.Run("usage without key", func(t *testing.T) {
		r := httptransport.New(testUtils.GetExternalHTTPUrl(), client.DefaultBasePath, client.DefaultSchemes)
		extClient := client.New(r, strfmt.Default)
		_, err := extClient.ExternalService.ExternalServiceExecutionBlock(&external_service.ExternalServiceExecutionBlockParams{BlockNumber: "1"}, nil)
		assert.Error(t, err)
		if apiErr, ok := err.(*external_service.ExternalServiceExecutionBlockDefault); ok {
			assert.Equal(t, 401, apiErr.Code())
		} else {
			t.Fatalf("expected unauthorized error, got: %v", err)
		}
	})
}

// Todo: Make sure to add tests trying to manage keys as a different user once internal auth is in place

// --- Helpers ---

func (*apiKeyMgmtTest) mustGetKey(t *testing.T, client model.InternalServiceClient, name string) *model.GetAPIKeyResponse {
	t.Helper()
	in := &model.GetAPIKeyRequest{Name: name}
	k, err := client.GetAPIKey(context.Background(), in)
	assert.NoError(t, err, "failed to get key %s", name)
	return k
}

func (*apiKeyMgmtTest) cleanState(t *testing.T, client model.InternalServiceClient, ctx context.Context) {
	keys, err := client.GetAPIKeys(ctx, &model.GetAPIKeysRequest{})
	if err != nil {
		t.Fatal("failed to list API keys for cleanup:", err)
	}
	for _, k := range keys.ApiKeys {
		if !strings.HasPrefix(k.Name, apiKeyMgmt.testRunID) { // scoped to cleanup just the keys of this test suite
			continue
		}
		_, err := client.DeleteAPIKey(ctx, &model.DeleteAPIKeyRequest{Name: k.Name})
		if err != nil {
			t.Fatal("failed to delete API key:", err)
		}
	}
}

func (apiKeyMgmt *apiKeyMgmtTest) newTestKey(name string) string {
	return fmt.Sprintf("%s-%s", apiKeyMgmt.testRunID, name) // prefix key to know what to cleanup
}

func (apiKeyMgmt *apiKeyMgmtTest) withCleanState(t *testing.T, test func(ctx context.Context, client model.InternalServiceClient)) {
	t.Helper()
	ctx, client := setupInternalAPIClient(t)

	apiKeyMgmt.cleanState(t, client, ctx)

	test(ctx, client)

	t.Cleanup(func() {
		apiKeyMgmt.cleanState(t, client, ctx)
	})
}
