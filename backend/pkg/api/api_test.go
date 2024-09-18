package api_test

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"os/exec"
	"slices"
	"sort"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/gavv/httpexpect/v2"
	"github.com/go-openapi/spec"
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
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var ts *httptest.Server
var dataAccessor dataaccess.DataAccessor
var postgres *embeddedpostgres.EmbeddedPostgres

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		log.Error(err, "error setting up test environment", 0)
		teardown()
		os.Exit(1)
	}
	log.Info("test setup completed")
	code := m.Run()
	teardown()

	// wait till service initialization is completed (TODO: find a better way to do this)
	// time.Sleep(30 * time.Second)

	os.Exit(code)
}

func teardown() {
	if dataAccessor != nil {
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
	}
}

func setup() error {
	configPath := flag.String("config", "", "Path to the config file, if empty string defaults will be used")

	flag.Parse()

	// terminate any currently running postgres instances
	_ = exec.Command("pkill", "-9", "postgres").Run()

	// start embedded postgres
	postgres = embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().Username("postgres"))
	err := postgres.Start()
	if err != nil {
		return fmt.Errorf("error starting embedded postgres: %w", err)
	}

	// connection the the embedded db and run migrations
	tempDb, err := sqlx.Connect("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable")
	if err != nil {
		return fmt.Errorf("error connection to test db: %w", err)
	}

	if err := goose.Up(tempDb.DB, "../../pkg/commons/db/migrations/postgres"); err != nil {
		return fmt.Errorf("error running migrations: %w", err)
	}

	// insert dummy user for testing (email: admin@admin, password: admin)
	pHash, _ := bcrypt.GenerateFromPassword([]byte("admin"), 10)
	_, err = tempDb.Exec(`
      INSERT INTO users (password, email, register_ts, api_key, email_confirmed)
      VALUES ($1, $2, TO_TIMESTAMP($3), $4, $5)`,
		string(pHash), "admin@admin.com", time.Now().Unix(), "admin", true,
	)
	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}

	// required for shared dashboard
	pHash, _ = bcrypt.GenerateFromPassword([]byte("admin"), 10)
	_, err = tempDb.Exec(`
      INSERT INTO users (id, password, email, register_ts, api_key, email_confirmed)
      VALUES ($1, $2, $3, TO_TIMESTAMP($4), $5, $6)`,
		122558, string(pHash), "admin2@admin.com", time.Now().Unix(), "admin2", true,
	)
	if err != nil {
		return fmt.Errorf("error inserting user 2: %w", err)
	}

	// insert dummy api weight for testing
	_, err = tempDb.Exec(`
      INSERT INTO api_weights (bucket, endpoint, method, params, weight, valid_from)
      VALUES ($1, $2, $3, $4, $5, TO_TIMESTAMP($6))`,
		"default", "/api/v2/test-ratelimit", "GET", "", 2, time.Now().Unix(),
	)
	if err != nil {
		return fmt.Errorf("error inserting api weight: %w", err)
	}

	cfg := &types.Config{}
	err = utils.ReadConfig(cfg, *configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
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

	log.Info("initializing data access service")
	dataAccessor = dataaccess.NewDataAccessService(cfg)
	dataAccessor.StartDataAccessServices()

	log.Info("initializing api router")
	router := api.NewApiRouter(dataAccessor, cfg)

	ts = httptest.NewTLSServer(router)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("error creating cookie jar: %w", err)
	}
	ts.Client().Jar = jar

	return nil
}

func getExpectConfig(t *testing.T, ts *httptest.Server) httpexpect.Config {
	return httpexpect.Config{
		BaseURL:  ts.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Jar: httpexpect.NewCookieJar(),
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					// accept any certificate; for testing only!
					//nolint: gosec
					InsecureSkipVerify: true,
				},
			},
		},
		Printers: []httpexpect.Printer{
			httpexpect.NewCurlPrinter(t),
		},
	}
}

func login(e *httpexpect.Expect) {
	e.POST("/api/i/login").
		WithHeader("Content-Type", "application/json").
		WithJSON(map[string]interface{}{"email": "admin@admin.com", "password": "admin"}).
		Expect().
		Status(http.StatusOK)
}

func logout(e *httpexpect.Expect) {
	e.POST("/api/i/logout").
		Expect().
		Status(http.StatusOK)
}

func TestInternalGetProductSummaryHandler(t *testing.T) {
	e := httpexpect.WithConfig(getExpectConfig(t, ts))

	respData := api_types.InternalGetProductSummaryResponse{}
	e.GET("/api/i/product-summary").Expect().Status(http.StatusOK).JSON().Decode(&respData)

	assert.NotEqual(t, 0, respData.Data.ValidatorsPerDashboardLimit, "ValidatorsPerDashboardLimit should not be 0")
	assert.NotEqual(t, 0, len(respData.Data.ApiProducts), "ApiProducts should not be empty")
	assert.NotEqual(t, 0, len(respData.Data.ExtraDashboardValidatorsPremiumAddon), "ExtraDashboardValidatorsPremiumAddon should not be empty")
	assert.NotEqual(t, 0, len(respData.Data.PremiumProducts), "PremiumProducts should not be empty")
}

func TestInternalGetLatestStateHandler(t *testing.T) {
	e := httpexpect.WithConfig(getExpectConfig(t, ts))

	respData := api_types.InternalGetLatestStateResponse{}
	e.GET("/api/i/latest-state").Expect().Status(http.StatusOK).JSON().Decode(&respData)

	assert.NotEqual(t, uint64(0), respData.Data.LatestSlot, "latest slot should not be 0")
	assert.NotEqual(t, uint64(0), respData.Data.FinalizedEpoch, "finalized epoch should not be 0")
}

func TestInternalLoginHandler(t *testing.T) {
	e := httpexpect.WithConfig(getExpectConfig(t, ts))
	t.Run("login with email in wrong format", func(t *testing.T) {
		e.POST("/api/i/login").
			WithHeader("Content-Type", "application/json").
			WithJSON(map[string]interface{}{"email": "admin", "password": "admin"}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().
			Object().
			HasValue("error", "email: given value 'admin' has incorrect format")
	})
	t.Run("login with correct user and wrong password", func(t *testing.T) {
		e.POST("/api/i/login").
			WithHeader("Content-Type", "application/json").
			WithJSON(map[string]interface{}{"email": "admin@admin.com", "password": "wrong"}).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().
			Object().
			HasValue("error", "unauthorized: invalid email or password")
	})

	t.Run("login with correct user and password", func(t *testing.T) {
		login(e)
	})

	t.Run("check if user is logged in and has a valid session", func(t *testing.T) {
		meResponse := &api_types.InternalGetUserInfoResponse{}
		e.GET("/api/i/users/me").
			Expect().
			Status(http.StatusOK).
			JSON().
			Decode(&meResponse)

		// check if email is censored
		assert.Equal(t, meResponse.Data.Email, "a***n@a***n.com", "email should be a***n@a***n.com")
	})

	t.Run("check if logout works", func(t *testing.T) {
		logout(e)
	})
	t.Run("// check if user is logged out", func(t *testing.T) {
		e.GET("/api/i/users/me").
			Expect().
			Status(http.StatusUnauthorized)
	})
}

func TestInternalSearchHandler(t *testing.T) {
	e := httpexpect.WithConfig(getExpectConfig(t, ts))

	// search for validator with index 5
	resp := api_types.InternalPostSearchResponse{}
	e.POST("/api/i/search").
		WithHeader("Content-Type", "application/json").
		WithBytes([]byte(`
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
	}`)).Expect().Status(http.StatusOK).JSON().Decode(&resp)

	assert.NotEqual(t, 0, len(resp.Data), "response data should not be empty")
	assert.NotNil(t, resp.Data[0].NumValue, "validator index should not be nil")
	assert.Equal(t, uint64(5), *resp.Data[0].NumValue, "validator index should be 5")

	// search for validator by pubkey
	resp = api_types.InternalPostSearchResponse{}
	e.POST("/api/i/search").
		WithHeader("Content-Type", "application/json").
		WithBytes([]byte(`
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
	}`)).Expect().Status(http.StatusOK).JSON().Decode(&resp)

	assert.NotEqual(t, 0, len(resp.Data), "response data should not be empty")
	assert.NotNil(t, resp.Data[0].NumValue, "validator index should not be nil")
	assert.Equal(t, uint64(5), *resp.Data[0].NumValue, "validator index should be 5")

	// search for validator by withdawal address
	resp = api_types.InternalPostSearchResponse{}
	e.POST("/api/i/search").
		WithHeader("Content-Type", "application/json").
		WithBytes([]byte(`{
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
	}`)).Expect().Status(http.StatusOK).JSON().Decode(&resp)

	assert.NotEqual(t, 0, len(resp.Data), "response data should not be empty")
	assert.NotNil(t, resp.Data[0].NumValue, "validator index should not be nil")
	assert.Greater(t, *resp.Data[0].NumValue, uint64(0), "returned number of validators should be greater than 0")
}

func TestPublicAndSharedDashboards(t *testing.T) {
	t.Parallel()
	e := httpexpect.WithConfig(getExpectConfig(t, ts))

	dashboardIds := []struct {
		id       string
		isShared bool
	}{
		{id: "NQ", isShared: false},
		{id: "MSwxNTU2MSwxNTY", isShared: false},
		{id: "v-80d7edaa-74fb-4129-a41e-7700756961cf", isShared: true},
	}

	for _, dashboardId := range dashboardIds {
		t.Run(fmt.Sprintf("[%s]: test slot viz", dashboardId.id), func(t *testing.T) {
			resp := api_types.GetValidatorDashboardSlotVizResponse{}
			e.GET("/api/i/validator-dashboards/{id}/slot-viz", dashboardId.id).
				Expect().
				Status(http.StatusOK).
				JSON().Decode(&resp)

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

				assert.GreaterOrEqual(t, attestationAssignments, 1, "epoch should have at least one attestation assignment")
			}
			assert.Equal(t, 1, headStateCount, "one of the last 4 epochs should be in head state")
		})

		t.Run(fmt.Sprintf("[%s]: test dashboard overview", dashboardId.id), func(t *testing.T) {
			resp := api_types.GetValidatorDashboardResponse{}
			e.GET("/api/i/validator-dashboards/{id}", dashboardId.id).
				Expect().
				Status(http.StatusOK).
				JSON().Decode(&resp)

			numValidators := resp.Data.Validators.Exited + resp.Data.Validators.Offline + resp.Data.Validators.Pending + resp.Data.Validators.Online + resp.Data.Validators.Slashed
			assert.Greater(t, numValidators, uint64(0), "dashboard should contain at least one validator")
			assert.Greater(t, len(resp.Data.Groups), 0, "dashboard should contain at least one group")
		})

		t.Run(fmt.Sprintf("[%s]: test group summary", dashboardId.id), func(t *testing.T) {
			resp := api_types.GetValidatorDashboardSummaryResponse{}
			e.GET("/api/i/validator-dashboards/{id}/summary", dashboardId.id).
				WithQuery("period", "last_24h").
				WithQuery("limit", "10").
				WithQuery("sort", "efficiency:desc").
				Expect().Status(http.StatusOK).JSON().Decode(&resp)

			assert.Greater(t, len(resp.Data), 0, "dashboard should contain at least one group summary row")

			t.Run(fmt.Sprintf("[%s / %d]: test group details", dashboardId.id, resp.Data[0].GroupId), func(t *testing.T) {
				groupResp := api_types.GetValidatorDashboardGroupSummaryResponse{}
				e.GET("/api/i/validator-dashboards/{id}/groups/{groupId}/summary", dashboardId.id, resp.Data[0].GroupId).
					WithQuery("period", "all_time").
					Expect().
					Status(http.StatusOK).
					JSON().Decode(&groupResp)

				assert.Greater(t, groupResp.Data.AttestationsHead.Success+groupResp.Data.AttestationsHead.Failed, uint64(0), "group should have at least head attestation")
				assert.Greater(t, groupResp.Data.AttestationsSource.Success+groupResp.Data.AttestationsSource.Failed, uint64(0), "group should have at least source attestation")
				assert.Greater(t, groupResp.Data.AttestationsTarget.Success+groupResp.Data.AttestationsTarget.Failed, uint64(0), "group should have at least target attestation")
			})
		})

		t.Run(fmt.Sprintf("[%s]: test group summary chart", dashboardId.id), func(t *testing.T) {
			resp := api_types.GetValidatorDashboardSummaryChartResponse{}
			e.GET("/api/i/validator-dashboards/{id}/summary-chart", dashboardId.id).
				WithQuery("aggregation", "hourly").
				WithQuery("before_ts", time.Now().Unix()).
				WithQuery("efficiency_type", "all").
				WithQuery("group_ids", "-1,-2").
				Expect().Status(http.StatusOK).JSON().Decode(&resp)

			assert.Greater(t, len(resp.Data.Categories), 0, "group summary chart categories should not be empty")
			assert.Greater(t, len(resp.Data.Series), 0, "group summary chart series should not be empty")
		})

		t.Run(fmt.Sprintf("[%s]: test rewards", dashboardId.id), func(t *testing.T) {
			resp := api_types.GetValidatorDashboardRewardsResponse{}
			e.GET("/api/i/validator-dashboards/{id}/rewards", dashboardId.id).
				WithQuery("limit", 10).
				WithQuery("sort", "epoch:desc").
				Expect().Status(http.StatusOK).JSON().Decode(&resp)

			assert.Greater(t, len(resp.Data), 0, "rewards response should not be empty")
			assert.LessOrEqual(t, len(resp.Data), 10, "rewards response should not contain more than 10 entries")
			assert.True(t, sort.SliceIsSorted(resp.Data, func(i, j int) bool {
				return resp.Data[i].Epoch > resp.Data[j].Epoch
			}), "rewards should be sorted by epoch in descending order")

			resp = api_types.GetValidatorDashboardRewardsResponse{}
			e.GET("/api/i/validator-dashboards/{id}/rewards", dashboardId.id).
				WithQuery("limit", 10).
				WithQuery("sort", "epoch:asc").
				Expect().Status(http.StatusOK).JSON().Decode(&resp)
			assert.Greater(t, len(resp.Data), 0, "rewards response should not be empty")
			assert.LessOrEqual(t, len(resp.Data), 10, "rewards response should not contain more than 10 entries")
			assert.True(t, sort.SliceIsSorted(resp.Data, func(i, j int) bool {
				return resp.Data[i].Epoch < resp.Data[j].Epoch
			}), "rewards should be sorted by epoch in ascending order")

			rewardDetails := api_types.GetValidatorDashboardGroupRewardsResponse{}
			e.GET("/api/i/validator-dashboards/{id}/groups/{group_id}/rewards/{epoch}", dashboardId.id, resp.Data[0].GroupId, resp.Data[0].Epoch).
				WithQuery("limit", 10).
				WithQuery("sort", "epoch:asc").
				Expect().Status(http.StatusOK).JSON().Decode(&rewardDetails)
		})

		t.Run(fmt.Sprintf("[%s]: test rewards chart", dashboardId.id), func(t *testing.T) {
			resp := api_types.GetValidatorDashboardRewardsChartResponse{}
			e.GET("/api/i/validator-dashboards/{id}/rewards-chart", dashboardId.id).
				Expect().Status(http.StatusOK).JSON().Decode(&resp)

			assert.Greater(t, len(resp.Data.Categories), 0, "rewards chart categories should not be empty")
			assert.Greater(t, len(resp.Data.Series), 0, "rewards chart series should not be empty")
		})
	}
}

func TestApiDoc(t *testing.T) {
	e := httpexpect.WithConfig(getExpectConfig(t, ts))

	t.Run("test api doc json", func(t *testing.T) {
		resp := spec.Swagger{}
		e.GET("/api/v2/docs/swagger.json").
			Expect().
			Status(http.StatusOK).JSON().Decode(&resp)

		assert.Equal(t, "/api/v2", resp.BasePath, "swagger base path should be '/api/v2'")
		require.NotNil(t, 0, resp.Paths, "swagger paths should not nil")
		assert.NotEqual(t, 0, len(resp.Paths.Paths), "swagger paths should not be empty")
		assert.NotEqual(t, 0, len(resp.Definitions), "swagger definitions should not be empty")
		assert.NotEqual(t, 0, len(resp.Host), "swagger host should not be empty")
	})

	t.Run("test api ratelimit weights endpoint", func(t *testing.T) {
		resp := api_types.InternalGetRatelimitWeightsResponse{}
		e.GET("/api/i/ratelimit-weights").
			Expect().
			Status(http.StatusOK).JSON().Decode(&resp)

		assert.GreaterOrEqual(t, len(resp.Data), 1, "ratelimit weights should contain at least one entry")
		testEndpointIndex := slices.IndexFunc(resp.Data, func(item api_types.ApiWeightItem) bool {
			return item.Endpoint == "/api/v2/test-ratelimit"
		})
		assert.GreaterOrEqual(t, testEndpointIndex, 0, "ratelimit weights should contain an entry for /api/v2/test-ratelimit")
	})
}
