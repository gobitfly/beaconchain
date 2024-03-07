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

func (h *HandlerService) InternalPostOauthAuthorize(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) InternalPostOauthToken(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

func (h *HandlerService) InternalPostApiKeys(w http.ResponseWriter, r *http.Request) {
	returnOk(w, nil)
}

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
	user, err := getUser(r)
	if err != nil {
		returnUnauthorized(w, err)
		return
	}
	data, err := h.da.GetUserDashboards(user.Id)
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.ApiDataResponse[types.UserDashboardsData]{
		Data: data,
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

var ErrParseDashboardId = errors.New("failed to parse dashboard id")

func (h *HandlerService) InternalPostValidatorDashboards(w http.ResponseWriter, r *http.Request) {
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

	data, err := h.da.CreateValidatorDashboard(user.Id, name, network)
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
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], true)
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	fetcher, err := h.getDashboardFetcher(r, dashboardId)
	if err != nil {
		handleError(w, err)
		return
	}
	data, err := fetcher.getValidatorDashboardOverview()
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalDeleteValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], false)
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	fetcher, err := h.getDashboardFetcher(r, dashboardId)
	if err != nil {
		handleError(w, err)
		return
	}
	err = fetcher.removeValidatorDashboard(r)
	if err != nil {
		handleError(w, err)
		return
	}

	returnNoContent(w)
}

func (h *HandlerService) InternalPostValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
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
		dashboardInfo, err = h.da.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	// TODO check if user is authorized for this dashboard
	// TODO check group limit reached
	data, err := h.da.CreateValidatorDashboardGroup(dashboardInfo.Id, name)
	if err != nil {
		handleError(w, err)
		return
	}

	response := types.ApiResponse{
		Data: data,
	}

	returnCreated(w, response)
}

func (h *HandlerService) InternalDeleteValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
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
		dashboardInfo, err = h.da.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	// TODO check if user is authorized for this dashboard
	err = h.da.RemoveValidatorDashboardGroup(dashboardInfo.Id, groupId)
	if err != nil {
		handleError(w, err)
		return
	}

	returnNoContent(w)
}

func (h *HandlerService) InternalPostValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
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
	validatorArr := checkValidatorArray(&err, req.Validators)
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	validators, err := h.da.GetValidatorsFromStrings(validatorArr)
	if err != nil {
		handleError(w, err)
		return
	}

	var dashboardInfo types.DashboardInfo
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		dashboardInfo, err = h.da.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	// TODO check if user is authorized for this dashboard
	// TODO check validator limit reached
	data, err := h.da.AddValidatorDashboardValidators(dashboardInfo.Id, groupId, validators)
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
		// TODO check if user is authorized for this dashboard
		data, paging, err = h.da.GetValidatorDashboardValidators(dashboardId, groupId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardValidators(dashboardInfo.Id, groupId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardValidatorsByValidators(validators, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}

	response := types.InternalGetValidatorDashboardValidatorsResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalDeleteValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var err error
	// TODO check body for validators, ignore query param if body is present
	vars := mux.Vars(r)
	q := r.URL.Query()
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], false)
	validatorArr := checkValidatorList(&err, q.Get("validators"))
	if err != nil {
		returnBadRequest(w, err)
		return
	}
	validators, err := h.da.GetValidatorsFromStrings(validatorArr)
	if err != nil {
		handleError(w, err)
		return
	}

	var dashboardInfo types.DashboardInfo
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		// TODO check if user is authorized for this dashboard
		dashboardInfo, err = h.da.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	err = h.da.RemoveValidatorDashboardValidators(dashboardInfo.Id, validators)
	if err != nil {
		handleError(w, err)
		return
	}

	returnNoContent(w)
}

func (h *HandlerService) InternalPostValidatorDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
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
		dashboardInfo, err = h.da.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	data, err := h.da.CreateValidatorDashboardPublicId(dashboardInfo.Id, name, req.ShareSettings.GroupNames)
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
		dashboardInfo, err = h.da.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	data, err := h.da.UpdateValidatorDashboardPublicId(dashboardInfo.Id, publicDashboardId, name, req.ShareSettings.GroupNames)
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
	dashboardId := checkDashboardId(&err, vars["dashboard_id"], false)
	publicDashboardId := checkValidatorDashboardPublicId(&err, vars["public_dashboard_id"])
	if err != nil {
		returnBadRequest(w, err)
		return
	}

	var dashboardInfo types.DashboardInfo
	switch dashboardId := dashboardId.(type) {
	case types.VDBIdPrimary:
		dashboardInfo, err = h.da.GetValidatorDashboardInfo(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, err = h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	err = h.da.RemoveValidatorDashboardPublicId(dashboardInfo.Id, publicDashboardId)
	if err != nil {
		handleError(w, err)
		return
	}

	returnNoContent(w)
}

func (h *HandlerService) InternalGetValidatorDashboardSlotViz(w http.ResponseWriter, r *http.Request) {
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
		data, err = h.da.GetValidatorDashboardSlotViz(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardSlotViz(dashboardInfo.Id)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardSlotVizByValidators(validators)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
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
		data, paging, err = h.da.GetValidatorDashboardSummary(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardSummary(dashboardInfo.Id, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardSummaryByValidators(validators, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardSummaryResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupSummary(w http.ResponseWriter, r *http.Request) {
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
		data, err = h.da.GetValidatorDashboardGroupSummary(dashboardId, groupId)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardGroupSummary(dashboardInfo.Id, groupId)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardGroupSummaryByValidators(validators)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardGroupSummaryResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardSummaryChart(w http.ResponseWriter, r *http.Request) {
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
		data, err = h.da.GetValidatorDashboardSummaryChart(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardSummaryChart(dashboardInfo.Id)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardSummaryChartByValidators(validators)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardSummaryChartResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardRewards(w http.ResponseWriter, r *http.Request) {
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
		data, paging, err = h.da.GetValidatorDashboardRewards(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardRewards(dashboardInfo.Id, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardRewardsByValidators(validators, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardRewardsResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupRewards(w http.ResponseWriter, r *http.Request) {
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
		data, err = h.da.GetValidatorDashboardGroupRewards(dashboardId, groupId, epoch)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardGroupRewards(dashboardInfo.Id, groupId, epoch)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardGroupRewardsByValidators(validators, epoch)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardGroupRewardsResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardRewardsChart(w http.ResponseWriter, r *http.Request) {
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
		data, err = h.da.GetValidatorDashboardRewardsChart(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardRewardsChart(dashboardInfo.Id)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardRewardsChartByValidators(validators)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardRewardsChartResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardDuties(w http.ResponseWriter, r *http.Request) {
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
		data, paging, err = h.da.GetValidatorDashboardDuties(dashboardId, epoch, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardDuties(dashboardInfo.Id, epoch, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardDutiesByValidators(validators, epoch, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardDutiesResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardBlocks(w http.ResponseWriter, r *http.Request) {
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
		data, paging, err = h.da.GetValidatorDashboardBlocks(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardBlocks(dashboardInfo.Id, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardBlocksByValidators(validators, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardBlocksResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardHeatmap(w http.ResponseWriter, r *http.Request) {
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
		data, err = h.da.GetValidatorDashboardHeatmap(dashboardId)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardHeatmap(dashboardInfo.Id)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardHeatmapByValidators(validators)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardHeatmapResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupHeatmap(w http.ResponseWriter, r *http.Request) {
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
		data, err = h.da.GetValidatorDashboardGroupHeatmap(dashboardId, groupId, epoch)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardGroupHeatmap(dashboardInfo.Id, groupId, epoch)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, err = h.da.GetValidatorDashboardGroupHeatmapByValidators(validators, epoch)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardGroupHeatmapResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
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
		data, paging, err = h.da.GetValidatorDashboardElDeposits(dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardElDeposits(dashboardInfo.Id, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardElDepositsByValidators(validators, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardExecutionLayerDepositsResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
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
		data, paging, err = h.da.GetValidatorDashboardClDeposits(dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardClDeposits(dashboardInfo.Id, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardClDepositsByValidators(validators, pagingParams.cursor, pagingParams.search, pagingParams.limit)
	}
	if err != nil {
		handleError(w, err)
		return
	}

	response := types.InternalGetValidatorDashboardConsensusLayerDepositsResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardWithdrawals(w http.ResponseWriter, r *http.Request) {
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
		data, paging, err = h.da.GetValidatorDashboardWithdrawals(dashboardId, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case types.VDBIdPublic:
		dashboardInfo, infoErr := h.da.GetValidatorDashboardInfoByPublicId(dashboardId)
		if infoErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardWithdrawals(dashboardInfo.Id, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	case []string:
		validators, convertErr := h.da.GetValidatorsFromStrings(dashboardId)
		if convertErr != nil {
			handleError(w, err)
			return
		}
		data, paging, err = h.da.GetValidatorDashboardWithdrawalsByValidators(validators, pagingParams.cursor, sort, pagingParams.search, pagingParams.limit)
	default:
		returnInternalServerError(w, ErrParseDashboardId)
		return
	}
	if err != nil {
		handleError(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardWithdrawalsResponse{
		Data:   data,
		Paging: paging,
	}
	returnOk(w, response)
}
