package handlers

import (
	"net/http"
	"strings"

	types "github.com/gobitfly/beaconchain/pkg/api/types"

	"github.com/gorilla/mux"
)

// All handler function names must include the HTTP method and the path they handle
// Internal handlers may only be authenticated by an OAuth token

func handleJsonError(w http.ResponseWriter, err error) {
	if err, ok := err.(RequestError); ok {
		switch err.StatusCode {
		case http.StatusBadRequest:
			returnBadRequest(w, err.Err)
		case http.StatusInternalServerError:
			fallthrough
		default:
			returnInternalServerError(w, err.Err)
		}
	}
	returnInternalServerError(w, err)
}

// --------------------------------------
// Authenication

func (h HandlerService) InternalPostOauthAuthorize(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalPostOauthToken(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalPostApiKeys(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

// --------------------------------------
// Ad Configurations

func (h HandlerService) InternalPostAdConfigurations(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func (h HandlerService) InternalGetAdConfigurations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalPutAdConfiguration(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalDeleteAdConfiguration(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

// --------------------------------------
// Dashboards

func (h HandlerService) InternalGetUserDashboards(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

// --------------------------------------
// Account Dashboards

func (h HandlerService) InternalPostAccountDashboards(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func (h HandlerService) InternalGetAccountDashboard(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalDeleteAccountDashboard(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h HandlerService) InternalPostAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func (h HandlerService) InternalDeleteAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h HandlerService) InternalPostAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func (h HandlerService) InternalGetAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalDeleteAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h HandlerService) InternalPutAccountDashboardAccount(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalPostAccountDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func (h HandlerService) InternalPutAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalDeleteAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h HandlerService) InternalGetAccountDashboardTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalPutAccountDashboardTransactionsSettings(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

// --------------------------------------
// Validator Dashboards

func (h HandlerService) InternalPostValidatorDashboards(w http.ResponseWriter, r *http.Request) {
	var err error
	userId, err := getUser(r)
	if err != nil {
		returnUnauthorized(w, err)
		return
	}
	req := struct {
		Name    string `json:"name"`
		Network string `json:"network"`
	}{}
	if err := CheckAndGetJson(r.Body, &req); err != nil {
		handleJsonError(w, err)
		return
	}
	name := checkNameNotEmpty(&err, req.Name)
	network := checkNetwork(&err, req.Network)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	data, err := h.dai.CreateValidatorDashboard(userId, name, network)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}
	ReturnCreated(w, response)
}

func (h HandlerService) InternalGetValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var err error
	dashboardId := checkUint(&err, vars["dashboard_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardOverview(1, dashboardId)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h HandlerService) InternalDeleteValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkUint(&err, vars["dashboard_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	TODO_RemoveThisLine(dashboardId)

	returnNoContent(w)
}

func (h HandlerService) InternalPostValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	var err error
	req := struct {
		Name string `json:"name"`
	}{}
	if err := CheckAndGetJson(r.Body, &req); err != nil {
		handleJsonError(w, err)
		return
	}
	vars := mux.Vars(r)
	dashboardId := checkUint(&err, vars["dashboard_id"])
	name := checkNameNotEmpty(&err, req.Name)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	TODO_RemoveThisLine(dashboardId, name)

	// TODO check group limit reached

	response := types.ApiResponse{
		Data: types.VDBOverviewGroup{},
	}

	ReturnCreated(w, response)
}

func (h HandlerService) InternalDeleteValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var err error
	dashboardId := checkUint(&err, vars["dashboard_id"])
	groupId := checkId(&err, vars["group_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	TODO_RemoveThisLine(dashboardId, groupId)

	returnNoContent(w)
}

func (h HandlerService) InternalPostValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var err error
	req := struct {
		Validators []string `json:"validators"`
		GroupId    string   `json:"group_id,omitempty"`
	}{}
	if err := CheckAndGetJson(r.Body, &req); err != nil {
		handleJsonError(w, err)
		return
	}
	vars := mux.Vars(r)
	dashboardId := checkUint(&err, vars["dashboard_id"])
	groupId := checkId(&err, req.GroupId)
	validators := CheckValidatorList(&err, req.Validators)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	TODO_RemoveThisLine(dashboardId, groupId, validators)

	// TODO check validator limit reached

	response := types.ApiResponse{
		Data: []struct {
			PubKey  []string `json:"public_key"`
			GroupId string   `json:"group_id"`
		}{},
	}
	ReturnCreated(w, response)
}

func (h HandlerService) InternalDeleteValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var err error
	dashboardId := checkUint(&err, vars["dashboard_id"])
	validators := CheckValidatorList(&err, strings.Split(r.URL.Query().Get("validators"), ","))
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	TODO_RemoveThisLine(dashboardId, validators)

	returnNoContent(w)
}

func (h HandlerService) InternalPostValidatorDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	var err error
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
	vars := mux.Vars(r)
	dashboardId := checkUint(&err, vars["dashboard_id"])
	name := checkNameNotEmpty(&err, req.Name)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	TODO_RemoveThisLine(dashboardId, name)

	//TODO check public id limit reached

	response := types.ApiResponse{
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

func (h HandlerService) InternalPutValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
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
	var err error
	dashboardId := checkUint(&err, vars["dashboard_id"])
	publicDashboardId := checkId(&err, vars["public_dashboard_id"])
	name := checkNameNotEmpty(&err, req.Name)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	TODO_RemoveThisLine(dashboardId, publicDashboardId, name)

	response := types.ApiResponse{
		Data: struct {
			AccessToken   string `json:"access_token"`
			ShareSettings struct {
				GroupNames bool `json:"group_names"`
			} `json:"share_settings"`
		}{},
	}
	returnOk(w, response)
}

func (h HandlerService) InternalDeleteValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkUint(&err, vars["dashboard_id"])
	publicDashboardId := checkId(&err, vars["public_dashboard_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	TODO_RemoveThisLine(dashboardId, publicDashboardId)

	returnNoContent(w)
}

func (h HandlerService) InternalGetValidatorDashboardSlotViz(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkUint(&err, vars["dashboard_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	TODO_RemoveThisLine(dashboardId)

	data, err := h.dai.GetValidatorDashboardSlotViz(dashboardId)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardSummary(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkUint(&err, vars["dashboard_id"])
	pagingParams := checkAndGetPaging(&err, r)
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	var sort []types.Sort[types.VDBSummaryTableColumn]
	data, paging, err := h.dai.GetValidatorDashboardSummary(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.ApiResponse{
		Data:   data,
		Paging: &paging,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardGroupSummary(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkUint(&err, vars["dashboard_id"])
	groupId := checkUint(&err, vars["group_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardGroupSummary(dashboardId, groupId)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardSummaryChart(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkUint(&err, vars["dashboard_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	TODO_RemoveThisLine(dashboardId)

	response := types.ApiResponse{
		Data: nil, // apitypes.VDBSummaryChartResponse{},
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardRewards(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkUint(&err, vars["dashboard_id"])
	paging := checkAndGetPaging(&err, r)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	TODO_RemoveThisLine(dashboardId, paging)

	response := types.ApiResponse{
		Data: types.VDBRewardsTableResponse{},
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardGroupRewards(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkUint(&err, vars["dashboard_id"])
	groupId := checkId(&err, vars["group_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	TODO_RemoveThisLine(dashboardId, groupId)

	response := types.ApiResponse{
		Data: types.VDBGroupRewardsResponse{},
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardRewardsChart(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalGetValidatorDashboardDuties(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalGetValidatorDashboardBlocks(w http.ResponseWriter, r *http.Request) {
	var dashboardId, limit uint64
	var cursor, search string
	var sort []types.Sort[types.VDBBlocksTableColumn]
	data, paging, err := h.dai.GetValidatorDashboardBlocks(dashboardId, cursor, sort, search, limit)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.ApiResponse{
		Data:   data,
		Paging: &paging,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardHeatmap(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalGetValidatorDashboardGroupHeatmap(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalGetValidatorDashboardExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalGetValidatorDashboardConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalGetValidatorDashboardWithdrawals(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func TODO_RemoveThisLine(args ...interface{}) {
	// This function is used to prevent the "declared and not used" error
	// Temporary solution until the code is complete
}
