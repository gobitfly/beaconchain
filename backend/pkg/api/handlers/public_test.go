package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	dataAccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/stretchr/testify/assert"
)

var dataAccessor dataAccess.DataAccessor
var hs *HandlerService

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() {
	dataAccessor = dataAccess.NewDummyService()
	hs = NewHandlerService(dataAccessor, dataAccessor, nil, false)
	endpoints = []endpoint{
		{http.MethodGet, "/healthz", hs.PublicGetHealthz},
		{http.MethodGet, "/healthz-loadbalancer", hs.PublicGetHealthzLoadbalancer},

		{http.MethodGet, "/ratelimit-weights", hs.InternalGetRatelimitWeights},
		{http.MethodPost, "/ratelimit-weights", hs.PublicPostValidatorDashboards},
	}
}

func shutdown() {
	/*if dataAccessor != nil {
		dataAccessor.Close()
	}
	if ts != nil {
		ts.Close()
	}
	if postgres != nil {
		err := postgres.Stop()
		if err != nil {
			log.Error(err, "error stopping embedded postgres", 0)
		}
	}*/
}

type endpoint struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

var endpoints = []endpoint{}

func TestPublicGetHealthz(t *testing.T) {
	req, err := http.NewRequest(endpoints[0].Method, endpoints[0].Path, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	endpoints[0].Handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	// TODO how to test healthy reply?
}

func TestPublicPostValidatorDashboards(t *testing.T) {
	tests := []map[string]string{}
	tests = append(tests, map[string]string{
		//"api_key": "123",
	})

	endpoint := endpoints[3]

	// base req
	req, err := http.NewRequest(endpoint.Method, endpoint.Path, nil)
	if err != nil {
		t.Fatal(err)
	}
	result := httptest.NewRecorder()

	t.Run("unauthenticated", func(t *testing.T) {
		if err != nil {
			t.Fatal(err)
		}
		endpoint.Handler.ServeHTTP(result, req)
		assert.Equal(t, http.StatusUnauthorized, result.Code, "should be unauthorized")
	})

	// add auth
	ctx := req.Context()
	ctx = context.WithValue(ctx, types.CtxUserIdKey, uint64(1))
	req = req.WithContext(ctx)

	t.Run("body missing", func(t *testing.T) {
		endpoint.Handler.ServeHTTP(result, req)
		assert.Equal(t, http.StatusOK, result.Code, "should be authorized")
	})

	t.Run("body no json", func(t *testing.T) {
		endpoint.Handler.ServeHTTP(result, req)
		assert.Equal(t, http.StatusOK, result.Code, "should be authorized")
	})

	t.Run("body json schema invalid", func(t *testing.T) {
		endpoint.Handler.ServeHTTP(result, req)
		assert.Equal(t, http.StatusOK, result.Code, "should be authorized")
	})

	q := req.URL.Query()
	for k, v := range tests[0] {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	endpoints[3].Handler.ServeHTTP(result, req)
	assert.Equal(t, http.StatusInternalServerError, result.Code)
}
