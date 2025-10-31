package app

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestBuildEndpointRateLimitsMap_Success(t *testing.T) {
	rawSpec := `
openapi: 3.0.0
paths:
  /users:
    get:
      operationId: listUsers
      summary: Retrieves a list of users
      x-ratelimits:
        free:
          steady_rate: 10.5
          bucket_capacity: 100
        pro:
          steady_rate: 50
          bucket_capacity: 500
  /items:
    post:
      operationId: createItem
      summary: Creates a new item (no rate limit)
      responses:
        "201":
          description: Created
`

	rawBytes := []byte(rawSpec)
	spec, err := openapi3.NewLoader().LoadFromData(rawBytes)
	assert.NoError(t, err, "Failed to load OpenAPI spec")

	expectedMap := map[endpointRateLimitKey]domain.RateLimitSettings{
		{operationID: "listUsers", tier: domain.Tier("FREE")}: {SteadyRate: 10.5, BucketCapacity: 100},
		{operationID: "listUsers", tier: domain.Tier("PRO")}:  {SteadyRate: 50.0, BucketCapacity: 500},
	}

	actual, err := buildEndpointRateLimitsMap(spec)

	assert.NoError(t, err, "Unexpected error from buildEndpointRateLimitsMap")
	assert.Equal(t, expectedMap, actual, "The actual rate limits map should match the expected map")
}

func TestBuildEndpointRateLimitsMap_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		rawSpec string
	}{
		{
			name: "MissingOperationID",
			rawSpec: `
openapi: 3.0.0
paths:
  /items:
    post:
      summary: Creates a new item (no operationId)
      responses:
        "201":
          description: Created
`,
		},
		{
			name: "WrongType_SteadyRate",
			rawSpec: `
openapi: 3.0.0
paths:
  /users:
    get:
      operationId: listUsers
      summary: Retrieves a list of users
      x-ratelimits:
        free:
          steady_rate: "test" # Should be a number
          bucket_capacity: 10
`,
		},
		{
			name: "WrongType_BucketCapacity",
			rawSpec: `
openapi: 3.0.0
paths:
  /users:
    get:
      operationId: listUsers
      summary: Retrieves a list of users
      x-ratelimits:
        free:
          steady_rate: 10
          bucket_capacity: "test" # Should be a number
`,
		},
		{
			name: "EmptyBucketCapacity",
			rawSpec: `
openapi: 3.0.0
paths:
  /users:
    get:
      operationId: listUsers
      summary: Retrieves a list of users
      x-ratelimits:
        free:
          steady_rate: 10.5
          # missing bucket_capacity
`,
		},
		{
			name: "EmptySteadyRate",
			rawSpec: `
openapi: 3.0.0
paths:
  /users:
    get:
      operationId: listUsers
      summary: Retrieves a list of users
      x-ratelimits:
        free:
          bucket_capacity: 100
          # missing steady_rate
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawBytes := []byte(tt.rawSpec)
			spec, err := openapi3.NewLoader().LoadFromData(rawBytes)
			assert.NoError(t, err, "Failed to load OpenAPI spec")

			_, err = buildEndpointRateLimitsMap(spec)
			assert.Error(t, err)
		})
	}
}
