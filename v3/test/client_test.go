package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/gobitfly/beaconchain-backend/api/external/client"
	"github.com/gobitfly/beaconchain-backend/test/testUtils"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {

	t.Run("withResponse", func(t *testing.T) {
		cl, _ := client.NewClientWithResponses(fmt.Sprintf("http://%s", testUtils.GetExternalHTTPUrl()))
		resp, err := ping(context.Background(), cl)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("normal", func(t *testing.T) {
		cl, _ := client.NewClientWithResponses(fmt.Sprintf("http://%s", testUtils.GetExternalHTTPUrl()))
		resp, err := ping(context.Background(), cl)
		assert.NoError(t, resp.Body.Close())
		assert.NoError(t, err)
		assert.Equal(t, 401, resp.StatusCode)
	})
}
