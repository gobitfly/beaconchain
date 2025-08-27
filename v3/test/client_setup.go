package integration

import (
	"context"
	"os"
	"testing"

	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/test/testUtils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// struct for creds for grpc clients
type testPerRPCCred struct {
	token string
}

func (ts testPerRPCCred) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": "Bearer " + ts.token,
	}, nil
}

func (ts testPerRPCCred) RequireTransportSecurity() bool {
	return false
}

func setupExternalAPIClient(t *testing.T) (context.Context, model.ExternalServiceClient) {
	apiKey := os.Getenv("API_KEY_ORCA_TEST")
	if apiKey == "" {
		t.Fatal("API_KEY_ORCA_TEST environment variable is not set")
	}
	return setupExternalAPIClientWithAPIKey(t, apiKey)
}

func setupExternalAPIClientWithAPIKey(t *testing.T, apikey string) (context.Context, model.ExternalServiceClient) {
	perRPC := testPerRPCCred{token: apikey}
	conn, err := grpc.NewClient(testUtils.GetExternalGRPCUrl(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(perRPC))
	if err != nil {
		t.Fatalf("failed to connect to gRPC server: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
	})
	return context.Background(), model.NewExternalServiceClient(conn)
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
