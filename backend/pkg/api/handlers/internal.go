package handlers

import (
	"errors"
	"net/http"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	types "github.com/gobitfly/beaconchain/pkg/api/types"

	"github.com/gorilla/mux"
)

// All handler function names must include the HTTP method and the path they handle
// Internal handlers may only be authenticated by an OAuth token

// --------------------------------------
// Ad Configurations

func (h *HandlerService) InternalPostAdConfigurations(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) InternalGetAdConfigurations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) InternalPutAdConfiguration(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) InternalDeleteAdConfiguration(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

// --------------------------------------
// Dashboards

func (h *HandlerService) InternalGetUserDashboards(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUser(r)
	if err != nil {
		returnUnauthorized(w, err)
		return
	}
	data, err := h.dai.GetUserDashboards(user.Id)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.ApiDataResponse[types.UserDashboardsData]{
		Data: *data,
	}
	returnOk(w, response)
}

// --------------------------------------
// Account Dashboards

func (h *HandlerService) InternalPostAccountDashboards(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) InternalGetAccountDashboard(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) InternalDeleteAccountDashboard(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h *HandlerService) InternalPostAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) InternalDeleteAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h *HandlerService) InternalPostAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) InternalGetAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) InternalDeleteAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h *HandlerService) InternalPutAccountDashboardAccount(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) InternalPostAccountDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h *HandlerService) InternalPutAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) InternalDeleteAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h *HandlerService) InternalGetAccountDashboardTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) InternalPutAccountDashboardTransactionsSettings(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

// --------------------------------------
// Validator Dashboards

var errMsgParsingId = errors.New("error parsing parameter 'dashboard_id'")

func (h *HandlerService) InternalPostValidatorDashboards(w http.ResponseWriter, r *http.Request) {
	var err error
	user, err := h.getUser(r)
	if err != nil {
		returnUnauthorized(w, err)
		return
	}
	req := struct {
		Name    string `json:"name"`
		Network string `json:"network"`
	}{}
	if bodyErr := checkBody(&err, &req, r.Body); bodyErr != nil {
		returnInternalServerError(w, bodyErr)
		return
	}
	name := checkNameNotEmpty(&err, req.Name)
	network := checkNetwork(&err, req.Network)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	data, err := h.dai.CreateValidatorDashboard(user.Id, name, network)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}
	returnCreated(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardOverview(*dashboardId)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardResponse{
		Data: *data,
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalDeleteValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId := checkDashboardPrimaryId(&err, mux.Vars(r)["dashboard_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	err = h.dai.RemoveValidatorDashboard(dashboardId)
	if err != nil {
		handleError(w, err)
		return
	}
	returnNoContent(w)
}

func (h *HandlerService) InternalPostValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId := checkDashboardPrimaryId(&err, mux.Vars(r)["dashboard_id"])
	req := struct {
		Name string `json:"name"`
	}{}
	if bodyErr := checkBody(&err, &req, r.Body); bodyErr != nil {
		returnInternalServerError(w, bodyErr)
		return
	}
	name := checkNameNotEmpty(&err, req.Name)
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	// TODO check group limit reached
	data, err := h.dai.CreateValidatorDashboardGroup(dashboardId, name)
	if err != nil {
		handleError(w, err)
		return
	}

	response := types.ApiResponse{
		Data: data,
	}

	returnCreated(w, response)
}

func (h *HandlerService) InternalPutValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardPrimaryId(&err, vars["dashboard_id"])
	groupId := checkExistingGroupId(&err, vars["group_id"])
	req := struct {
		Name string `json:"name"`
	}{}
	if bodyErr := checkBody(&err, &req, r.Body); bodyErr != nil {
		returnInternalServerError(w, bodyErr)
		return
	}
	name := checkNameNotEmpty(&err, req.Name)
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	groupExists, err := h.dai.GetValidatorDashboardGroupExists(dashboardId, uint64(groupId))
	if err != nil {
		handleError(w, err)
		return
	}
	if !groupExists {
		returnNotFound(w, errors.New("group not found"))
		return
	}
	data, err := h.dai.UpdateValidatorDashboardGroup(dashboardId, uint64(groupId), name)
	if err != nil {
		handleError(w, err)
		return
	}

	response := types.ApiResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalDeleteValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardPrimaryId(&err, mux.Vars(r)["dashboard_id"])
	groupId := checkExistingGroupId(&err, vars["group_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	if groupId == types.DefaultGroupId {
		returnBadRequest(w, errors.New("cannot delete default group"))
		return
	}
	groupExists, err := h.dai.GetValidatorDashboardGroupExists(dashboardId, uint64(groupId))
	if err != nil {
		handleError(w, err)
		return
	}
	if !groupExists {
		returnNotFound(w, errors.New("group not found"))
		return
	}
	err = h.dai.RemoveValidatorDashboardGroup(dashboardId, uint64(groupId))
	if err != nil {
		handleError(w, err)
		return
	}

	returnNoContent(w)
}

func (h *HandlerService) InternalPostValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId := checkDashboardPrimaryId(&err, mux.Vars(r)["dashboard_id"])
	req := struct {
		Validators []string `json:"validators"`
		GroupId    string   `json:"group_id,omitempty"`
	}{}
	if bodyErr := checkBody(&err, &req, r.Body); bodyErr != nil {
		returnInternalServerError(w, bodyErr)
		return
	}
	indices, pubkeys := checkValidatorArray(&err, req.Validators)
	groupId := checkGroupId(&err, req.GroupId, allowEmpty)
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	// empty group id becomes default group
	if groupId == types.AllGroups {
		groupId = types.DefaultGroupId
	}
	groupExists, err := h.dai.GetValidatorDashboardGroupExists(dashboardId, uint64(groupId))
	if err != nil {
		handleError(w, err)
		return
	}
	if !groupExists {
		returnNotFound(w, errors.New("group not found"))
		return
	}
	validators, err := h.dai.GetValidatorsFromSlices(indices, pubkeys)
	if err != nil {
		handleError(w, err)
		return
	}
	// TODO check validator limit reached
	data, err := h.dai.AddValidatorDashboardValidators(dashboardId, groupId, validators)
	if err != nil {
		handleError(w, err)
		return
	}

	response := types.ApiResponse{
		Data: data,
	}

	returnCreated(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	q := r.URL.Query()
	groupId := checkGroupId(&err, q.Get("group_id"), allowEmpty)
	pagingParams := checkPagingParams(&err, q)
	sort := checkSort[enums.VDBManageValidatorsColumn](&err, q.Get("sort"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	data, paging, err := h.dai.GetValidatorDashboardValidators(*dashboardId, groupId, pagingParams.cursor, sort[0], pagingParams.search, pagingParams.limit)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardValidatorsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalDeleteValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId := checkDashboardPrimaryId(&err, mux.Vars(r)["dashboard_id"])
	var indices []uint64
	var publicKeys []string
	if validatorsParam := r.URL.Query().Get("validators"); validatorsParam != "" {
		indices, publicKeys = checkValidatorList(&err, validatorsParam)
		if err != nil {
			returnBadRequest(w, err)
			return
		}
	}
	validators, err := h.dai.GetValidatorsFromSlices(indices, publicKeys)
	if err != nil {
		handleError(w, err)
		return
	}
	err = h.dai.RemoveValidatorDashboardValidators(dashboardId, validators)
	if err != nil {
		handleError(w, err)
		return
	}

	returnNoContent(w)
}

func (h *HandlerService) InternalPostValidatorDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId := checkDashboardPrimaryId(&err, mux.Vars(r)["dashboard_id"])
	req := struct {
		Name          string `json:"name"`
		ShareSettings struct {
			GroupNames bool `json:"group_names"`
		} `json:"share_settings"`
	}{}
	if bodyErr := checkBody(&err, &req, r.Body); bodyErr != nil {
		returnInternalServerError(w, bodyErr)
		return
	}
	name := checkNameNotEmpty(&err, req.Name)
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	data, err := h.dai.CreateValidatorDashboardPublicId(dashboardId, name, req.ShareSettings.GroupNames)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}

	returnCreated(w, response)
}

func (h *HandlerService) InternalPutValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardPrimaryId(&err, mux.Vars(r)["dashboard_id"])
	req := struct {
		Name          string `json:"name"`
		ShareSettings struct {
			GroupNames bool `json:"group_names"`
		} `json:"share_settings"`
	}{}
	if bodyErr := checkBody(&err, &req, r.Body); bodyErr != nil {
		returnInternalServerError(w, bodyErr)
		return
	}
	name := checkNameNotEmpty(&err, req.Name)
	publicDashboardId := checkValidatorDashboardPublicId(&err, vars["public_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	dashboardInfo, err := h.dai.GetValidatorDashboardInfoByPublicId(publicDashboardId)
	if err != nil {
		handleError(w, err)
		return
	}
	if dashboardInfo.Id != dashboardId {
		returnNotFound(w, errors.New("public id not found"))
	}

	data, err := h.dai.UpdateValidatorDashboardPublicId(publicDashboardId, name, req.ShareSettings.GroupNames)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalDeleteValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardPrimaryId(&err, mux.Vars(r)["dashboard_id"])
	publicDashboardId := checkValidatorDashboardPublicId(&err, vars["public_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	dashboardInfo, err := h.dai.GetValidatorDashboardInfoByPublicId(publicDashboardId)
	if err != nil {
		handleError(w, err)
		return
	}
	if dashboardInfo.Id != dashboardId {
		returnNotFound(w, errors.New("public id not found"))
	}

	err = h.dai.RemoveValidatorDashboardPublicId(publicDashboardId)
	if err != nil {
		handleError(w, err)
		return
	}

	returnNoContent(w)
}

func (h *HandlerService) InternalGetValidatorDashboardSlotViz(w http.ResponseWriter, r *http.Request) {
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}

	data, err := h.dai.GetValidatorDashboardSlotViz(*dashboardId)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardSlotVizResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardSummary(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := checkPagingParams(&err, q)
	sort := checkSort[enums.VDBSummaryColumn](&err, q.Get("sort"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardSummary(*dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardSummaryResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupSummary(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(vars["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	groupId := checkGroupId(&err, vars["group_id"], forbidEmpty)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	data, err := h.dai.GetValidatorDashboardGroupSummary(*dashboardId, groupId)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardGroupSummaryResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardSummaryChart(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardSummaryChart(*dashboardId)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardSummaryChartResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardValidatorIndices(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	groupId := checkGroupId(&err, r.URL.Query().Get("group_id"), allowEmpty)
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	q := r.URL.Query()
	period := checkEnum[enums.TimePeriod](&err, q.Get("period"), "period")
	duty := checkEnum[enums.ValidatorDuty](&err, q.Get("duty"), "duty")
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	data, err := h.dai.GetValidatorDashboardValidatorIndices(*dashboardId, groupId, duty, period)
	if err != nil {
		handleError(w, err)
		return
	}

	response := types.InternalGetValidatorDashboardValidatorIndicesResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardRewards(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := checkPagingParams(&err, q)
	sort := checkSort[enums.VDBRewardsColumn](&err, q.Get("sort"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardRewards(*dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardRewardsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupRewards(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(vars["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	groupId := checkGroupId(&err, vars["group_id"], forbidEmpty)
	epoch := checkUint(&err, vars["epoch"], "epoch")
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	data, err := h.dai.GetValidatorDashboardGroupRewards(*dashboardId, groupId, epoch)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardGroupRewardsResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardRewardsChart(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(vars["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}

	data, err := h.dai.GetValidatorDashboardRewardsChart(*dashboardId)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardRewardsChartResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardDuties(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(vars["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	q := r.URL.Query()
	groupId := checkGroupId(&err, q.Get("group_id"), allowEmpty)
	epoch := checkUint(&err, vars["epoch"], "epoch")
	pagingParams := checkPagingParams(&err, q)
	sort := checkSort[enums.VDBDutiesColumn](&err, q.Get("sort"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardDuties(*dashboardId, epoch, groupId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardDutiesResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardBlocks(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := checkPagingParams(&err, q)
	sort := checkSort[enums.VDBBlocksColumn](&err, q.Get("sort"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardBlocks(*dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardBlocksResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardHeatmap(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}

	data, err := h.dai.GetValidatorDashboardHeatmap(*dashboardId)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardHeatmapResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupHeatmap(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	groupId := checkExistingGroupId(&err, vars["group_id"])
	epoch := checkUint(&err, vars["epoch"], "epoch")
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	data, err := h.dai.GetValidatorDashboardGroupHeatmap(*dashboardId, uint64(groupId), epoch)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardGroupHeatmapResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	pagingParams := checkPagingParams(&err, r.URL.Query())
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardElDeposits(*dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardExecutionLayerDepositsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	pagingParams := checkPagingParams(&err, r.URL.Query())
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardClDeposits(*dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleError(w, err)
		return
	}

	response := types.InternalGetValidatorDashboardConsensusLayerDepositsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardWithdrawals(w http.ResponseWriter, r *http.Request) {
	var err error
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleError(w, err)
		return
	}
	pagingParams := checkPagingParams(&err, q)
	sort := checkSort[enums.VDBWithdrawalsColumn](&err, q.Get("sort"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardWithdrawals(*dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardWithdrawalsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}
