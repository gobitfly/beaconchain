package integration

import (
	"context"
	"os"
	"testing"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/api/gen/client/client"
	"github.com/gobitfly/beaconchain-backend/api/gen/client/client/external_service"
	"github.com/gobitfly/beaconchain-backend/test/testUtils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func setupExternalAPIClient(t *testing.T) (context.Context, external_service.ClientService) {
	apiKey := os.Getenv("API_KEY_ORCA_TEST")
	if apiKey == "" {
		t.Fatal("API_KEY_ORCA_TEST environment variable is not set")
	}
	return setupExternalAPIClientWithAuth(apiKey)
}

func setupExternalAPIClientWithAuth(apiKey string) (context.Context, external_service.ClientService) {
	r := httptransport.New(testUtils.GetExternalHTTPUrl(), client.DefaultBasePath, client.DefaultSchemes)
	r.DefaultAuthentication = httptransport.BearerToken(apiKey)
	cl := client.New(r, strfmt.Default)

	return context.Background(), cl.ExternalService
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
