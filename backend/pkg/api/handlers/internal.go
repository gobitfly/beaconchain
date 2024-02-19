package apihandlers

import (
	"errors"
	"net/http"
	"strings"

	apitypes "github.com/gobitfly/beaconchain/pkg/types/api"

	"github.com/gorilla/mux"
)

// All handler function names must include the HTTP method and the path they handle
// Internal handlers may only be authenticated by an OAuth token

func handleJsonError(w http.ResponseWriter, err error) {
	if err, ok := err.(RequestError); ok {
		switch err.StatusCode {
		case http.StatusBadRequest:
			ReturnBadRequest(w, err.Err)
		case http.StatusInternalServerError:
			fallthrough
		default:
			ReturnInternalServerError(w, err.Err)
		}
	} else {
		ReturnInternalServerError(w, err)
	}
}

func InternalPostOauthAuthorize(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalPostOauthToken(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalPostApiKeys(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalPostAdConfigurations(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func InternalGetAdConfigurations(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalPutAdConfiguration(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalDeleteAdConfiguration(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func InternalGetDashboards(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalPostAccountDashboards(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func InternalGetAccountDashboard(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalDeleteAccountDashboard(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func InternalPostAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func InternalDeleteAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func InternalPostAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func InternalGetAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalDeleteAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func InternalPutAccountDashboardAccount(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalPostAccountDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func InternalPutAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalDeleteAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func InternalGetAccountDashboardTransactions(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalPutAccountDashboardTransactionsSettings(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalPostValidatorDashboards(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Name    string `json:"name"`
		Network string `json:"network"`
	}{}
	if err := CheckAndGetJson(r.Body, &req); err != nil {
		handleJsonError(w, err)
		return
	}
	if err := errors.Join(
		CheckNameNotEmpty(req.Name),
		CheckNetwork(req.Network),
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	// get data from backend
	success := false
	if success {
		response := apitypes.ApiResponse{
			Data: struct {
				Id      string `json:"id"`
				Network uint64 `json:"network"`
				Name    string `json:"name"`
				// CreatedAt time.Time `json:created_at`
			}{"01_981723", 1, req.Name /*, time.Now()*/},
		}
		ReturnCreated(w, response)
	} else {
		ReturnInternalServerError(w, errors.New("General error"))
	}
}

func InternalGetValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	err := CheckId(dashboardId)
	if err != nil {
		ReturnBadRequest(w, err)
		return
	}
	// get data from backend
	response := apitypes.ApiResponse{
		Data: apitypes.VDBOverviewResponse{},
	}
	ReturnOk(w, response)
}

func InternalDeleteValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	if err := CheckId(dashboardId); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	ReturnNoContent(w)
}

func InternalPostValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	req := struct {
		Name string `json:"name"`
	}{}
	if err := CheckAndGetJson(r.Body, &req); err != nil {
		handleJsonError(w, err)
		return
	}
	if err := errors.Join(
		CheckNameNotEmpty(req.Name),
		CheckId(dashboardId),
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	if false {
		ReturnConflict(w, errors.New("Group limit reached"))
		return
	}
	response := apitypes.ApiResponse{
		Data: apitypes.VDBOverviewGroup{},
	}
	ReturnCreated(w, response)
}

func InternalDeleteValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	groupId := vars["group_id"]
	if err := errors.Join(
		CheckId(dashboardId),
		CheckId(groupId),
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	ReturnNoContent(w)
}

func InternalPostValidatorDashboardvalidators(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	req := struct {
		Validators []string `json:"validators"`
		GroupId    string   `json:"group_id,omitempty"`
	}{}
	if err := CheckAndGetJson(r.Body, &req); err != nil {
		handleJsonError(w, err)
		return
	}
	if err := errors.Join(
		CheckId(dashboardId),
		CheckValidatorList(req.Validators),
		CheckId(req.GroupId),
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	if false {
		ReturnConflict(w, errors.New("Dashboard validator limit reached"))
		return
	}
	response := apitypes.ApiResponse{
		Data: []struct {
			PubKey  []string `json:"public_key"`
			GroupId string   `json:"group_id"`
		}{},
	}
	ReturnCreated(w, response)
}

func InternalDeleteValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	q := r.URL.Query()
	validators := strings.Split(q.Get("validators"), ",")
	if err := errors.Join(
		CheckId(dashboardId),
		CheckValidatorList(validators),
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	ReturnNoContent(w)
}

func InternalPostValidatorDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	req := struct {
		Name          string `json:"name"`
		ShareSettings struct {
			GroupNames bool `json:"group_names"`
		} `json:"share_settings"`
	}{}
	if err := CheckAndGetJson(r.Body, &req); err != nil {
		handleJsonError(w, err)
		return
	}
	if err := errors.Join(
		CheckNameNotEmpty(req.Name),
		CheckId(dashboardId),
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	if false {
		ReturnConflict(w, errors.New("Public ID limit reached"))
		return
	}
	response := apitypes.ApiResponse{
		Data: struct {
			AccessToken   string `json:"access_token"`
			Name          string `json:"name"`
			ShareSettings struct {
				GroupNames bool `json:"group_names"`
			} `json:"share_settings"`
		}{},
	}
	ReturnCreated(w, response)
}

func InternalPutValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	publicDashboardId := vars["public_dashboard_id"]
	req := struct {
		Name          string `json:"name"`
		ShareSettings struct {
			GroupNames bool `json:"group_names"`
		} `json:"share_settings"`
	}{}
	if err := CheckAndGetJson(r.Body, &req); err != nil {
		handleJsonError(w, err)
		return
	}
	if err := errors.Join(
		CheckNameNotEmpty(req.Name),
		CheckId(dashboardId),
		CheckId(publicDashboardId),
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	response := apitypes.ApiResponse{
		Data: struct {
			AccessToken   string `json:"access_token"`
			ShareSettings struct {
				GroupNames bool `json:"group_names"`
			} `json:"share_settings"`
		}{},
	}
	ReturnOk(w, response)
}

func InternalDeleteValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	publicDashboardId := vars["public_dashboard_id"]
	if err := errors.Join(
		CheckId(dashboardId),
		CheckId(publicDashboardId),
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	ReturnNoContent(w)
}

func InternalGetValidatorDashboardSlotViz(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]

	if err := CheckId(dashboardId); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	response := apitypes.ApiResponse{
		Data: apitypes.VDBSlotVizResponse{},
	}
	ReturnOk(w, response)
}

func InternalGetValidatorDashboardSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	paging, paging_error := CheckAndGetPaging(r)
	if err := errors.Join(
		CheckId(dashboardId),
		paging_error,
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	response := apitypes.ApiResponse{
		Data: apitypes.VDBSummaryTableResponse{},
	}
	ReturnOk(w, response)
}

func InternalGetValidatorDashboardGroupSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	groupId := vars["group_id"]
	if err := errors.Join(
		CheckId(dashboardId),
		CheckId(groupId),
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	response := apitypes.ApiResponse{
		Data: apitypes.VDBGroupSummaryResponse{},
	}
	ReturnOk(w, response)
}

func InternalGetValidatorDashboardSummaryChart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	if err := CheckId(dashboardId); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	response := apitypes.ApiResponse{
		Data: nil, // apitypes.VDBSummaryChartResponse{},
	}
	ReturnOk(w, response)
}

func InternalGetValidatorDashboardRewards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	paging, paging_error := CheckAndGetPaging(r)
	if err := errors.Join(
		CheckId(dashboardId),
		paging_error,
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	response := apitypes.ApiResponse{
		Data: apitypes.VDBRewardsTableResponse{},
	}
	ReturnOk(w, response)
}

func InternalGetValidatorDashboardGroupRewards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	groupId := vars["group_id"]
	if err := errors.Join(
		CheckId(dashboardId),
		CheckId(groupId),
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	response := apitypes.ApiResponse{
		Data: apitypes.VDBGroupRewardsResponse{},
	}
	ReturnOk(w, response)
}

func InternalGetValidatorDashboardRewardsChart(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalGetValidatorDashboardDuties(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalValidatorDashboardBlocks(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalValidatorDashboardHeatmap(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalValidatorDashboardGroupHeatmap(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalValidatorDashboardExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalValidatorDashboardConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func InternalValidatorDashboardWithdrawals(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}
