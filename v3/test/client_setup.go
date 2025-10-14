package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	extclient "github.com/gobitfly/beaconchain-backend/api/external/client"
	inhouseclient "github.com/gobitfly/beaconchain-backend/api/inhouse/client"
	"github.com/gobitfly/beaconchain-backend/test/testUtils"
)

func setupExternalAPIClient(t *testing.T) (context.Context, *extclient.ClientWithResponses) {
	apiKey := os.Getenv("API_KEY_ORCA_TEST")
	if apiKey == "" {
		t.Fatal("API_KEY_ORCA_TEST environment variable is not set")
	}
	return setupExternalAPIClientWithAuth(apiKey)
}

func setupExternalAPIClientWithAuth(apiKey string) (context.Context, *extclient.ClientWithResponses) {
	authFn := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+apiKey)
		return nil
	}

	cl, err := extclient.NewClientWithResponses(
		fmt.Sprintf("http://%s", testUtils.GetExternalHTTPUrl()),
		extclient.WithRequestEditorFn(authFn),
	)
	if err != nil {
		panic(err)
	}

	return context.Background(), cl
}

func setupInhouseAPIClient(t *testing.T) (context.Context, *inhouseclient.ClientWithResponses) {
	sessionID := os.Getenv("SESSION_ID_ORCA_TEST")
	if sessionID == "" {
		t.Fatal("SESSION_ID_ORCA_TEST environment variable is not set")
	}
	return setupInhouseAPIClientWithAuth(sessionID)
}

func setupInhouseAPIClientWithAuth(sessionID string) (context.Context, *inhouseclient.ClientWithResponses) {
	authFn := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Cookie", "session_id="+sessionID)
		return nil
	}

	cl, err := inhouseclient.NewClientWithResponses(
		fmt.Sprintf("http://%s", testUtils.GetInternalHTTPUrl()),
		inhouseclient.WithRequestEditorFn(authFn),
	)
	if err != nil {
		panic(err)
	}

	return context.Background(), cl
}
