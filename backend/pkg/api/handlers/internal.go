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
	returnCreated(w, nil)
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
	user, err := getUser(r)
	if err != nil {
		returnUnauthorized(w, err)
		return
	}
	data, err := h.dai.GetUserDashboards(user.Id)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.ApiDataResponse[types.UserDashboardsData]{
		Data: data,
	}
	returnOk(w, response)
}

// --------------------------------------
// Account Dashboards

func (h HandlerService) InternalPostAccountDashboards(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h HandlerService) InternalGetAccountDashboard(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h HandlerService) InternalDeleteAccountDashboard(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h HandlerService) InternalPostAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
}

func (h HandlerService) InternalDeleteAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w)
}

func (h HandlerService) InternalPostAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, nil)
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
	returnCreated(w, nil)
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

const errorMsgParsingId = "error parsing parameter 'dashboard_id'"

func (h HandlerService) InternalPostValidatorDashboards(w http.ResponseWriter, r *http.Request) {
	var err error
	user, err := getUser(r)
	if err != nil {
		returnUnauthorized(w, err)
		return
	}
	req := struct {
		Name    string `json:"name"`
		Network string `json:"network"`
	}{}
	if internalErr := checkBody(&err, &req, r.Body); internalErr != nil {
		returnInternalServerError(w, internalErr)
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
		returnInternalServerError(w, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}
	returnCreated(w, response)
}

func (h HandlerService) InternalGetValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	var data types.VDBOverviewData
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		// TODO auth check
		data, err = h.dai.GetValidatorDashboardOverview(dashboardId)
	case types.VDBIdPublic:
		data, err = h.dai.GetValidatorDashboardOverviewByPublicId(dashboardId)
	case types.VDBIdValidatorSet:
		data, err = h.dai.GetValidatorDashboardOverviewByValidators(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}

	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h HandlerService) InternalDeleteValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], false)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var dashboardInfo types.DashboardInfo
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	err = h.dai.RemoveValidatorDashboard(dashboardInfo.Id)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}

	returnNoContent(w)
}

func (h HandlerService) InternalPostValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	var err error
	req := struct {
		Name string `json:"name"`
	}{}
	if internalErr := checkBody(&err, &req, r.Body); internalErr != nil {
		returnInternalServerError(w, internalErr)
		return
	}
	vars := mux.Vars(r)
	name := checkNameNotEmpty(&err, req.Name)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], false)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var dashboardInfo types.DashboardInfo
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	// TODO check group limit reached
	data, err := h.dai.CreateValidatorDashboardGroup(dashboardInfo.Id, name)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}

	response := types.ApiResponse{
		Data: data,
	}

	returnCreated(w, response)
}

func (h HandlerService) InternalDeleteValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], false)
	groupId := checkGroupId(&err, vars["group_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var dashboardInfo types.DashboardInfo
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	err = h.dai.RemoveValidatorDashboardGroup(dashboardInfo.Id, groupId)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}

	returnNoContent(w)
}

func (h HandlerService) InternalPostValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var err error
	req := struct {
		Validators []string `json:"validators"`
		GroupId    string   `json:"group_id,omitempty"`
	}{}
	if internalErr := checkBody(&err, &req, r.Body); internalErr != nil {
		returnInternalServerError(w, internalErr)
		return
	}
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], false)
	groupId := checkGroupId(&err, req.GroupId)
	validators := checkValidatorArray(&err, req.Validators)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var dashboardInfo types.DashboardInfo
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	// TODO check validator limit reached
	data, err := h.dai.AddValidatorDashboardValidators(dashboardInfo.Id, groupId, validators)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}

	response := types.ApiResponse{
		Data: data,
	}

	returnCreated(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	q := r.URL.Query()
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	groupId := checkGroupId(&err, q.Get("group_id"))
	pagingParams := checkPagingParams(&err, q)
	sort := checkSort[enums.VDBManageValidatorsColumn](&err, q.Get("sort"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data []types.VDBManageValidatorsTableRow
	var paging types.Paging
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, paging, err = h.dai.GetValidatorDashboardValidators(dashboardId, groupId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		data, paging, err = h.dai.GetValidatorDashboardValidatorsByPublicId(dashboardId, groupId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdValidatorSet:
		data, paging, err = h.dai.GetValidatorDashboardValidatorsByValidators(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}

	response := types.InternalGetValidatorDashboardValidatorsResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalDeleteValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var err error
	// TODO check body for validators, ignore query param if body is present
	vars := mux.Vars(r)
	q := r.URL.Query()
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], false)
	validators := checkValidatorList(&err, q.Get("validators"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var dashboardInfo types.DashboardInfo
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	err = h.dai.RemoveValidatorDashboardValidators(dashboardInfo.Id, validators)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}

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
	if internalErr := checkBody(&err, &req, r.Body); internalErr != nil {
		returnInternalServerError(w, internalErr)
		return
	}
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], false)
	name := checkNameNotEmpty(&err, req.Name)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	//TODO check public id limit reached

	var dashboardInfo types.DashboardInfo
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	data, err := h.dai.CreateValidatorDashboardPublicId(dashboardInfo.Id, name, req.ShareSettings.GroupNames)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}

	returnCreated(w, response)
}

func (h HandlerService) InternalPutValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	var err error
	req := struct {
		Name          string `json:"name"`
		ShareSettings struct {
			GroupNames bool `json:"group_names"`
		} `json:"share_settings"`
	}{}
	if internalErr := checkBody(&err, &req, r.Body); internalErr != nil {
		returnInternalServerError(w, internalErr)
		return
	}
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], false)
	publicDashboardId := checkValidatorDashboardPublicId(&err, vars["public_dashboard_id"])
	name := checkNameNotEmpty(&err, req.Name)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var dashboardInfo types.DashboardInfo
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	data, err := h.dai.UpdateValidatorDashboardPublicId(dashboardInfo.Id, publicDashboardId, name, req.ShareSettings.GroupNames)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h HandlerService) InternalDeleteValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], false)
	publicDashboardId := checkValidatorDashboardPublicId(&err, vars["public_dashboard_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var dashboardInfo types.DashboardInfo
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.dai.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	err = h.dai.RemoveValidatorDashboardPublicId(dashboardInfo.Id, publicDashboardId)
	if err != nil {
		returnInternalServerError(w, err)
		return
	}

	returnNoContent(w)
}

func (h HandlerService) InternalGetValidatorDashboardSlotViz(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data []types.SlotVizEpoch
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, err = h.dai.GetValidatorDashboardSlotViz(dashboardId)
	case types.VDBIdPublic:
		data, err = h.dai.GetValidatorDashboardSlotVizByPublicId(dashboardId)
	case types.VDBIdValidatorSet:
		data, err = h.dai.GetValidatorDashboardSlotVizByValidators(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardSlotVizResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardSummary(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	q := r.URL.Query()
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	pagingParams := checkPagingParams(&err, q)
	sort := checkSort[enums.VDBSummaryColumn](&err, q.Get("sort"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data []types.VDBSummaryTableRow
	var paging types.Paging
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, paging, err = h.dai.GetValidatorDashboardSummary(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		data, paging, err = h.dai.GetValidatorDashboardSummaryByPublicId(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdValidatorSet:
		data, paging, err = h.dai.GetValidatorDashboardSummaryByValidators(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardSummaryResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardGroupSummary(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	groupId := checkGroupId(&err, vars["group_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data types.VDBGroupSummaryData
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, err = h.dai.GetValidatorDashboardGroupSummary(dashboardId, groupId)
	case types.VDBIdPublic:
		data, err = h.dai.GetValidatorDashboardGroupSummaryByPublicId(dashboardId, groupId)
	case types.VDBIdValidatorSet:
		data, err = h.dai.GetValidatorDashboardGroupSummaryByValidators(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardGroupSummaryResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardSummaryChart(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data types.ChartData[int]
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, err = h.dai.GetValidatorDashboardSummaryChart(dashboardId)
	case types.VDBIdPublic:
		data, err = h.dai.GetValidatorDashboardSummaryChartByPublicId(dashboardId)
	case types.VDBIdValidatorSet:
		data, err = h.dai.GetValidatorDashboardSummaryChartByValidators(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardSummaryChartResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardRewards(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	q := r.URL.Query()
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	pagingParams := checkPagingParams(&err, q)
	sort := checkSort[enums.VDBRewardsColumn](&err, q.Get("sort"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data []types.VDBRewardsTableRow
	var paging types.Paging
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, paging, err = h.dai.GetValidatorDashboardRewards(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		data, paging, err = h.dai.GetValidatorDashboardRewardsByPublicId(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdValidatorSet:
		data, paging, err = h.dai.GetValidatorDashboardRewardsByValidators(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardRewardsResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardGroupRewards(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	groupId := checkGroupId(&err, vars["group_id"])
	epoch := checkUint(&err, vars["epoch"], "epoch")
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data types.VDBGroupRewardsData
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, err = h.dai.GetValidatorDashboardGroupRewards(dashboardId, groupId, epoch)
	case types.VDBIdPublic:
		data, err = h.dai.GetValidatorDashboardGroupRewardsByPublicId(dashboardId, groupId, epoch)
	case types.VDBIdValidatorSet:
		data, err = h.dai.GetValidatorDashboardGroupRewardsByValidators(dashboardId, epoch)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardGroupRewardsResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardRewardsChart(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data types.ChartData[int]
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, err = h.dai.GetValidatorDashboardRewardsChart(dashboardId)
	case types.VDBIdPublic:
		data, err = h.dai.GetValidatorDashboardRewardsChartByPublicId(dashboardId)
	case types.VDBIdValidatorSet:
		data, err = h.dai.GetValidatorDashboardRewardsChartByValidators(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardRewardsChartResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardDuties(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	q := r.URL.Query()
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	epoch := checkUint(&err, vars["epoch"], "epoch")
	pagingParams := checkPagingParams(&err, q)
	sort := checkSort[enums.VDBDutiesColumn](&err, q.Get("sort"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data []types.VDBEpochDutiesTableRow
	var paging types.Paging
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, paging, err = h.dai.GetValidatorDashboardDuties(dashboardId, epoch, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		data, paging, err = h.dai.GetValidatorDashboardDutiesByPublicId(dashboardId, epoch, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdValidatorSet:
		data, paging, err = h.dai.GetValidatorDashboardDutiesByValidators(dashboardId, epoch, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardDutiesResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardBlocks(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	q := r.URL.Query()
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	pagingParams := checkPagingParams(&err, q)
	sort := checkSort[enums.VDBBlocksColumn](&err, q.Get("sort"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data []types.VDBBlocksTableRow
	var paging types.Paging
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, paging, err = h.dai.GetValidatorDashboardBlocks(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		data, paging, err = h.dai.GetValidatorDashboardBlocksByPublicId(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdValidatorSet:
		data, paging, err = h.dai.GetValidatorDashboardBlocksByValidators(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardBlocksResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardHeatmap(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data types.VDBHeatmap
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, err = h.dai.GetValidatorDashboardHeatmap(dashboardId)
	case types.VDBIdPublic:
		data, err = h.dai.GetValidatorDashboardHeatmapByPublicId(dashboardId)
	case types.VDBIdValidatorSet:
		data, err = h.dai.GetValidatorDashboardHeatmapByValidators(dashboardId)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardHeatmapResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardGroupHeatmap(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	groupId := checkGroupId(&err, vars["group_id"])
	epoch := checkUint(&err, vars["epoch"], "epoch")
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data types.VDBHeatmapTooltipData
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, err = h.dai.GetValidatorDashboardGroupHeatmap(dashboardId, groupId, epoch)
	case types.VDBIdPublic:
		data, err = h.dai.GetValidatorDashboardGroupHeatmapByPublicId(dashboardId, groupId, epoch)
	case types.VDBIdValidatorSet:
		data, err = h.dai.GetValidatorDashboardGroupHeatmapByValidators(dashboardId, epoch)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardGroupHeatmapResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	pagingParams := checkPagingParams(&err, r.URL.Query())
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data []types.VDBExecutionDepositsTableRow
	var paging types.Paging
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, paging, err = h.dai.GetValidatorDashboardElDeposits(dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		data, paging, err = h.dai.GetValidatorDashboardElDepositsByPublicId(dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	case types.VDBIdValidatorSet:
		data, paging, err = h.dai.GetValidatorDashboardElDepositsByValidators(dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardExecutionLayerDepositsResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	pagingParams := checkPagingParams(&err, r.URL.Query())
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data []types.VDBConsensusDepositsTableRow
	var paging types.Paging
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, paging, err = h.dai.GetValidatorDashboardClDeposits(dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		data, paging, err = h.dai.GetValidatorDashboardClDepositsByPublicId(dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	case types.VDBIdValidatorSet:
		data, paging, err = h.dai.GetValidatorDashboardClDepositsByValidators(dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardConsensusLayerDepositsResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h HandlerService) InternalGetValidatorDashboardWithdrawals(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	q := r.URL.Query()
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	pagingParams := checkPagingParams(&err, q)
	sort := checkSort[enums.VDBWithdrawalsColumn](&err, q.Get("sort"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var data []types.VDBWithdrawalsTableRow
	var paging types.Paging
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		data, paging, err = h.dai.GetValidatorDashboardWithdrawals(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		data, paging, err = h.dai.GetValidatorDashboardWithdrawalsByPublicId(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdValidatorSet:
		data, paging, err = h.dai.GetValidatorDashboardWithdrawalsByValidators(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, errors.New(errorMsgParsingId))
		return
	}
	if err != nil {
		returnInternalServerError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardWithdrawalsResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}
