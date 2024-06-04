package handlers

import (
	"errors"
	"net/http"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	types "github.com/gobitfly/beaconchain/pkg/api/types"

	"github.com/gorilla/mux"
)

// --------------------------------------
// Premium Plans

func (h *HandlerService) InternalGetProductSummary(w http.ResponseWriter, r *http.Request) {
	data, err := h.dai.GetProductSummary()
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetProductSummaryResponse{
		Data: *data,
	}
	returnOk(w, response)
}

// --------------------------------------
// Latest State

func (h *HandlerService) InternalGetLatestState(w http.ResponseWriter, r *http.Request) {
	latestSlot, err := h.dai.GetLatestSlot()
	if err != nil {
		handleErr(w, err)
		return
	}

	exchangeRates, err := h.dai.GetLatestExchangeRates()
	if err != nil {
		handleErr(w, err)
		return
	}
	data := types.LatestStateData{
		LatestSlot:    latestSlot,
		ExchangeRates: exchangeRates,
	}

	response := types.InternalGetLatestStateResponse{
		Data: data,
	}
	returnOk(w, response)
}

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
// User

func (h *HandlerService) InternalGetUserInfo(w http.ResponseWriter, r *http.Request) {
	// TODO patrick
	user, err := h.getUser(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	userInfo, err := h.dai.GetUserInfo(user.Id)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserInfoResponse{
		Data: *userInfo,
	}
	returnOk(w, response)
}

// --------------------------------------
// Dashboards

func (h *HandlerService) InternalGetUserDashboards(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUser(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	data, err := h.dai.GetUserDashboards(user.Id)
	if err != nil {
		handleErr(w, err)
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

func (h *HandlerService) InternalPostValidatorDashboards(w http.ResponseWriter, r *http.Request) {
	var v validationError
	user, err := h.getUser(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	req := struct {
		Name    string      `json:"name"`
		Network intOrString `json:"network"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	name := v.checkNameNotEmpty(req.Name)
	chainId := v.checkNetwork(req.Network)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.CreateValidatorDashboard(user.Id, name, chainId)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.ApiDataResponse[types.VDBPostReturnData]{
		Data: *data,
	}
	returnCreated(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	dashboardIdParam := mux.Vars(r)["dashboard_id"]
	dashboardId, err := h.handleDashboardId(dashboardIdParam)
	if err != nil {
		handleErr(w, err)
		return
	}
	// set variables depending on public id being used
	var name string
	if reValidatorDashboardPublicId.MatchString(dashboardIdParam) {
		var publicIdInfo *types.VDBPublicId
		publicIdInfo, err = h.dai.GetValidatorDashboardPublicId(types.VDBIdPublic(dashboardIdParam))
		name = publicIdInfo.Name
	} else {
		name, err = h.dai.GetValidatorDashboardName(dashboardId.Id)
	}
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetValidatorDashboardOverview(*dashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	data.Name = name

	response := types.InternalGetValidatorDashboardResponse{
		Data: *data,
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalDeleteValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	err := h.dai.RemoveValidatorDashboard(dashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	returnNoContent(w)
}

func (h *HandlerService) InternalPutValidatorDashboardName(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	req := struct {
		Name string `json:"name"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	name := v.checkNameNotEmpty(req.Name)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	data, err := h.dai.UpdateValidatorDashboardName(dashboardId, name)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.ApiDataResponse[types.VDBPostReturnData]{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalPostValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	req := struct {
		Name string `json:"name"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	name := v.checkNameNotEmpty(req.Name)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	// TODO check group limit reached
	data, err := h.dai.CreateValidatorDashboardGroup(dashboardId, name)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.ApiResponse{
		Data: data,
	}

	returnCreated(w, response)
}

func (h *HandlerService) InternalPutValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId := v.checkPrimaryDashboardId(vars["dashboard_id"])
	groupId := v.checkExistingGroupId(vars["group_id"])
	req := struct {
		Name string `json:"name"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	name := v.checkNameNotEmpty(req.Name)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	groupExists, err := h.dai.GetValidatorDashboardGroupExists(dashboardId, groupId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if !groupExists {
		returnNotFound(w, errors.New("group not found"))
		return
	}
	data, err := h.dai.UpdateValidatorDashboardGroup(dashboardId, groupId, name)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.ApiResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalDeleteValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	groupId := v.checkExistingGroupId(vars["group_id"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	if groupId == types.DefaultGroupId {
		returnBadRequest(w, errors.New("cannot delete default group"))
		return
	}
	groupExists, err := h.dai.GetValidatorDashboardGroupExists(dashboardId, groupId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if !groupExists {
		returnNotFound(w, errors.New("group not found"))
		return
	}
	err = h.dai.RemoveValidatorDashboardGroup(dashboardId, groupId)
	if err != nil {
		handleErr(w, err)
		return
	}

	returnNoContent(w)
}

func (h *HandlerService) InternalPostValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	req := struct {
		Validators []string `json:"validators"`
		GroupId    string   `json:"group_id,omitempty"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	indices, pubkeys := v.checkValidatorArray(req.Validators, forbidEmpty)
	groupId := v.checkGroupId(req.GroupId, allowEmpty)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	// empty group id becomes default group
	if groupId == types.AllGroups {
		groupId = types.DefaultGroupId
	}
	groupExists, err := h.dai.GetValidatorDashboardGroupExists(dashboardId, uint64(groupId))
	if err != nil {
		handleErr(w, err)
		return
	}
	if !groupExists {
		returnNotFound(w, errors.New("group not found"))
		return
	}
	validators, err := h.dai.GetValidatorsFromSlices(indices, pubkeys)
	if err != nil {
		handleErr(w, err)
		return
	}
	// TODO check validator limit reached
	data, err := h.dai.AddValidatorDashboardValidators(dashboardId, groupId, validators)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.ApiResponse{
		Data: data,
	}

	returnCreated(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	groupId := v.checkGroupId(q.Get("group_id"), allowEmpty)
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBManageValidatorsColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	data, paging, err := h.dai.GetValidatorDashboardValidators(*dashboardId, groupId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardValidatorsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalDeleteValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	var indices []uint64
	var publicKeys []string
	if validatorsParam := r.URL.Query().Get("validators"); validatorsParam != "" {
		indices, publicKeys = v.checkValidatorList(validatorsParam, allowEmpty)
		if v.hasErrors() {
			handleErr(w, v)
			return
		}
	}
	validators, err := h.dai.GetValidatorsFromSlices(indices, publicKeys)
	if err != nil {
		handleErr(w, err)
		return
	}
	err = h.dai.RemoveValidatorDashboardValidators(dashboardId, validators)
	if err != nil {
		handleErr(w, err)
		return
	}

	returnNoContent(w)
}

func (h *HandlerService) InternalPostValidatorDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	req := struct {
		Name          string `json:"name,omitempty"`
		ShareSettings struct {
			ShareGroups bool `json:"share_groups"`
		} `json:"share_settings"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	name := v.checkName(req.Name, 0)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	data, err := h.dai.CreateValidatorDashboardPublicId(dashboardId, name, req.ShareSettings.ShareGroups)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}

	returnCreated(w, response)
}

func (h *HandlerService) InternalPutValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	req := struct {
		Name          string `json:"name"`
		ShareSettings struct {
			ShareGroups bool `json:"share_groups"`
		} `json:"share_settings"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	name := v.checkNameNotEmpty(req.Name)
	publicDashboardId := v.checkValidatorDashboardPublicId(vars["public_id"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	dashboardInfo, err := h.dai.GetValidatorDashboardInfoByPublicId(publicDashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if dashboardInfo.Id != dashboardId {
		handleErr(w, newNotFoundErr("public id %v not found", publicDashboardId))
	}

	data, err := h.dai.UpdateValidatorDashboardPublicId(publicDashboardId, name, req.ShareSettings.ShareGroups)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalDeleteValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	publicDashboardId := v.checkValidatorDashboardPublicId(vars["public_id"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	dashboardInfo, err := h.dai.GetValidatorDashboardInfoByPublicId(publicDashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if dashboardInfo.Id != dashboardId {
		handleErr(w, newNotFoundErr("public id %v not found", publicDashboardId))
	}

	err = h.dai.RemoveValidatorDashboardPublicId(publicDashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}

	returnNoContent(w)
}

func (h *HandlerService) InternalGetValidatorDashboardSlotViz(w http.ResponseWriter, r *http.Request) {
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetValidatorDashboardSlotViz(*dashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardSlotVizResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardSummary(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBSummaryColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardSummary(*dashboardId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardSummaryResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupSummary(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(vars["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	groupId := v.checkGroupId(vars["group_id"], forbidEmpty)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardGroupSummary(*dashboardId, groupId)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardGroupSummaryResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardSummaryChart(w http.ResponseWriter, r *http.Request) {
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardSummaryChart(*dashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardSummaryChartResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardValidatorIndices(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	groupId := v.checkGroupId(r.URL.Query().Get("group_id"), allowEmpty)
	q := r.URL.Query()
	duty := checkEnum[enums.ValidatorDuty](&v, q.Get("duty"), "duty")
	period := checkEnum[enums.TimePeriod](&v, q.Get("period"), "period")
	// allowed periods are: all_time, last_24h, last_7d, last_30d
	allowedPeriods := []enums.Enum{enums.TimePeriods.AllTime, enums.TimePeriods.Last24h, enums.TimePeriods.Last7d, enums.TimePeriods.Last30d}
	v.checkEnumIsAllowed(period, allowedPeriods, "period")
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardValidatorIndices(*dashboardId, groupId, duty, period)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetValidatorDashboardValidatorIndicesResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardRewards(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBRewardsColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardRewards(*dashboardId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardRewardsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupRewards(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(vars["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	groupId := v.checkGroupId(vars["group_id"], forbidEmpty)
	epoch := v.checkUint(vars["epoch"], "epoch")
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardGroupRewards(*dashboardId, groupId, epoch)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardGroupRewardsResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardRewardsChart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(vars["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetValidatorDashboardRewardsChart(*dashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardRewardsChartResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardDuties(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(vars["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	groupId := v.checkGroupId(q.Get("group_id"), allowEmpty)
	epoch := v.checkUint(vars["epoch"], "epoch")
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBDutiesColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardDuties(*dashboardId, epoch, groupId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardDutiesResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardBlocks(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBBlocksColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardBlocks(*dashboardId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardBlocksResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardEpochHeatmap(w http.ResponseWriter, r *http.Request) {
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}

	// implicit time period is last hour
	data, err := h.dai.GetValidatorDashboardEpochHeatmap(*dashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardHeatmapResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardDailyHeatmap(w http.ResponseWriter, r *http.Request) {
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}

	var v validationError
	period := checkEnum[enums.TimePeriod](&v, r.URL.Query().Get("period"), "period")
	// allowed periods are: last_7d, last_30d, last_365d
	allowedPeriods := []enums.Enum{enums.TimePeriods.Last7d, enums.TimePeriods.Last30d, enums.TimePeriods.Last365d}
	v.checkEnumIsAllowed(period, allowedPeriods, "period")
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	data, err := h.dai.GetValidatorDashboardDailyHeatmap(*dashboardId, period)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardHeatmapResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupEpochHeatmap(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	groupId := v.checkExistingGroupId(vars["group_id"])
	epoch := v.checkUint(vars["epoch"], "epoch")
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardGroupEpochHeatmap(*dashboardId, groupId, epoch)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardGroupHeatmapResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupDailyHeatmap(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	groupId := v.checkExistingGroupId(vars["group_id"])
	date := v.checkDate(vars["date"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardGroupDailyHeatmap(*dashboardId, groupId, date)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardGroupHeatmapResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	pagingParams := v.checkPagingParams(r.URL.Query())
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardElDeposits(*dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardExecutionLayerDepositsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	pagingParams := v.checkPagingParams(r.URL.Query())
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardClDeposits(*dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetValidatorDashboardConsensusLayerDepositsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardTotalConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardTotalClDeposits(*dashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetValidatorDashboardTotalConsensusDepositsResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardTotalExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardTotalElDeposits(*dashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetValidatorDashboardTotalExecutionDepositsResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardWithdrawals(w http.ResponseWriter, r *http.Request) {
	var v validationError
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBWithdrawalsColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardWithdrawals(*dashboardId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardWithdrawalsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardTotalWithdrawals(w http.ResponseWriter, r *http.Request) {
	var v validationError
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	pagingParams := v.checkPagingParams(q)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardTotalWithdrawals(*dashboardId, pagingParams.search)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetValidatorDashboardTotalWithdrawalsResponse{
		Data: *data,
	}
	returnOk(w, response)
}
