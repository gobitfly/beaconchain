package handlers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"reflect"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	types "github.com/gobitfly/beaconchain/pkg/api/types"

	"github.com/gorilla/mux"
)

// --------------------------------------
// Premium Plans

func (h *HandlerService) InternalGetProductSummary(w http.ResponseWriter, r *http.Request) {
	data, err := h.dai.GetProductSummary(r.Context())
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

	finalizedEpoch, err := h.dai.GetLatestFinalizedEpoch()
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
		LatestSlot:     latestSlot,
		FinalizedEpoch: finalizedEpoch,
		ExchangeRates:  exchangeRates,
	}

	response := types.InternalGetLatestStateResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetRocketPool(w http.ResponseWriter, r *http.Request) {
	data, err := h.dai.GetRocketPoolOverview(r.Context())
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetRocketPoolResponse{
		Data: *data,
	}
	returnOk(w, response)
}

// All handler function names must include the HTTP method and the path they handle
// Internal handlers may only be authenticated by an OAuth token

// --------------------------------------
// Ad Configurations

func (h *HandlerService) InternalPostAdConfigurations(w http.ResponseWriter, r *http.Request) {
	var v validationError
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	if user.UserGroup != "ADMIN" {
		returnForbidden(w, errors.New("user is not an admin"))
		return
	}

	var req types.AdConfigurationData
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	key := v.checkKeyNotEmpty(req.Key)
	if len(req.JQuerySelector) == 0 {
		v.add("jquery_selector", "must not be empty")
	}
	insertMode := checkEnum[enums.AdInsertMode](&v, req.InsertMode, "insert_mode")
	if req.RefreshInterval == 0 {
		v.add("refresh_interval", "must be greater than 0")
	}
	if (req.BannerId == 0) == (req.HtmlContent == "") {
		returnBadRequest(w, errors.New("provide either banner_id or html_content"))
		return
	}
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	err = h.dai.CreateAdConfiguration(r.Context(), key, req.JQuerySelector, insertMode, req.RefreshInterval, req.ForAllUsers, req.BannerId, req.HtmlContent, req.Enabled)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.ApiDataResponse[types.AdConfigurationData]{
		Data: req,
	}
	returnCreated(w, response)
}

func (h *HandlerService) InternalGetAdConfigurations(w http.ResponseWriter, r *http.Request) {
	var v validationError
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	if user.UserGroup != "ADMIN" {
		returnForbidden(w, errors.New("user is not an admin"))
		return
	}

	keys := v.checkAdConfigurationKeys(r.URL.Query().Get("keys"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.GetAdConfigurations(r.Context(), keys)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.ApiDataResponse[[]types.AdConfigurationData]{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalPutAdConfiguration(w http.ResponseWriter, r *http.Request) {
	var v validationError
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	if user.UserGroup != "ADMIN" {
		returnForbidden(w, errors.New("user is not an admin"))
		return
	}

	key := v.checkKeyNotEmpty(mux.Vars(r)["key"])
	var req types.AdConfigurationUpdateData
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	if len(req.JQuerySelector) == 0 {
		v.add("jquery_selector", "must not be empty")
	}
	insertMode := checkEnum[enums.AdInsertMode](&v, req.InsertMode, "insert_mode")
	if req.RefreshInterval == 0 {
		v.add("refresh_interval", "must be greater than 0")
	}
	if (req.BannerId == 0) == (req.HtmlContent == "") {
		returnConflict(w, errors.New("provide either banner_id or html_content"))
		return
	}
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	err = h.dai.UpdateAdConfiguration(r.Context(), key, req.JQuerySelector, insertMode, req.RefreshInterval, req.ForAllUsers, req.BannerId, req.HtmlContent, req.Enabled)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.ApiDataResponse[types.AdConfigurationData]{
		Data: types.AdConfigurationData{Key: key, AdConfigurationUpdateData: &req},
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalDeleteAdConfiguration(w http.ResponseWriter, r *http.Request) {
	var v validationError
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	if user.UserGroup != "ADMIN" {
		returnForbidden(w, errors.New("user is not an admin"))
		return
	}

	key := v.checkKeyNotEmpty(mux.Vars(r)["key"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	err = h.dai.RemoveAdConfiguration(r.Context(), key)
	if err != nil {
		handleErr(w, err)
		return
	}

	returnNoContent(w)
}

// --------------------------------------
// User

func (h *HandlerService) InternalGetUserInfo(w http.ResponseWriter, r *http.Request) {
	// TODO patrick
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	userInfo, err := h.dai.GetUserInfo(r.Context(), user.Id)
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
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	data, err := h.dai.GetUserDashboards(r.Context(), userId)
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
	userId, err := GetUserIdByContext(r)
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

	userInfo, err := h.dai.GetUserInfo(r.Context(), userId)
	if err != nil {
		handleErr(w, err)
		return
	}
	dashboardCount, err := h.dai.GetUserValidatorDashboardCount(r.Context(), userId, true)
	if err != nil {
		handleErr(w, err)
		return
	}
	if dashboardCount >= userInfo.PremiumPerks.ValidatorDashboards && !isUserAdmin(userInfo) {
		returnConflict(w, errors.New("maximum number of validator dashboards reached"))
		return
	}

	data, err := h.dai.CreateValidatorDashboard(r.Context(), userId, name, chainId)
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
	var v validationError
	dashboardIdParam := mux.Vars(r)["dashboard_id"]
	dashboardId, err := h.handleDashboardId(r.Context(), dashboardIdParam)
	if err != nil {
		handleErr(w, err)
		return
	}

	q := r.URL.Query()
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	// set name depending on dashboard id
	var name string
	if reInteger.MatchString(dashboardIdParam) {
		name, err = h.dai.GetValidatorDashboardName(r.Context(), dashboardId.Id)
	} else if reValidatorDashboardPublicId.MatchString(dashboardIdParam) {
		var publicIdInfo *types.VDBPublicId
		publicIdInfo, err = h.dai.GetValidatorDashboardPublicId(r.Context(), types.VDBIdPublic(dashboardIdParam))
		name = publicIdInfo.Name
	}
	if err != nil {
		handleErr(w, err)
		return
	}

	// add premium chart perk info for shared dashboards
	premiumPerks, err := h.getDashboardPremiumPerks(r.Context(), *dashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardOverview(r.Context(), *dashboardId, protocolModes)
	if err != nil {
		handleErr(w, err)
		return
	}
	data.ChartHistorySeconds = premiumPerks.ChartHistorySeconds
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
	err := h.dai.RemoveValidatorDashboard(r.Context(), dashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	returnNoContent(w)
}

func (h *HandlerService) InternalPutValidatorDashboardArchiving(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	req := struct {
		IsArchived bool `json:"is_archived"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	// check conditions for changing archival status
	dashboardInfo, err := h.dai.GetValidatorDashboard(r.Context(), types.VDBId{Id: dashboardId})
	if err != nil {
		handleErr(w, err)
		return
	}
	if dashboardInfo.IsArchived == req.IsArchived {
		// nothing to do
		returnOk(w, types.ApiDataResponse[types.VDBPostArchivingReturnData]{
			Data: types.VDBPostArchivingReturnData{Id: uint64(dashboardId), IsArchived: req.IsArchived},
		})
		return
	}

	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	dashboardCount, err := h.dai.GetUserValidatorDashboardCount(r.Context(), userId, !req.IsArchived)
	if err != nil {
		handleErr(w, err)
		return
	}

	userInfo, err := h.dai.GetUserInfo(r.Context(), userId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if !isUserAdmin(userInfo) {
		if req.IsArchived {
			if dashboardCount >= MaxArchivedDashboardsCount {
				returnConflict(w, errors.New("maximum number of archived validator dashboards reached"))
				return
			}
		} else {
			if dashboardCount >= userInfo.PremiumPerks.ValidatorDashboards {
				returnConflict(w, errors.New("maximum number of active validator dashboards reached"))
				return
			}
			if dashboardInfo.GroupCount >= userInfo.PremiumPerks.ValidatorGroupsPerDashboard {
				returnConflict(w, errors.New("maximum number of groups in dashboards reached"))
				return
			}
			if dashboardInfo.ValidatorCount >= userInfo.PremiumPerks.ValidatorsPerDashboard {
				returnConflict(w, errors.New("maximum number of validators in dashboards reached"))
				return
			}
		}
	}

	var archivedReason *enums.VDBArchivedReason
	if req.IsArchived {
		archivedReason = &enums.VDBArchivedReasons.User
	}

	data, err := h.dai.UpdateValidatorDashboardArchiving(r.Context(), dashboardId, archivedReason)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.ApiDataResponse[types.VDBPostArchivingReturnData]{
		Data: *data,
	}
	returnOk(w, response)
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
	data, err := h.dai.UpdateValidatorDashboardName(r.Context(), dashboardId, name)
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
	ctx := r.Context()
	// check if user has reached the maximum number of groups
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	userInfo, err := h.dai.GetUserInfo(ctx, userId)
	if err != nil {
		handleErr(w, err)
		return
	}
	groupCount, err := h.dai.GetValidatorDashboardGroupCount(ctx, dashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if groupCount >= userInfo.PremiumPerks.ValidatorGroupsPerDashboard && !isUserAdmin(userInfo) {
		returnConflict(w, errors.New("maximum number of validator dashboard groups reached"))
		return
	}

	data, err := h.dai.CreateValidatorDashboardGroup(ctx, dashboardId, name)
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
	groupExists, err := h.dai.GetValidatorDashboardGroupExists(r.Context(), dashboardId, groupId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if !groupExists {
		returnNotFound(w, errors.New("group not found"))
		return
	}
	data, err := h.dai.UpdateValidatorDashboardGroup(r.Context(), dashboardId, groupId, name)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.ApiResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalDeleteValidatorDashboardGroup(w http.ResponseWriter, r *http.Request) {
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
	groupExists, err := h.dai.GetValidatorDashboardGroupExists(r.Context(), dashboardId, groupId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if !groupExists {
		returnNotFound(w, errors.New("group not found"))
		return
	}
	err = h.dai.RemoveValidatorDashboardGroup(r.Context(), dashboardId, groupId)
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
		GroupId           uint64        `json:"group_id,omitempty"`
		Validators        []intOrString `json:"validators,omitempty"`
		DepositAddress    string        `json:"deposit_address,omitempty"`
		WithdrawalAddress string        `json:"withdrawal_address,omitempty"`
		Graffiti          string        `json:"graffiti,omitempty"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	// check if exactly one of validators, deposit_address, withdrawal_address, graffiti is set
	fields := []interface{}{req.Validators, req.DepositAddress, req.WithdrawalAddress, req.Graffiti}
	var count int
	for _, set := range fields {
		if !reflect.ValueOf(set).IsZero() {
			count++
		}
	}
	if count != 1 {
		v.add("request body", "exactly one of `validators`, `deposit_address`, `withdrawal_address`, `graffiti` must be set. please check the API documentation for more information")
	}
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	groupId := req.GroupId
	ctx := r.Context()
	groupExists, err := h.dai.GetValidatorDashboardGroupExists(ctx, dashboardId, groupId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if !groupExists {
		returnNotFound(w, errors.New("group not found"))
		return
	}
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	userInfo, err := h.dai.GetUserInfo(ctx, userId)
	if err != nil {
		handleErr(w, err)
		return
	}
	limit := userInfo.PremiumPerks.ValidatorsPerDashboard
	if req.Validators == nil && !userInfo.PremiumPerks.BulkAdding && !isUserAdmin(userInfo) {
		returnConflict(w, errors.New("bulk adding not allowed with current subscription plan"))
		return
	}
	var data []types.VDBPostValidatorsData
	var dataErr error
	switch {
	case req.Validators != nil:
		indices, pubkeys := v.checkValidators(req.Validators, forbidEmpty)
		if v.hasErrors() {
			handleErr(w, v)
			return
		}
		validators, err := h.dai.GetValidatorsFromSlices(indices, pubkeys)
		if err != nil {
			handleErr(w, err)
			return
		}
		// check if adding more validators than allowed
		existingValidatorCount, err := h.dai.GetValidatorDashboardExistingValidatorCount(ctx, dashboardId, validators)
		if err != nil {
			handleErr(w, err)
			return
		}
		if uint64(len(validators)) > existingValidatorCount+limit {
			returnConflict(w, fmt.Errorf("adding more validators than allowed, limit is %v new validators", limit))
			return
		}
		data, dataErr = h.dai.AddValidatorDashboardValidators(ctx, dashboardId, groupId, validators)

	case req.DepositAddress != "":
		depositAddress := v.checkRegex(reEthereumAddress, req.DepositAddress, "deposit_address")
		if v.hasErrors() {
			handleErr(w, v)
			return
		}
		data, dataErr = h.dai.AddValidatorDashboardValidatorsByDepositAddress(ctx, dashboardId, groupId, depositAddress, limit)

	case req.WithdrawalAddress != "":
		withdrawalAddress := v.checkRegex(reWithdrawalCredential, req.WithdrawalAddress, "withdrawal_address")
		if v.hasErrors() {
			handleErr(w, v)
			return
		}
		data, dataErr = h.dai.AddValidatorDashboardValidatorsByWithdrawalAddress(ctx, dashboardId, groupId, withdrawalAddress, limit)

	case req.Graffiti != "":
		graffiti := v.checkRegex(reNonEmpty, req.Graffiti, "graffiti")
		if v.hasErrors() {
			handleErr(w, v)
			return
		}
		data, dataErr = h.dai.AddValidatorDashboardValidatorsByGraffiti(ctx, dashboardId, groupId, graffiti, limit)
	}

	if dataErr != nil {
		handleErr(w, dataErr)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}

	returnCreated(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
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
	data, paging, err := h.dai.GetValidatorDashboardValidators(r.Context(), *dashboardId, groupId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
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
	err = h.dai.RemoveValidatorDashboardValidators(r.Context(), dashboardId, validators)
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
	publicIdCount, err := h.dai.GetValidatorDashboardPublicIdCount(r.Context(), dashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if publicIdCount >= 1 {
		returnConflict(w, errors.New("cannot create more than one public id"))
		return
	}

	data, err := h.dai.CreateValidatorDashboardPublicId(r.Context(), dashboardId, name, req.ShareSettings.ShareGroups)
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
	dashboardInfo, err := h.dai.GetValidatorDashboardInfoByPublicId(r.Context(), publicDashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if dashboardInfo.Id != dashboardId {
		handleErr(w, newNotFoundErr("public id %v not found", publicDashboardId))
	}

	data, err := h.dai.UpdateValidatorDashboardPublicId(r.Context(), publicDashboardId, name, req.ShareSettings.ShareGroups)
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
	dashboardId := v.checkPrimaryDashboardId(vars["dashboard_id"])
	publicDashboardId := v.checkValidatorDashboardPublicId(vars["public_id"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	dashboardInfo, err := h.dai.GetValidatorDashboardInfoByPublicId(r.Context(), publicDashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}
	if dashboardInfo.Id != dashboardId {
		handleErr(w, newNotFoundErr("public id %v not found", publicDashboardId))
	}

	err = h.dai.RemoveValidatorDashboardPublicId(r.Context(), publicDashboardId)
	if err != nil {
		handleErr(w, err)
		return
	}

	returnNoContent(w)
}

func (h *HandlerService) InternalGetValidatorDashboardSlotViz(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}

	groupIds := v.checkExistingGroupIdList(r.URL.Query().Get("group_ids"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	data, err := h.dai.GetValidatorDashboardSlotViz(r.Context(), *dashboardId, groupIds)
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
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBSummaryColumn](&v, q.Get("sort"))
	protocolModes := v.checkProtocolModes(q.Get("modes"))

	period := checkEnum[enums.TimePeriod](&v, q.Get("period"), "period")
	// allowed periods are: all_time, last_30d, last_7d, last_24h, last_1h
	allowedPeriods := []enums.Enum{enums.TimePeriods.AllTime, enums.TimePeriods.Last30d, enums.TimePeriods.Last7d, enums.TimePeriods.Last24h, enums.TimePeriods.Last1h}
	v.checkEnumIsAllowed(period, allowedPeriods, "period")
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardSummary(r.Context(), *dashboardId, period, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit, protocolModes)
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
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	q := r.URL.Query()
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	if err != nil {
		handleErr(w, err)
		return
	}
	groupId := v.checkGroupId(vars["group_id"], forbidEmpty)
	period := checkEnum[enums.TimePeriod](&v, r.URL.Query().Get("period"), "period")
	// allowed periods are: all_time, last_30d, last_7d, last_24h, last_1h
	allowedPeriods := []enums.Enum{enums.TimePeriods.AllTime, enums.TimePeriods.Last30d, enums.TimePeriods.Last7d, enums.TimePeriods.Last24h, enums.TimePeriods.Last1h}
	v.checkEnumIsAllowed(period, allowedPeriods, "period")
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardGroupSummary(r.Context(), *dashboardId, groupId, period, protocolModes)
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
	var v validationError
	ctx := r.Context()
	dashboardId, err := h.handleDashboardId(ctx, mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	groupIds := v.checkGroupIdList(q.Get("group_ids"))
	efficiencyType := checkEnum[enums.VDBSummaryChartEfficiencyType](&v, q.Get("efficiency_type"), "efficiency_type")

	aggregation := checkEnum[enums.ChartAggregation](&v, r.URL.Query().Get("aggregation"), "aggregation")
	chartLimits, err := h.getCurrentChartTimeLimitsForDashboard(ctx, dashboardId, aggregation)
	if err != nil {
		handleErr(w, err)
		return
	}
	afterTs, beforeTs := v.checkTimestamps(r, chartLimits)
	if v.hasErrors() {
		handleErr(w, err)
		return
	}
	if afterTs < chartLimits.MinAllowedTs || beforeTs < chartLimits.MinAllowedTs {
		returnConflict(w, fmt.Errorf("requested time range is too old, minimum timestamp for dashboard owner's premium subscription for this aggregation is %v", chartLimits.MinAllowedTs))
		return
	}

	data, err := h.dai.GetValidatorDashboardSummaryChart(ctx, *dashboardId, groupIds, efficiencyType, aggregation, afterTs, beforeTs)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardSummaryChartResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardSummaryValidators(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	groupId := v.checkGroupId(r.URL.Query().Get("group_id"), allowEmpty)
	q := r.URL.Query()
	duty := checkEnum[enums.ValidatorDuty](&v, q.Get("duty"), "duty")
	period := checkEnum[enums.TimePeriod](&v, q.Get("period"), "period")
	// allowed periods are: all_time, last_30d, last_7d, last_24h, last_1h
	allowedPeriods := []enums.Enum{enums.TimePeriods.AllTime, enums.TimePeriods.Last30d, enums.TimePeriods.Last7d, enums.TimePeriods.Last24h, enums.TimePeriods.Last1h}
	v.checkEnumIsAllowed(period, allowedPeriods, "period")
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	// get indices based on duty
	var indices interface{}
	duties := enums.ValidatorDuties
	switch duty {
	case duties.None:
		indices, err = h.dai.GetValidatorDashboardSummaryValidators(r.Context(), *dashboardId, groupId)
	case duties.Sync:
		indices, err = h.dai.GetValidatorDashboardSyncSummaryValidators(r.Context(), *dashboardId, groupId, period)
	case duties.Slashed:
		indices, err = h.dai.GetValidatorDashboardSlashingsSummaryValidators(r.Context(), *dashboardId, groupId, period)
	case duties.Proposal:
		indices, err = h.dai.GetValidatorDashboardProposalSummaryValidators(r.Context(), *dashboardId, groupId, period)
	}
	if err != nil {
		handleErr(w, err)
		return
	}
	// map indices to response format
	data, err := mapVDBIndices(indices)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetValidatorDashboardSummaryValidatorsResponse{
		Data: data,
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardRewards(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBRewardsColumn](&v, q.Get("sort"))
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardRewards(r.Context(), *dashboardId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit, protocolModes)
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
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	groupId := v.checkGroupId(vars["group_id"], forbidEmpty)
	epoch := v.checkUint(vars["epoch"], "epoch")
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardGroupRewards(r.Context(), *dashboardId, groupId, epoch, protocolModes)
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
	var v validationError
	vars := mux.Vars(r)
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardRewardsChart(r.Context(), *dashboardId, protocolModes)
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
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	groupId := v.checkGroupId(q.Get("group_id"), allowEmpty)
	epoch := v.checkUint(vars["epoch"], "epoch")
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBDutiesColumn](&v, q.Get("sort"))
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardDuties(r.Context(), *dashboardId, epoch, groupId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit, protocolModes)
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
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBBlocksColumn](&v, q.Get("sort"))
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardBlocks(r.Context(), *dashboardId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit, protocolModes)
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

func (h *HandlerService) InternalGetValidatorDashboardHeatmap(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	aggregation := checkEnum[enums.ChartAggregation](&v, r.URL.Query().Get("aggregation"), "aggregation")
	chartLimits, err := h.getCurrentChartTimeLimitsForDashboard(r.Context(), dashboardId, aggregation)
	if err != nil {
		handleErr(w, err)
		return
	}
	afterTs, beforeTs := v.checkTimestamps(r, chartLimits)
	if v.hasErrors() {
		handleErr(w, err)
	}
	if afterTs < chartLimits.MinAllowedTs || beforeTs < chartLimits.MinAllowedTs {
		returnConflict(w, fmt.Errorf("requested time range is too old, minimum timestamp for dashboard owner's premium subscription for this aggregation is %v", chartLimits.MinAllowedTs))
		return
	}

	data, err := h.dai.GetValidatorDashboardHeatmap(r.Context(), *dashboardId, protocolModes, aggregation, afterTs, beforeTs)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardHeatmapResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupHeatmap(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	groupId := v.checkExistingGroupId(vars["group_id"])
	requestedTimestamp := v.checkUint(vars["timestamp"], "timestamp")
	protocolModes := v.checkProtocolModes(r.URL.Query().Get("modes"))
	aggregation := checkEnum[enums.ChartAggregation](&v, r.URL.Query().Get("aggregation"), "aggregation")
	if v.hasErrors() {
		handleErr(w, err)
	}
	chartLimits, err := h.getCurrentChartTimeLimitsForDashboard(r.Context(), dashboardId, aggregation)
	if err != nil {
		handleErr(w, err)
		return
	}
	if requestedTimestamp < chartLimits.MinAllowedTs || requestedTimestamp > chartLimits.LatestExportedTs {
		handleErr(w, newConflictErr("requested timestamp is outside of allowed chart history for dashboard owner's premium subscription"))
		return
	}

	data, err := h.dai.GetValidatorDashboardGroupHeatmap(r.Context(), *dashboardId, groupId, protocolModes, aggregation, requestedTimestamp)
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
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	pagingParams := v.checkPagingParams(r.URL.Query())
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardElDeposits(r.Context(), *dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
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
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	pagingParams := v.checkPagingParams(r.URL.Query())
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardClDeposits(r.Context(), *dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
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
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardTotalClDeposits(r.Context(), *dashboardId)
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
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardTotalElDeposits(r.Context(), *dashboardId)
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
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBWithdrawalsColumn](&v, q.Get("sort"))
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardWithdrawals(r.Context(), *dashboardId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit, protocolModes)
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
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	pagingParams := v.checkPagingParams(q)
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardTotalWithdrawals(r.Context(), *dashboardId, pagingParams.search, protocolModes)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetValidatorDashboardTotalWithdrawalsResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardRocketPool(w http.ResponseWriter, r *http.Request) {
	var v validationError
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBRocketPoolColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardRocketPool(r.Context(), *dashboardId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardRocketPoolResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardTotalRocketPool(w http.ResponseWriter, r *http.Request) {
	var v validationError
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	pagingParams := v.checkPagingParams(q)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardTotalRocketPool(r.Context(), *dashboardId, pagingParams.search)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardTotalRocketPoolResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardNodeRocketPool(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	// support ENS names ?
	nodeAddress := v.checkAddress(vars["node_address"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardNodeRocketPool(r.Context(), *dashboardId, nodeAddress)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardNodeRocketPoolResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetValidatorDashboardRocketPoolMinipools(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	if err != nil {
		handleErr(w, err)
		return
	}
	// support ENS names ?
	nodeAddress := v.checkAddress(vars["node_address"])
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBRocketPoolMinipoolsColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardRocketPoolMinipools(r.Context(), *dashboardId, nodeAddress, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetValidatorDashboardRocketPoolMinipoolsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

// --------------------------------------
// Notifications

func (h *HandlerService) InternalGetUserNotifications(w http.ResponseWriter, r *http.Request) {
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	data, err := h.dai.GetNotificationOverview(r.Context(), userId)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationsResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetUserNotificationDashboards(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.NotificationDashboardsColumn](&v, q.Get("sort"))
	chainId := v.checkNetworkParameter(q.Get("network"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	data, paging, err := h.dai.GetDashboardNotifications(r.Context(), userId, chainId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationDashboardsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetUserNotificationsValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	var v validationError
	notificationId := v.checkRegex(reNonEmpty, mux.Vars(r)["notification_id"], "notification_id")
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	data, err := h.dai.GetValidatorDashboardNotificationDetails(r.Context(), notificationId)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationsValidatorDashboardResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetUserNotificationsAccountDashboard(w http.ResponseWriter, r *http.Request) {
	var v validationError
	notificationId := v.checkRegex(reNonEmpty, mux.Vars(r)["notification_id"], "notification_id")
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	data, err := h.dai.GetAccountDashboardNotificationDetails(r.Context(), notificationId)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationsAccountDashboardResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetUserNotificationMachines(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.NotificationMachinesColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	data, paging, err := h.dai.GetMachineNotifications(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationMachinesResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetUserNotificationClients(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.NotificationClientsColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	data, paging, err := h.dai.GetClientNotifications(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationClientsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetUserNotificationRocketPool(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.NotificationRocketPoolColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	data, paging, err := h.dai.GetRocketPoolNotifications(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationRocketPoolResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetUserNotificationNetworks(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.NotificationNetworksColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	data, paging, err := h.dai.GetNetworkNotifications(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationNetworksResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetUserNotificationSettings(w http.ResponseWriter, r *http.Request) {
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	data, err := h.dai.GetNotificationSettings(r.Context(), userId)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationSettingsResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalPutUserNotificationSettingsGeneral(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	var req types.NotificationSettingsGeneral
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	checkMinMax(&v, req.MachineStorageUsageThreshold, 0, 1, "machine_storage_usage_threshold")
	checkMinMax(&v, req.MachineCpuUsageThreshold, 0, 1, "machine_cpu_usage_threshold")
	checkMinMax(&v, req.MachineMemoryUsageThreshold, 0, 1, "machine_memory_usage_threshold")
	checkMinMax(&v, req.RocketPoolMaxCollateralThreshold, 0, 1, "rocket_pool_max_collateral_threshold")
	checkMinMax(&v, req.RocketPoolMinCollateralThreshold, 0, 1, "rocket_pool_min_collateral_threshold")
	// TODO: check validity of clients
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	err = h.dai.UpdateNotificationSettingsGeneral(r.Context(), userId, req)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalPutUserNotificationSettingsGeneralResponse{
		Data: req,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalPutUserNotificationSettingsNetworks(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	var req types.NotificationSettingsNetwork
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	checkMinMax(&v, req.ParticipationRateThreshold, 0, 1, "participation_rate_threshold")

	chainId := v.checkNetworkParameter(mux.Vars(r)["network"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	err = h.dai.UpdateNotificationSettingsNetworks(r.Context(), userId, chainId, req)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalPutUserNotificationSettingsNetworksResponse{
		Data: types.NotificationNetwork{
			ChainId:  chainId,
			Settings: req,
		},
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalPutUserNotificationSettingsPairedDevices(w http.ResponseWriter, r *http.Request) {
	var v validationError
	req := struct {
		Name                   string `json:"name,omitempty"`
		IsNotificationsEnabled bool   `json:"is_notifications_enabled"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	// TODO use a better way to validate the paired device id
	pairedDeviceId := v.checkRegex(reNonEmpty, mux.Vars(r)["paired_device_id"], "paired_device_id")
	name := v.checkNameNotEmpty(req.Name)
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	err := h.dai.UpdateNotificationSettingsPairedDevice(r.Context(), pairedDeviceId, name, req.IsNotificationsEnabled)
	if err != nil {
		handleErr(w, err)
		return
	}
	// TODO timestamp
	response := types.InternalPutUserNotificationSettingsPairedDevicesResponse{
		Data: types.NotificationPairedDevice{
			Id:                     pairedDeviceId,
			Name:                   req.Name,
			IsNotificationsEnabled: req.IsNotificationsEnabled,
		},
	}

	returnOk(w, response)
}

func (h *HandlerService) InternalDeleteUserNotificationSettingsPairedDevices(w http.ResponseWriter, r *http.Request) {
	var v validationError
	// TODO use a better way to validate the paired device id
	pairedDeviceId := v.checkRegex(reNonEmpty, mux.Vars(r)["paired_device_id"], "paired_device_id")
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	err := h.dai.DeleteNotificationSettingsPairedDevice(r.Context(), pairedDeviceId)
	if err != nil {
		handleErr(w, err)
		return
	}
	returnNoContent(w)
}

func (h *HandlerService) InternalGetUserNotificationSettingsDashboards(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.NotificationSettingsDashboardColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	data, paging, err := h.dai.GetNotificationSettingsDashboards(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationSettingsDashboardsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalPutUserNotificationSettingsValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	var v validationError
	var req types.NotificationSettingsValidatorDashboard
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	checkMinMax(&v, req.GroupOfflineThreshold, 0, 1, "group_offline_threshold")
	vars := mux.Vars(r)
	dashboardId := v.checkPrimaryDashboardId(vars["dashboard_id"])
	groupId := v.checkExistingGroupId(vars["group_id"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	err := h.dai.UpdateNotificationSettingsValidatorDashboard(r.Context(), dashboardId, groupId, req)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalPutUserNotificationSettingsValidatorDashboardResponse{
		Data: req,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalPutUserNotificationSettingsAccountDashboard(w http.ResponseWriter, r *http.Request) {
	var v validationError
	req := struct {
		WebhookUrl                      string        `json:"webhook_url"`
		IsWebhookDiscordEnabled         bool          `json:"is_webhook_discord_enabled"`
		IsIgnoreSpamTransactionsEnabled bool          `json:"is_ignore_spam_transactions_enabled"`
		SubscribedChainIds              []intOrString `json:"subscribed_chain_ids"`

		IsIncomingTransactionsSubscribed  bool    `json:"is_incoming_transactions_subscribed"`
		IsOutgoingTransactionsSubscribed  bool    `json:"is_outgoing_transactions_subscribed"`
		IsERC20TokenTransfersSubscribed   bool    `json:"is_erc20_token_transfers_subscribed"`
		ERC20TokenTransfersValueThreshold float64 `json:"erc20_token_transfers_value_threshold"` // 0 does not disable, is_erc20_token_transfers_subscribed determines if it's enabled
		IsERC721TokenTransfersSubscribed  bool    `json:"is_erc721_token_transfers_subscribed"`
		IsERC1155TokenTransfersSubscribed bool    `json:"is_erc1155_token_transfers_subscribed"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	chainIdMap := v.checkNetworkSlice(req.SubscribedChainIds)
	// convert to uint64[] slice
	chainIds := make([]uint64, len(chainIdMap))
	i := 0
	for k := range chainIdMap {
		chainIds[i] = k
		i++
	}
	checkMinMax(&v, req.ERC20TokenTransfersValueThreshold, 0, math.MaxFloat64, "group_offline_threshold")
	vars := mux.Vars(r)
	dashboardId := v.checkPrimaryDashboardId(vars["dashboard_id"])
	groupId := v.checkExistingGroupId(vars["group_id"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	settings := types.NotificationSettingsAccountDashboard{
		WebhookUrl:                      req.WebhookUrl,
		IsWebhookDiscordEnabled:         req.IsWebhookDiscordEnabled,
		IsIgnoreSpamTransactionsEnabled: req.IsIgnoreSpamTransactionsEnabled,
		SubscribedChainIds:              chainIds,

		IsIncomingTransactionsSubscribed:  req.IsIncomingTransactionsSubscribed,
		IsOutgoingTransactionsSubscribed:  req.IsOutgoingTransactionsSubscribed,
		IsERC20TokenTransfersSubscribed:   req.IsERC20TokenTransfersSubscribed,
		ERC20TokenTransfersValueThreshold: req.ERC20TokenTransfersValueThreshold,
		IsERC721TokenTransfersSubscribed:  req.IsERC721TokenTransfersSubscribed,
		IsERC1155TokenTransfersSubscribed: req.IsERC1155TokenTransfersSubscribed,
	}
	err := h.dai.UpdateNotificationSettingsAccountDashboard(r.Context(), dashboardId, groupId, settings)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalPutUserNotificationSettingsAccountDashboardResponse{
		Data: settings,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalPostUserNotificationsTestEmail(w http.ResponseWriter, r *http.Request) {
	// TODO
	returnOk(w, nil)
}

func (h *HandlerService) InternalPostUserNotificationsTestPush(w http.ResponseWriter, r *http.Request) {
	// TODO
	returnOk(w, nil)
}

func (h *HandlerService) InternalPostUserNotificationsTestWebhook(w http.ResponseWriter, r *http.Request) {
	var v validationError
	req := struct {
		WebhookUrl              string `json:"webhook_url"`
		IsDiscordWebhookEnabled bool   `json:"is_discord_webhook_enabled,omitempty"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	// TODO
	returnOk(w, nil)
}

// --------------------------------------
// Blocks

func (h *HandlerService) InternalGetBlock(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetBlock(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetBlockOverview(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetBlockOverview(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetBlockOverviewResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetBlockTransactions(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetBlockTransactions(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockTransactionsResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetBlockVotes(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetBlockVotes(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockVotesResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetBlockAttestations(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetBlockAttestations(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockAttestationsResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetBlockWithdrawals(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetBlockWithdrawals(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockWtihdrawalsResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetBlockBlsChanges(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetBlockBlsChanges(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockBlsChangesResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetBlockVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetBlockVoluntaryExits(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockVoluntaryExitsResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetBlockBlobs(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetBlockBlobs(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockBlobsResponse{
		Data: data,
	}
	returnOk(w, response)
}

// --------------------------------------
// Slots

func (h *HandlerService) InternalGetSlot(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetSlot(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetSlotOverview(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetSlotOverview(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetBlockOverviewResponse{
		Data: *data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetSlotTransactions(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetSlotTransactions(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockTransactionsResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetSlotVotes(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetSlotVotes(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockVotesResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetSlotAttestations(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetSlotAttestations(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockAttestationsResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetSlotWithdrawals(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetSlotWithdrawals(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockWtihdrawalsResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetSlotBlsChanges(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetSlotBlsChanges(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockBlsChangesResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetSlotVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetSlotVoluntaryExits(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockVoluntaryExitsResponse{
		Data: data,
	}
	returnOk(w, response)
}

func (h *HandlerService) InternalGetSlotBlobs(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, err)
		return
	}

	data, err := h.dai.GetSlotBlobs(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, err)
		return
	}

	response := types.InternalGetBlockBlobsResponse{
		Data: data,
	}
	returnOk(w, response)
}
