package integration

import (
	"context"
	"fmt"
	"testing"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/test/testUtils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var testKeyName = "test-key"

func TestAPIKeyLifecycle(t *testing.T) {
	var key *model.CreateAPIKeyResponse

	withCleanState(t, func(ctx context.Context, client model.InternalServiceClient) {

		// Creation
		t.Run("create key", func(t *testing.T) {
			var err error
			key, err = client.CreateAPIKey(ctx, &model.CreateAPIKeyRequest{Name: testKeyName})
			assert.NoError(t, err)
		})

		t.Run("get after creation", func(t *testing.T) {
			got := mustGetKey(t, client, testKeyName)
			assert.Equal(t, testKeyName, got.ApiKey.Name)
			assert.Nil(t, got.ApiKey.LastUsedAt)
			assert.Nil(t, got.ApiKey.DisabledAt)
			assert.NotNil(t, got.ApiKey.CreatedAt)
		})

		// Usage

		t.Run("use key", func(t *testing.T) {
			extCtx, extClient := setupExternalAPIClientWithAPIKey(t, key.RawApiKey)
			_, err := extClient.ExecutionBlock(extCtx, &model.ExecutionBlockRequest{BlockNumber: "1"})
			assert.NoError(t, err)
		})

		t.Run("last used timestamp updated", func(t *testing.T) {
			got := mustGetKey(t, client, testKeyName)
			assert.NotNil(t, got.ApiKey.LastUsedAt)
		})

		// Disable

		t.Run("disable key", func(t *testing.T) {
			_, err := client.DisableAPIKey(ctx, &model.DisableAPIKeyRequest{Name: testKeyName})
			assert.NoError(t, err)

			got := mustGetKey(t, client, testKeyName)
			assert.NotNil(t, got.ApiKey.DisabledAt)
		})

		t.Run("disabled key cannot be used", func(t *testing.T) {
			extCtx, extClient := setupExternalAPIClientWithAPIKey(t, key.RawApiKey)
			_, err := extClient.ExecutionBlock(extCtx, &model.ExecutionBlockRequest{BlockNumber: "1"})
			assert.Error(t, err)
			assert.Equal(t, codes.Unauthenticated, status.Code(err))
		})

		t.Run("disabled key timestamp not changing after disabling again", func(t *testing.T) {
			key := mustGetKey(t, client, testKeyName)
			assert.NotNil(t, key.ApiKey.DisabledAt)

			keyDisabledAgain, err := client.DisableAPIKey(ctx, &model.DisableAPIKeyRequest{Name: testKeyName})
			assert.NoError(t, err)

			assert.Equal(t, key.ApiKey.DisabledAt, keyDisabledAgain.ApiKey.DisabledAt)
		})

		// Enable
		t.Run("enable key", func(t *testing.T) {
			_, err := client.EnableAPIKey(ctx, &model.EnableAPIKeyRequest{Name: testKeyName})
			assert.NoError(t, err)

			got := mustGetKey(t, client, testKeyName)
			assert.Nil(t, got.ApiKey.DisabledAt)
		})

		t.Run("enabled key can be used again", func(t *testing.T) {
			extCtx, extClient := setupExternalAPIClientWithAPIKey(t, key.RawApiKey)
			_, err := extClient.ExecutionBlock(extCtx, &model.ExecutionBlockRequest{BlockNumber: "1"})
			assert.NoError(t, err)
		})

		t.Run("enabling an enabled key should just return ok", func(t *testing.T) {
			_, err := client.EnableAPIKey(ctx, &model.EnableAPIKeyRequest{Name: testKeyName})
			assert.NoError(t, err)
		})

		// Deletion

		t.Run("delete key", func(t *testing.T) {
			_, err := client.DeleteAPIKey(ctx, &model.DeleteAPIKeyRequest{Name: testKeyName})
			assert.Nil(t, err)
			assert.Equal(t, codes.OK, status.Code(err))
		})

		t.Run("deleted key cannot be used", func(t *testing.T) {
			extCtx, extClient := setupExternalAPIClientWithAPIKey(t, key.RawApiKey)
			_, err := extClient.ExecutionBlock(extCtx, &model.ExecutionBlockRequest{BlockNumber: "1"})
			assert.Error(t, err)
			assert.Equal(t, codes.Unauthenticated, status.Code(err))
		})

	})
}

func TestAPIKeyCreationMaxLimit(t *testing.T) {
	withCleanState(t, func(ctx context.Context, client model.InternalServiceClient) {
		var hasReachedLimit bool
		for i := 0; i < 20; i++ {
			name := fmt.Sprintf("test-key-%d", i)
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
	withCleanState(t, func(ctx context.Context, client model.InternalServiceClient) {
		// Create some keys
		keysToCreate := []string{"key1", "key2", "key3"}
		for _, name := range keysToCreate {
			_, err := client.CreateAPIKey(ctx, &model.CreateAPIKeyRequest{Name: name})
			assert.NoError(t, err)
		}

		t.Run("list keys", func(t *testing.T) {
			resp, err := client.GetAPIKeys(ctx, &model.GetAPIKeysRequest{})
			assert.NoError(t, err)
			assert.Len(t, resp.ApiKeys, len(keysToCreate))
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

func TestAPIKeyNotFound(t *testing.T) {
	t.Run("delete non-existent key", func(t *testing.T) {
		ctx, client := setupInternalAPIClient(t)
		_, err := client.DeleteAPIKey(ctx, &model.DeleteAPIKeyRequest{Name: "non-existent"})
		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("disable non-existent key", func(t *testing.T) {
		ctx, client := setupInternalAPIClient(t)
		_, err := client.DisableAPIKey(ctx, &model.DisableAPIKeyRequest{Name: "non-existent"})
		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("enable non-existent key", func(t *testing.T) {
		ctx, client := setupInternalAPIClient(t)
		_, err := client.EnableAPIKey(ctx, &model.EnableAPIKeyRequest{Name: "non-existent"})
		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("get non-existent key", func(t *testing.T) {
		ctx, client := setupInternalAPIClient(t)
		_, err := client.GetAPIKey(ctx, &model.GetAPIKeyRequest{Name: "non-existent"})
		assert.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}

func TestAPIKeyInvalidUsages(t *testing.T) {
	t.Run("invalid key cannot be used", func(t *testing.T) {
		extCtx, extClient := setupExternalAPIClientWithAPIKey(t, "invalid-api-key")
		_, err := extClient.ExecutionBlock(extCtx, &model.ExecutionBlockRequest{BlockNumber: "1"})
		assert.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("usage without key", func(t *testing.T) {
		conn, err := grpc.NewClient(testUtils.GetExternalGRPCUrl(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatalf("failed to connect to gRPC server: %v", err)
		}
		t.Cleanup(func() {
			conn.Close()
		})

		extClient := model.NewExternalServiceClient(conn)
		extCtx := context.Background()
		_, err = extClient.ExecutionBlock(extCtx, &model.ExecutionBlockRequest{BlockNumber: "1"})
		assert.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestAPIKeyCreationFailures(t *testing.T) {
	withCleanState(t, func(ctx context.Context, client model.InternalServiceClient) {
		t.Run("create duplicate fails", func(t *testing.T) {
			_, err := client.CreateAPIKey(ctx, &model.CreateAPIKeyRequest{Name: testKeyName})
			assert.Nil(t, err)

			_, err = client.CreateAPIKey(ctx, &model.CreateAPIKeyRequest{Name: testKeyName})
			assert.Error(t, err)
			assert.Equal(t, codes.AlreadyExists, status.Code(err))
		})

		t.Run("create invalid name fails", func(t *testing.T) {
			_, err := client.CreateAPIKey(ctx, &model.CreateAPIKeyRequest{Name: "invalid key name with spaces"})
			assert.Error(t, err)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
		t.Run("create empty name fails", func(t *testing.T) {
			_, err := client.CreateAPIKey(ctx, &model.CreateAPIKeyRequest{Name: ""})
			assert.Error(t, err)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	})
}

// Todo: Make sure to add tests trying to manage keys as a different user once internal auth is in place

// --- Helpers ---

func mustGetKey(t *testing.T, client model.InternalServiceClient, name string) *model.GetAPIKeyResponse {
	t.Helper()
	in := &model.GetAPIKeyRequest{Name: name}
	k, err := client.GetAPIKey(context.Background(), in)
	assert.NoError(t, err, "failed to get key %s", name)
	return k
}

func cleanState(client model.InternalServiceClient, ctx context.Context) {
	keys, err := client.GetAPIKeys(ctx, &model.GetAPIKeysRequest{})
	if err != nil {
		panic(err)
	}
	for _, k := range keys.ApiKeys {
		_, err := client.DeleteAPIKey(ctx, &model.DeleteAPIKeyRequest{Name: k.Name})
		if err != nil {
			panic(err)
		}
	}
}

func withCleanState(t *testing.T, test func(ctx context.Context, client model.InternalServiceClient)) {
	t.Helper()
	ctx, client := setupInternalAPIClient(t)

	cleanState(client, ctx)

	test(ctx, client)

	t.Cleanup(func() {
		cleanState(client, ctx)
	})
}
