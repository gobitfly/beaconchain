package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	model "github.com/gobitfly/beaconchain-backend/api/gen/api_service/v1"
	"github.com/gobitfly/beaconchain-backend/api/gen/client/client"
	"github.com/gobitfly/beaconchain-backend/api/gen/client/client/external_service"
	"github.com/gobitfly/beaconchain-backend/test/testUtils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// struct for creds for grpc clients
// bearerAuth implements ClientAuthInfoWriter
type bearerAuth struct {
	token string
}

// AuthenticateRequest sets the Authorization header
func (b *bearerAuth) AuthenticateRequest(req runtime.ClientRequest, _ strfmt.Registry) error {
	return req.SetHeaderParam("Authorization", fmt.Sprintf("Bearer %s", b.token))
}

// NewBearerAuth returns a ClientAuthInfoWriter that injects a Bearer token
func NewBearerAuth(token string) runtime.ClientAuthInfoWriter {
	return &bearerAuth{token: token}
}

func getExternalAuth(t *testing.T) runtime.ClientAuthInfoWriter {
	apiKey := os.Getenv("API_KEY_ORCA_TEST")
	if apiKey == "" {
		t.Fatal("API_KEY_ORCA_TEST environment variable is not set")
	}
	return getExternalAuthFromAPIKey(apiKey)
}

func getExternalAuthFromAPIKey(apiKey string) runtime.ClientAuthInfoWriter {
	return NewBearerAuth(apiKey)
}

func setupExternalAPIClient(t *testing.T) (context.Context, external_service.ClientService) {
	cl := client.NewHTTPClientWithConfig(
		nil,
		&client.TransportConfig{
			Host:     testUtils.GetExternalHTTPUrl(),
			BasePath: "/",
			Schemes:  []string{"http"},
		},
	)
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
