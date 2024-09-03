package handlers_test

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
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
	"github.com/gorilla/mux"
)

var router *mux.Router
var dataAccessor dataaccess.DataAccessor

func TestMain(m *testing.M) {
	setup()
	defer teardown()

	// wait till service initialization is completed (TODO: find a better way to do this)
	// time.Sleep(30 * time.Second)

	os.Exit(m.Run())
}

func teardown() {
	dataAccessor.Close()
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
	router = api.NewApiRouter(dataAccessor, cfg)
}

func TestInternalGetProductSummaryHandler(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/i/product-summary", nil)

	router.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	respData := api_types.InternalGetProductSummaryResponse{}
	err = json.Unmarshal(data, &respData)
	if err != nil {
		log.Infof("%s", string(data))
		t.Fatal(err)
	}

	if respData.Data.ValidatorsPerDashboardLimit == 0 {
		t.Fatal("ValidatorsPerDashboardLimit is 0")
	}

	if len(respData.Data.ApiProducts) == 0 {
		t.Fatal("ApiProducts length is 0")
	}
}

func TestInternalGetLatestStateHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/i/latest-state", nil)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		t.Fatal(err)
	}

	respData := api_types.InternalGetLatestStateResponse{}
	err = json.Unmarshal(data, &respData)
	if err != nil {
		t.Fatal(err)
	}

	if respData.Data.LatestSlot == 0 {
		t.Fatal("latest slot is 0")
	}

	if respData.Data.FinalizedEpoch == 0 {
		t.Fatal("finalized epoch is 0")
	}
}

func TestInternalPostAdConfigurationsHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/i/ad-configurations", nil)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status code 401, got %d", resp.StatusCode)
	}
}
