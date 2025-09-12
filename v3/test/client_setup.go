package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gobitfly/beaconchain-backend/api/external/client"
	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/test/testUtils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func setupExternalAPIClient(t *testing.T) (context.Context, *client.ClientWithResponses) {
	apiKey := os.Getenv("API_KEY_ORCA_TEST")
	if apiKey == "" {
		t.Fatal("API_KEY_ORCA_TEST environment variable is not set")
	}
	return setupExternalAPIClientWithAuth(apiKey)
}

func setupExternalAPIClientWithAuth(apiKey string) (context.Context, *client.ClientWithResponses) {
	authFn := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+apiKey)
		return nil
	}

	cl, err := client.NewClientWithResponses(
		fmt.Sprintf("http://%s", testUtils.GetExternalHTTPUrl()),
		client.WithRequestEditorFn(authFn),
	)
	if err != nil {
		panic(err)
	}

	return context.Background(), cl
}

func setupInternalAPIClient(t *testing.T) (context.Context, model.InternalServiceClient) {
	conn, err := grpc.NewClient(testUtils.GetInternalGRPCUrl(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect to gRPC server: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
	})
	return context.Background(), model.NewInternalServiceClient(conn)
}
