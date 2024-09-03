package handlers_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gobitfly/beaconchain/pkg/api"
	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	api_types "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/stretchr/testify/assert"
)

var ts *testServer
var dataAccessor dataaccess.DataAccessor

type testServer struct {
	*httptest.Server
}

// Implement a get() method on our custom testServer type. This makes a GET
// request to a given url path using the test server client, and returns the
// response status code, headers and body.
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)
	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) post(t *testing.T, urlPath string, data io.Reader) (int, http.Header, string) {
	rs, err := ts.Client().Post(ts.URL+urlPath, "application/json", data)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)
	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) parseErrorResonse(t *testing.T, body string) api_types.ApiErrorResponse {
	resp := api_types.ApiErrorResponse{}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestMain(m *testing.M) {
	setup()
	defer teardown()

	// wait till service initialization is completed (TODO: find a better way to do this)
	// time.Sleep(30 * time.Second)

	os.Exit(m.Run())
}

func teardown() {
	dataAccessor.Close()
	ts.Close()
}

func setup() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")

	flag.Parse()

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	}
	utils.Config = cfg

	log.InfoWithFields(log.Fields{"config": *configPath, "version": version.Version, "commit": version.GitCommit, "chainName": utils.Config.Chain.ClConfig.ConfigName}, "starting")

	dataAccessor = dataaccess.NewDataAccessService(cfg)
	router := api.NewApiRouter(dataAccessor, cfg)

	ts = &testServer{httptest.NewTLSServer(router)}

	jar, _ := cookiejar.New(nil)
	ts.Server.Client().Jar = jar
}

func TestInternalGetProductSummaryHandler(t *testing.T) {
	code, _, body := ts.get(t, "/api/i/product-summary")

	assert.Equal(t, http.StatusOK, code)

	respData := api_types.InternalGetProductSummaryResponse{}
	err := json.Unmarshal([]byte(body), &respData)
	if err != nil {
		log.Infof("%s", body)
		t.Fatal(err)
	}

	assert.NotEqual(t, 0, respData.Data.ValidatorsPerDashboardLimit, "ValidatorsPerDashboardLimit should not be 0")
	assert.NotEqual(t, 0, len(respData.Data.ApiProducts), "ApiProducts should not be empty")
	assert.NotEqual(t, 0, len(respData.Data.ExtraDashboardValidatorsPremiumAddon), "ExtraDashboardValidatorsPremiumAddon should not be empty")
	assert.NotEqual(t, 0, len(respData.Data.PremiumProducts), "PremiumProducts should not be empty")
}

func TestInternalGetLatestStateHandler(t *testing.T) {
	code, _, body := ts.get(t, "//api/i/latest-state")
	assert.Equal(t, http.StatusOK, code)

	respData := api_types.InternalGetLatestStateResponse{}
	if err := json.Unmarshal([]byte(body), &respData); err != nil {
		t.Fatal(err)
	}

	assert.NotEqual(t, uint64(0), respData.Data.LatestSlot, "latest slot should not be 0")
	assert.NotEqual(t, uint64(0), respData.Data.FinalizedEpoch, "finalized epoch should not be 0")
}

func TestInternalPostAdConfigurationsHandler(t *testing.T) {
	code, _, body := ts.get(t, "/api/i/ad-configurations")
	assert.Equal(t, http.StatusUnauthorized, code)

	resp := ts.parseErrorResonse(t, body)
	assert.Equal(t, "unauthorized: not authenticated", resp.Error)

	// login
	code, _, body = ts.post(t, "/api/i/login", bytes.NewBuffer([]byte(`{"email": "admin@admin.com", "password": "admin"}`)))
	assert.Equal(t, http.StatusNotFound, code)
	resp = ts.parseErrorResonse(t, body)
	assert.Equal(t, "not found: user not found", resp.Error)
}

func TestInternalLoginHandler(t *testing.T) {
	// login with email in wrong format
	code, _, body := ts.post(t, "/api/i/login", bytes.NewBuffer([]byte(`{"email": "admin", "password": "admin"}`)))
	assert.Equal(t, http.StatusBadRequest, code)
	resp := ts.parseErrorResonse(t, body)
	assert.Equal(t, "email: given value 'admin' has incorrect format", resp.Error, "unexpected error message")

	// login with wrong user
	code, _, body = ts.post(t, "/api/i/login", bytes.NewBufferString(`{"email": "admin@admin.com", "password": "admin"}`))
	assert.Equal(t, http.StatusNotFound, code)
	resp = ts.parseErrorResonse(t, body)
	assert.Equal(t, "not found: user not found", resp.Error, "unexpected error message") // TODO: this should not return the same error as a user with a wrong password
}

func TestInternalSearchHandler(t *testing.T) {
	// search for validator with index 5
	code, _, body := ts.post(t, "/api/i/search", bytes.NewBufferString(`{"input":"5","networks":[17000],"types":["validators_by_deposit_ens_name","validators_by_deposit_address","validators_by_withdrawal_ens_name","validators_by_withdrawal_address","validators_by_withdrawal_credential","validator_by_index","validator_by_public_key","validators_by_graffiti"]}`))
	assert.Equal(t, 200, code)

	resp := api_types.InternalPostSearchResponse{}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatal(err)
	}

	assert.NotEqual(t, 0, len(resp.Data), "response data should not be empty")
	assert.NotNil(t, resp.Data[0].NumValue, "validator index should not be nil")
	assert.Equal(t, uint64(5), *resp.Data[0].NumValue, "validator index should be 5")
}
