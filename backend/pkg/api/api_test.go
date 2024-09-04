package api_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/gobitfly/beaconchain/pkg/api"
	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	api_types "github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

var ts *testServer
var dataAccessor dataaccess.DataAccessor
var postgres *embeddedpostgres.EmbeddedPostgres

type testServer struct {
	*httptest.Server
}

func (ts *testServer) request(t *testing.T, method, urlPath string, data io.Reader) (int, string) {
	req, err := http.NewRequest(method, ts.URL+urlPath, data)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)
	return rs.StatusCode, string(body)
}
func (ts *testServer) get(t *testing.T, urlPath string) (int, string) {
	return ts.request(t, http.MethodGet, urlPath, nil)
}
func (ts *testServer) post(t *testing.T, urlPath string, data io.Reader) (int, string) {
	return ts.request(t, http.MethodPost, urlPath, data)
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
	code := m.Run()
	teardown()

	// wait till service initialization is completed (TODO: find a better way to do this)
	// time.Sleep(30 * time.Second)

	os.Exit(code)
}

func teardown() {
	dataAccessor.Close()
	ts.Close()
	err := postgres.Stop()
	if err != nil {
		log.Error(err, "error stopping embedded postgres", 0)
	}
}

func setup() {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")

	flag.Parse()

	// terminate any currently running postgres instances
	_ = exec.Command("pkill", "-9", "postgres").Run()

	// start embedded postgres
	postgres = embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().Username("postgres"))
	err := postgres.Start()
	if err != nil {
		log.Fatal(err, "error starting embedded postgres", 0)
	}

	// connection the the embedded db and run migrations
	tempDb, err := sqlx.Connect("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatal(err, "error connection to test db", 0)
	}

	if err := goose.Up(tempDb.DB, "../../pkg/commons/db/migrations/postgres"); err != nil {
		log.Fatal(err, "error running migrations", 0)
	}

	// insert dummy user for testing (email: admin@admin, password: admin)
	pHash, _ := bcrypt.GenerateFromPassword([]byte("admin"), 10)
	_, err = tempDb.Exec(`
      INSERT INTO users (password, email, register_ts, api_key, email_confirmed)
      VALUES ($1, $2, TO_TIMESTAMP($3), $4, $5)`,
		string(pHash), "admin@admin.com", time.Now().Unix(), "admin", true,
	)
	if err != nil {
		log.Fatal(err, "error inserting user", 0)
	}

	cfg := &types.Config{}
	err = utils.ReadConfig(cfg, *configPath)
	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	}

	// hardcode db connection details for testing
	cfg.Frontend.ReaderDatabase.Host = "localhost"
	cfg.Frontend.ReaderDatabase.Port = "5432"
	cfg.Frontend.ReaderDatabase.Name = "postgres"
	cfg.Frontend.ReaderDatabase.Password = "postgres"
	cfg.Frontend.ReaderDatabase.Username = "postgres"

	cfg.Frontend.WriterDatabase.Host = "localhost"
	cfg.Frontend.WriterDatabase.Port = "5432"
	cfg.Frontend.WriterDatabase.Name = "postgres"
	cfg.Frontend.WriterDatabase.Password = "postgres"
	cfg.Frontend.WriterDatabase.Username = "postgres"

	utils.Config = cfg

	log.InfoWithFields(log.Fields{"config": *configPath, "version": version.Version, "commit": version.GitCommit, "chainName": utils.Config.Chain.ClConfig.ConfigName}, "starting")

	dataAccessor = dataaccess.NewDataAccessService(cfg)
	router := api.NewApiRouter(dataAccessor, cfg)

	ts = &testServer{httptest.NewTLSServer(router)}
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err, "error creating cookie jar", 0)
	}
	ts.Server.Client().Jar = jar
}

func TestInternalGetProductSummaryHandler(t *testing.T) {
	code, body := ts.get(t, "/api/i/product-summary")
	assert.Equal(t, http.StatusOK, code)

	respData := api_types.InternalGetProductSummaryResponse{}
	err := json.Unmarshal([]byte(body), &respData)
	assert.Nil(t, err, "error unmarshalling response")
	assert.NotEqual(t, 0, respData.Data.ValidatorsPerDashboardLimit, "ValidatorsPerDashboardLimit should not be 0")
	assert.NotEqual(t, 0, len(respData.Data.ApiProducts), "ApiProducts should not be empty")
	assert.NotEqual(t, 0, len(respData.Data.ExtraDashboardValidatorsPremiumAddon), "ExtraDashboardValidatorsPremiumAddon should not be empty")
	assert.NotEqual(t, 0, len(respData.Data.PremiumProducts), "PremiumProducts should not be empty")
}

func TestInternalGetLatestStateHandler(t *testing.T) {
	code, body := ts.get(t, "/api/i/latest-state")
	assert.Equal(t, http.StatusOK, code)

	respData := api_types.InternalGetLatestStateResponse{}
	err := json.Unmarshal([]byte(body), &respData)
	assert.Nil(t, err, "error unmarshalling response")
	assert.NotEqual(t, uint64(0), respData.Data.LatestSlot, "latest slot should not be 0")
	assert.NotEqual(t, uint64(0), respData.Data.FinalizedEpoch, "finalized epoch should not be 0")
}

func TestInternalLoginHandler(t *testing.T) {
	t.Run("login with email in wrong format", func(t *testing.T) {
		code, body := ts.post(t, "/api/i/login", bytes.NewBuffer([]byte(`{"email": "admin", "password": "admin"}`)))
		assert.Equal(t, http.StatusBadRequest, code)
		resp := ts.parseErrorResonse(t, body)
		assert.Equal(t, "email: given value 'admin' has incorrect format", resp.Error, "unexpected error message")
	})
	t.Run("login with correct user and wrong password", func(t *testing.T) {
		code, body := ts.post(t, "/api/i/login", bytes.NewBufferString(`{"email": "admin@admin.com", "password": "wrong"}`))
		assert.Equal(t, http.StatusUnauthorized, code, "login should not be successful")
		resp := ts.parseErrorResonse(t, body)
		assert.Equal(t, "unauthorized: invalid email or password", resp.Error, "unexpected error message")
	})

	t.Run("login with correct user and password", func(t *testing.T) {
		code, _ := ts.post(t, "/api/i/login", bytes.NewBufferString(`{"email": "admin@admin.com", "password": "admin"}`))
		assert.Equal(t, http.StatusOK, code, "login should be successful")
	})

	t.Run("check if user is logged in and has a valid session", func(t *testing.T) {
		code, body := ts.get(t, "/api/i/users/me")
		assert.Equal(t, http.StatusOK, code, "call to users/me should be successful")

		meResponse := &api_types.InternalGetUserInfoResponse{}
		err := json.Unmarshal([]byte(body), meResponse)
		assert.Nil(t, err, "error unmarshalling response")
		// check if email is censored
		assert.Equal(t, meResponse.Data.Email, "a***n@a***n.com", "email should be a***n@a***n.com")
	})

	t.Run("check if logout works", func(t *testing.T) {
		code, _ := ts.post(t, "/api/i/logout", bytes.NewBufferString(``))
		assert.Equal(t, http.StatusOK, code, "logout should be successful")
	})
	t.Run("// check if user is logged out", func(t *testing.T) {
		code, _ := ts.get(t, "/api/i/users/me")
		assert.Equal(t, http.StatusUnauthorized, code, "call to users/me should be unauthorized")
	})
}

func TestInternalSearchHandler(t *testing.T) {
	// search for validator with index 5
	code, body := ts.post(t, "/api/i/search", bytes.NewBufferString(`
	{
		"input":"5",
		"networks":[
			17000
		],
		"types":[
			"validators_by_deposit_ens_name",
			"validators_by_deposit_address",
			"validators_by_withdrawal_ens_name",
			"validators_by_withdrawal_address",
			"validators_by_withdrawal_credential",
			"validator_by_index",
			"validator_by_public_key",
			"validators_by_graffiti"
		]
	}`))
	assert.Equal(t, http.StatusOK, code)

	resp := api_types.InternalPostSearchResponse{}
	err := json.Unmarshal([]byte(body), &resp)
	assert.Nil(t, err, "error unmarshalling response")
	assert.NotEqual(t, 0, len(resp.Data), "response data should not be empty")
	assert.NotNil(t, resp.Data[0].NumValue, "validator index should not be nil")
	assert.Equal(t, uint64(5), *resp.Data[0].NumValue, "validator index should be 5")

	// search for validator by pubkey
	code, body = ts.post(t, "/api/i/search", bytes.NewBufferString(`
	{
		"input":"0x9699af2bad9826694a480cb523cbe545dc41db955356b3b0d4871f1cf3e4924ae4132fa8c374a0505ae2076d3d65b3e0",
		"networks":[
			17000
		],
		"types":[
			"validators_by_deposit_ens_name",
			"validators_by_deposit_address",
			"validators_by_withdrawal_ens_name",
			"validators_by_withdrawal_address",
			"validators_by_withdrawal_credential",
			"validator_by_index",
			"validator_by_public_key",
			"validators_by_graffiti"
		]
	}`))
	assert.Equal(t, http.StatusOK, code)

	resp = api_types.InternalPostSearchResponse{}
	err = json.Unmarshal([]byte(body), &resp)
	assert.Nil(t, err, "error unmarshalling response")
	assert.NotEqual(t, 0, len(resp.Data), "response data should not be empty")
	assert.NotNil(t, resp.Data[0].NumValue, "validator index should not be nil")
	assert.Equal(t, uint64(5), *resp.Data[0].NumValue, "validator index should be 5")

	// search for validator by withdawal address
	code, body = ts.post(t, "/api/i/search", bytes.NewBufferString(`
	{
		"input":"0x0e5dda855eb1de2a212cd1f62b2a3ee49d20c444",
		"networks":[
			17000
		],
		"types":[
			"validators_by_deposit_ens_name",
			"validators_by_deposit_address",
			"validators_by_withdrawal_ens_name",
			"validators_by_withdrawal_address",
			"validators_by_withdrawal_credential",
			"validator_by_index",
			"validator_by_public_key",
			"validators_by_graffiti"
		]
	}`))
	assert.Equal(t, http.StatusOK, code)

	resp = api_types.InternalPostSearchResponse{}
	err = json.Unmarshal([]byte(body), &resp)
	assert.Nil(t, err, "error unmarshalling response")
	assert.NotEqual(t, 0, len(resp.Data), "response data should not be empty")
	assert.NotNil(t, resp.Data[0].NumValue, "validator index should not be nil")
	assert.Greater(t, *resp.Data[0].NumValue, uint64(0), "returned number of validators should be greater than 0")
}

func TestSlotVizHandler(t *testing.T) {
	code, body := ts.get(t, "/api/i/validator-dashboards/NQ/slot-viz")
	assert.Equal(t, http.StatusOK, code)

	resp := api_types.GetValidatorDashboardSlotVizResponse{}
	err := json.Unmarshal([]byte(body), &resp)
	assert.Nil(t, err, "error unmarshalling response")
	assert.Equal(t, 4, len(resp.Data), "response data should contain the last 4 epochs")

	headStateCount := 0
	for _, epoch := range resp.Data {
		if epoch.State == "head" { // count the amount of head epochs returned, should be exactly 1
			headStateCount++
		}
		attestationAssignments := 0
		assert.Equal(t, 32, len(epoch.Slots), "each epoch should contain 32 slots")

		for _, slot := range epoch.Slots {
			if slot.Attestations != nil { // count the amount of attestation assignments for each epoch, should be exactly 1
				attestationAssignments++
			}
		}

		assert.Equal(t, attestationAssignments, 1, "epoch should have exactly one attestation assignment")
	}
	assert.Equal(t, 1, headStateCount, "one of the last 4 epochs should be in head state")
}
