package apihandlers

import (
	"errors"
	"fmt"
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
	}
	ReturnInternalServerError(w, err)
}

// --------------------------------------
// Authenication

func (h HandlerService) InternalPostOauthAuthorize(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalPostOauthToken(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalPostApiKeys(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

// --------------------------------------
// Ad Configurations

func (h HandlerService) InternalPostAdConfigurations(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func (h HandlerService) InternalGetAdConfigurations(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalPutAdConfiguration(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalDeleteAdConfiguration(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

// --------------------------------------
// Dashboards

func (h HandlerService) InternalGetUserDashboards(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

// --------------------------------------
// Account Dashboards

func (h HandlerService) InternalPostAccountDashboards(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func (h HandlerService) InternalGetAccountDashboard(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalDeleteAccountDashboard(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func (h HandlerService) InternalPostAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func (h HandlerService) InternalDeleteAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func (h HandlerService) InternalPostAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func (h HandlerService) InternalGetAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalDeleteAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func (h HandlerService) InternalPutAccountDashboardAccount(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalPostAccountDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	ReturnCreated(w, nil)
}

func (h HandlerService) InternalPutAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalDeleteAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	ReturnNoContent(w)
}

func (h HandlerService) InternalGetAccountDashboardTransactions(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalPutAccountDashboardTransactionsSettings(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

// --------------------------------------
// Validator Dashboards

func (h HandlerService) InternalPostValidatorDashboards(w http.ResponseWriter, r *http.Request) {
	userId, err := getUser(r)
	if err != nil {
		ReturnUnauthorized(w, err)
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
	if err := errors.Join(
		CheckNameNotEmpty(req.Name),
		CheckNetwork(req.Network),
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	data, err := h.dai.CreateValidatorDashboard(userId, req.Name, apitypes.Network(1))
	if err != nil {
		ReturnInternalServerError(w, err)
		return
	}
	response := apitypes.ApiResponse{
		Data: data,
	}
	ReturnCreated(w, response)
}

func (h HandlerService) InternalGetValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	err := CheckId(dashboardId)
	if err != nil {
		ReturnBadRequest(w, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardOverview(1, dashboardId)
	if err != nil {
		ReturnInternalServerError(w, err)
		return
	}
	response := apitypes.ApiResponse{
		Data: data,
	}
	ReturnOk(w, response)
}

func (h HandlerService) InternalDeleteValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	if err := CheckId(dashboardId); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	ReturnNoContent(w)
}

func (h HandlerService) InternalPostValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
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
		ReturnConflict(w, errors.New("group limit reached"))
		return
	}
	response := apitypes.ApiResponse{
		Data: apitypes.VDBOverviewGroup{},
	}
	ReturnCreated(w, response)
}

func (h HandlerService) InternalDeleteValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
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

func (h HandlerService) InternalPostValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
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
		ReturnConflict(w, errors.New("dashboard validator limit reached"))
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

func (h HandlerService) InternalDeleteValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
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

func (h HandlerService) InternalPostValidatorDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
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
		ReturnConflict(w, errors.New("public ID limit reached"))
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

func (h HandlerService) InternalPutValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
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

func (h HandlerService) InternalDeleteValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
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

func (h HandlerService) InternalGetValidatorDashboardSlotViz(w http.ResponseWriter, r *http.Request) {
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

func (h HandlerService) InternalGetValidatorDashboardSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	pagingParams, pagingErr := CheckAndGetPaging(r)
	if err := errors.Join(
		CheckId(dashboardId),
		pagingErr,
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}
	//TODO remove line
	fmt.Println(pagingParams)

	var limit uint64
	var cursor, search string
	var sort []apitypes.Sort[apitypes.VDBSummaryTableColumn]
	data, paging, err := h.dai.GetValidatorDashboardSummary(dashboardId, cursor, sort, search, limit)
	if err != nil {
		ReturnInternalServerError(w, err)
		return
	}
	response := apitypes.ApiResponse{
		Data:   data,
		Paging: &paging,
	}
	ReturnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardGroupSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId := vars["dashboard_id"]
	groupIdParam := vars["group_id"]
	if err := errors.Join(
		CheckId(dashboardId),
		CheckId(groupIdParam),
	); err != nil {
		ReturnBadRequest(w, err)
		return
	}

	var groupId uint64
	data, err := h.dai.GetValidatorDashboardGroupSummary(dashboardId, groupId)
	if err != nil {
		ReturnInternalServerError(w, err)
		return
	}
	response := apitypes.ApiResponse{
		Data: data,
	}
	ReturnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardSummaryChart(w http.ResponseWriter, r *http.Request) {
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

func (h HandlerService) InternalGetValidatorDashboardRewards(w http.ResponseWriter, r *http.Request) {
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
	//TODO remove line
	fmt.Println(paging)

	response := apitypes.ApiResponse{
		Data: apitypes.VDBRewardsTableResponse{},
	}
	ReturnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardGroupRewards(w http.ResponseWriter, r *http.Request) {
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

func (h HandlerService) InternalGetValidatorDashboardRewardsChart(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalGetValidatorDashboardDuties(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalGetValidatorDashboardBlocks(w http.ResponseWriter, r *http.Request) {
	var limit uint64
	var dashboardId, cursor, search string
	var sort []apitypes.Sort[apitypes.VDBBlocksTableColumn]
	data, paging, err := h.dai.GetValidatorDashboardBlocks(dashboardId, cursor, sort, search, limit)
	if err != nil {
		ReturnInternalServerError(w, err)
		return
	}
	response := apitypes.ApiResponse{
		Data:   data,
		Paging: &paging,
	}
	ReturnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardHeatmap(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalGetValidatorDashboardGroupHeatmap(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalGetValidatorDashboardExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalGetValidatorDashboardConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}

func (h HandlerService) InternalGetValidatorDashboardWithdrawals(w http.ResponseWriter, r *http.Request) {
	ReturnOk(w, nil)
}
