package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gorilla/mux"
)

// All handler function names must include the HTTP method and the path they handle
// Public handlers may only be authenticated by an API key
// Public handlers must never call internal handlers

func (h *HandlerService) PublicGetHealthz(w http.ResponseWriter, r *http.Request) {
	var v validationError
	showAll := v.checkBool(r.URL.Query().Get("show_all"), "show_all")
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	data := h.dai.GetHealthz(ctx, showAll)

	responseCode := http.StatusOK
	if data.TotalOkPercentage != 1 {
		responseCode = http.StatusInternalServerError
	}
	writeResponse(w, r, responseCode, data)
}

func (h *HandlerService) PublicGetHealthzLoadbalancer(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetUserDashboards(w http.ResponseWriter, r *http.Request) {
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	data, err := h.dai.GetUserDashboards(r.Context(), userId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.ApiDataResponse[types.UserDashboardsData]{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicPostAccountDashboards(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, r, nil)
}

func (h *HandlerService) PublicGetAccountDashboard(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicDeleteAccountDashboard(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w, r)
}

func (h *HandlerService) PublicPostAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, r, nil)
}

func (h *HandlerService) PublicDeleteAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w, r)
}

func (h *HandlerService) PublicPostAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, r, nil)
}

func (h *HandlerService) PublicGetAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicDeleteAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w, r)
}

func (h *HandlerService) PublicPutAccountDashboardAccount(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicPostAccountDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, r, nil)
}

func (h *HandlerService) PublicPutAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicDeleteAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w, r)
}

func (h *HandlerService) PublicGetAccountDashboardTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicPutAccountDashboardTransactionsSettings(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicPostValidatorDashboards(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	type request struct {
		Name    string      `json:"name"`
		Network intOrString `json:"network" swaggertype:"string" enums:"ethereum,gnosis"`
	}
	var req request
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	name := v.checkNameNotEmpty(req.Name)
	chainId := v.checkNetwork(req.Network)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	userInfo, err := h.dai.GetUserInfo(r.Context(), userId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	dashboardCount, err := h.dai.GetUserValidatorDashboardCount(r.Context(), userId, true)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if dashboardCount >= userInfo.PremiumPerks.ValidatorDashboards && !isUserAdmin(userInfo) {
		returnConflict(w, r, errors.New("maximum number of validator dashboards reached"))
		return
	}

	data, err := h.dai.CreateValidatorDashboard(r.Context(), userId, name, chainId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.ApiDataResponse[types.VDBPostReturnData]{
		Data: *data,
	}
	returnCreated(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardIdParam := mux.Vars(r)["dashboard_id"]
	dashboardId, err := h.handleDashboardId(r.Context(), dashboardIdParam)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	q := r.URL.Query()
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, r, v)
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
		handleErr(w, r, err)
		return
	}

	// add premium chart perk info for shared dashboards
	premiumPerks, err := h.getDashboardPremiumPerks(r.Context(), *dashboardId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardOverview(r.Context(), *dashboardId, protocolModes)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	data.ChartHistorySeconds = premiumPerks.ChartHistorySeconds
	data.Name = name

	response := types.GetValidatorDashboardResponse{
		Data: *data,
	}

	returnOk(w, r, response)
}

func (h *HandlerService) PublicDeleteValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	err := h.dai.RemoveValidatorDashboard(r.Context(), dashboardId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	returnNoContent(w, r)
}

func (h *HandlerService) PublicPutValidatorDashboardName(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	type request struct {
		Name string `json:"name"`
	}
	var req request
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	name := v.checkNameNotEmpty(req.Name)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	data, err := h.dai.UpdateValidatorDashboardName(r.Context(), dashboardId, name)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.ApiDataResponse[types.VDBPostReturnData]{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicPostValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	type request struct {
		Name string `json:"name"`
	}
	var req request
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	name := v.checkNameNotEmpty(req.Name)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	ctx := r.Context()
	// check if user has reached the maximum number of groups
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	userInfo, err := h.dai.GetUserInfo(ctx, userId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	groupCount, err := h.dai.GetValidatorDashboardGroupCount(ctx, dashboardId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if groupCount >= userInfo.PremiumPerks.ValidatorGroupsPerDashboard && !isUserAdmin(userInfo) {
		returnConflict(w, r, errors.New("maximum number of validator dashboard groups reached"))
		return
	}

	data, err := h.dai.CreateValidatorDashboardGroup(ctx, dashboardId, name)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.ApiDataResponse[types.VDBPostCreateGroupData]{
		Data: *data,
	}

	returnCreated(w, r, response)
}

func (h *HandlerService) PublicPutValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId := v.checkPrimaryDashboardId(vars["dashboard_id"])
	groupId := v.checkExistingGroupId(vars["group_id"])
	type request struct {
		Name string `json:"name"`
	}
	var req request
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	name := v.checkNameNotEmpty(req.Name)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	groupExists, err := h.dai.GetValidatorDashboardGroupExists(r.Context(), dashboardId, groupId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if !groupExists {
		returnNotFound(w, r, errors.New("group not found"))
		return
	}
	data, err := h.dai.UpdateValidatorDashboardGroup(r.Context(), dashboardId, groupId, name)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.ApiDataResponse[types.VDBPostCreateGroupData]{
		Data: *data,
	}

	returnOk(w, r, response)
}

func (h *HandlerService) PublicDeleteValidatorDashboardGroup(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	groupId := v.checkExistingGroupId(vars["group_id"])
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	if groupId == types.DefaultGroupId {
		returnBadRequest(w, r, errors.New("cannot delete default group"))
		return
	}
	groupExists, err := h.dai.GetValidatorDashboardGroupExists(r.Context(), dashboardId, groupId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if !groupExists {
		returnNotFound(w, r, errors.New("group not found"))
		return
	}
	err = h.dai.RemoveValidatorDashboardGroup(r.Context(), dashboardId, groupId)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	returnNoContent(w, r)
}

func (h *HandlerService) PublicPostValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	type request struct {
		GroupId           uint64        `json:"group_id,omitempty" x-nullable:"true"`
		Validators        []intOrString `json:"validators,omitempty"`
		DepositAddress    string        `json:"deposit_address,omitempty"`
		WithdrawalAddress string        `json:"withdrawal_address,omitempty"`
		Graffiti          string        `json:"graffiti,omitempty"`
	}
	var req request
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	if v.hasErrors() {
		handleErr(w, r, v)
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
		handleErr(w, r, v)
		return
	}

	groupId := req.GroupId
	ctx := r.Context()
	groupExists, err := h.dai.GetValidatorDashboardGroupExists(ctx, dashboardId, groupId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if !groupExists {
		returnNotFound(w, r, errors.New("group not found"))
		return
	}
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	userInfo, err := h.dai.GetUserInfo(ctx, userId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	limit := userInfo.PremiumPerks.ValidatorsPerDashboard
	if req.Validators == nil && !userInfo.PremiumPerks.BulkAdding && !isUserAdmin(userInfo) {
		returnConflict(w, r, errors.New("bulk adding not allowed with current subscription plan"))
		return
	}
	var data []types.VDBPostValidatorsData
	var dataErr error
	switch {
	case req.Validators != nil:
		indices, pubkeys := v.checkValidators(req.Validators, forbidEmpty)
		if v.hasErrors() {
			handleErr(w, r, v)
			return
		}
		validators, err := h.dai.GetValidatorsFromSlices(indices, pubkeys)
		if err != nil {
			handleErr(w, r, err)
			return
		}
		// check if adding more validators than allowed
		existingValidatorCount, err := h.dai.GetValidatorDashboardExistingValidatorCount(ctx, dashboardId, validators)
		if err != nil {
			handleErr(w, r, err)
			return
		}
		if uint64(len(validators)) > existingValidatorCount+limit {
			returnConflict(w, r, fmt.Errorf("adding more validators than allowed, limit is %v new validators", limit))
			return
		}
		data, dataErr = h.dai.AddValidatorDashboardValidators(ctx, dashboardId, groupId, validators)

	case req.DepositAddress != "":
		depositAddress := v.checkRegex(reEthereumAddress, req.DepositAddress, "deposit_address")
		if v.hasErrors() {
			handleErr(w, r, v)
			return
		}
		data, dataErr = h.dai.AddValidatorDashboardValidatorsByDepositAddress(ctx, dashboardId, groupId, depositAddress, limit)

	case req.WithdrawalAddress != "":
		withdrawalAddress := v.checkRegex(reWithdrawalCredential, req.WithdrawalAddress, "withdrawal_address")
		if v.hasErrors() {
			handleErr(w, r, v)
			return
		}
		data, dataErr = h.dai.AddValidatorDashboardValidatorsByWithdrawalAddress(ctx, dashboardId, groupId, withdrawalAddress, limit)

	case req.Graffiti != "":
		graffiti := v.checkRegex(reNonEmpty, req.Graffiti, "graffiti")
		if v.hasErrors() {
			handleErr(w, r, v)
			return
		}
		data, dataErr = h.dai.AddValidatorDashboardValidatorsByGraffiti(ctx, dashboardId, groupId, graffiti, limit)
	}

	if dataErr != nil {
		handleErr(w, r, dataErr)
		return
	}
	response := types.ApiDataResponse[[]types.VDBPostValidatorsData]{
		Data: data,
	}

	returnCreated(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	q := r.URL.Query()
	groupId := v.checkGroupId(q.Get("group_id"), allowEmpty)
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBManageValidatorsColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	data, paging, err := h.dai.GetValidatorDashboardValidators(r.Context(), *dashboardId, groupId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardValidatorsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicDeleteValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	var indices []uint64
	var publicKeys []string
	req := struct {
		Validators []intOrString `json:"validators"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	indices, publicKeys = v.checkValidators(req.Validators, false)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	validators, err := h.dai.GetValidatorsFromSlices(indices, publicKeys)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	err = h.dai.RemoveValidatorDashboardValidators(r.Context(), dashboardId, validators)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	returnNoContent(w, r)
}

func (h *HandlerService) PublicPostValidatorDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	type request struct {
		Name          string `json:"name,omitempty"`
		ShareSettings struct {
			ShareGroups bool `json:"share_groups"`
		} `json:"share_settings"`
	}
	var req request
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	name := v.checkName(req.Name, 0)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	publicIdCount, err := h.dai.GetValidatorDashboardPublicIdCount(r.Context(), dashboardId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if publicIdCount >= 1 {
		returnConflict(w, r, errors.New("cannot create more than one public id"))
		return
	}

	data, err := h.dai.CreateValidatorDashboardPublicId(r.Context(), dashboardId, name, req.ShareSettings.ShareGroups)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}

	returnCreated(w, r, response)
}

func (h *HandlerService) PublicPutValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	type request struct {
		Name          string `json:"name,omitempty"`
		ShareSettings struct {
			ShareGroups bool `json:"share_groups"`
		} `json:"share_settings"`
	}
	var req request
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	name := v.checkName(req.Name, 0)
	publicDashboardId := v.checkValidatorDashboardPublicId(vars["public_id"])
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	fetchedId, err := h.dai.GetValidatorDashboardIdByPublicId(r.Context(), publicDashboardId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if *fetchedId != dashboardId {
		handleErr(w, r, newNotFoundErr("public id %v not found", publicDashboardId))
		return
	}

	data, err := h.dai.UpdateValidatorDashboardPublicId(r.Context(), publicDashboardId, name, req.ShareSettings.ShareGroups)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.ApiResponse{
		Data: data,
	}

	returnOk(w, r, response)
}

func (h *HandlerService) PublicDeleteValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId := v.checkPrimaryDashboardId(vars["dashboard_id"])
	publicDashboardId := v.checkValidatorDashboardPublicId(vars["public_id"])
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	fetchedId, err := h.dai.GetValidatorDashboardIdByPublicId(r.Context(), publicDashboardId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if *fetchedId != dashboardId {
		handleErr(w, r, newNotFoundErr("public id %v not found", publicDashboardId))
		return
	}

	err = h.dai.RemoveValidatorDashboardPublicId(r.Context(), publicDashboardId)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	returnNoContent(w, r)
}

func (h *HandlerService) PublicPutValidatorDashboardArchiving(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	req := struct {
		IsArchived bool `json:"is_archived"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	// check conditions for changing archival status
	dashboardInfo, err := h.dai.GetValidatorDashboardInfo(r.Context(), dashboardId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if dashboardInfo.IsArchived == req.IsArchived {
		// nothing to do
		returnOk(w, r, types.ApiDataResponse[types.VDBPostArchivingReturnData]{
			Data: types.VDBPostArchivingReturnData{Id: uint64(dashboardId), IsArchived: req.IsArchived},
		})
		return
	}

	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	dashboardCount, err := h.dai.GetUserValidatorDashboardCount(r.Context(), userId, !req.IsArchived)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	userInfo, err := h.dai.GetUserInfo(r.Context(), userId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if !isUserAdmin(userInfo) {
		if req.IsArchived {
			if dashboardCount >= MaxArchivedDashboardsCount {
				returnConflict(w, r, errors.New("maximum number of archived validator dashboards reached"))
				return
			}
		} else {
			if dashboardCount >= userInfo.PremiumPerks.ValidatorDashboards {
				returnConflict(w, r, errors.New("maximum number of active validator dashboards reached"))
				return
			}
			if dashboardInfo.GroupCount >= userInfo.PremiumPerks.ValidatorGroupsPerDashboard {
				returnConflict(w, r, errors.New("maximum number of groups in dashboards reached"))
				return
			}
			if dashboardInfo.ValidatorCount >= userInfo.PremiumPerks.ValidatorsPerDashboard {
				returnConflict(w, r, errors.New("maximum number of validators in dashboards reached"))
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
		handleErr(w, r, err)
		return
	}
	response := types.ApiDataResponse[types.VDBPostArchivingReturnData]{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardSlotViz(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}

	groupIds := v.checkExistingGroupIdList(r.URL.Query().Get("group_ids"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	data, err := h.dai.GetValidatorDashboardSlotViz(r.Context(), *dashboardId, groupIds)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardSlotVizResponse{
		Data: data,
	}

	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardSummary(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
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
		handleErr(w, r, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardSummary(r.Context(), *dashboardId, period, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit, protocolModes)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardSummaryResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardGroupSummary(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	q := r.URL.Query()
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	if err != nil {
		handleErr(w, r, err)
		return
	}
	groupId := v.checkGroupId(vars["group_id"], forbidEmpty)
	period := checkEnum[enums.TimePeriod](&v, r.URL.Query().Get("period"), "period")
	// allowed periods are: all_time, last_30d, last_7d, last_24h, last_1h
	allowedPeriods := []enums.Enum{enums.TimePeriods.AllTime, enums.TimePeriods.Last30d, enums.TimePeriods.Last7d, enums.TimePeriods.Last24h, enums.TimePeriods.Last1h}
	v.checkEnumIsAllowed(period, allowedPeriods, "period")
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardGroupSummary(r.Context(), *dashboardId, groupId, period, protocolModes)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardGroupSummaryResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardSummaryChart(w http.ResponseWriter, r *http.Request) {
	var v validationError
	ctx := r.Context()
	dashboardId, err := h.handleDashboardId(ctx, mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	q := r.URL.Query()
	groupIds := v.checkGroupIdList(q.Get("group_ids"))
	efficiencyType := checkEnum[enums.VDBSummaryChartEfficiencyType](&v, q.Get("efficiency_type"), "efficiency_type")

	aggregation := checkEnum[enums.ChartAggregation](&v, r.URL.Query().Get("aggregation"), "aggregation")
	chartLimits, err := h.getCurrentChartTimeLimitsForDashboard(ctx, dashboardId, aggregation)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	afterTs, beforeTs := v.checkTimestamps(r, chartLimits)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	if afterTs < chartLimits.MinAllowedTs || beforeTs < chartLimits.MinAllowedTs {
		returnConflict(w, r, fmt.Errorf("requested time range is too old, minimum timestamp for dashboard owner's premium subscription for this aggregation is %v", chartLimits.MinAllowedTs))
		return
	}

	data, err := h.dai.GetValidatorDashboardSummaryChart(ctx, *dashboardId, groupIds, efficiencyType, aggregation, afterTs, beforeTs)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardSummaryChartResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardSummaryValidators(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
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
		handleErr(w, r, v)
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
		handleErr(w, r, err)
		return
	}
	// map indices to response format
	data, err := mapVDBIndices(indices)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.GetValidatorDashboardSummaryValidatorsResponse{
		Data: data,
	}

	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardRewards(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBRewardsColumn](&v, q.Get("sort"))
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardRewards(r.Context(), *dashboardId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit, protocolModes)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardRewardsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardGroupRewards(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	groupId := v.checkGroupId(vars["group_id"], forbidEmpty)
	epoch := v.checkUint(vars["epoch"], "epoch")
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardGroupRewards(r.Context(), *dashboardId, groupId, epoch, protocolModes)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardGroupRewardsResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardRewardsChart(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardRewardsChart(r.Context(), *dashboardId, protocolModes)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardRewardsChartResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardDuties(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	q := r.URL.Query()
	groupId := v.checkGroupId(q.Get("group_id"), allowEmpty)
	epoch := v.checkUint(vars["epoch"], "epoch")
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBDutiesColumn](&v, q.Get("sort"))
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardDuties(r.Context(), *dashboardId, epoch, groupId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit, protocolModes)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardDutiesResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardBlocks(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBBlocksColumn](&v, q.Get("sort"))
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardBlocks(r.Context(), *dashboardId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit, protocolModes)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardBlocksResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardHeatmap(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	q := r.URL.Query()
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	aggregation := checkEnum[enums.ChartAggregation](&v, r.URL.Query().Get("aggregation"), "aggregation")
	chartLimits, err := h.getCurrentChartTimeLimitsForDashboard(r.Context(), dashboardId, aggregation)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	afterTs, beforeTs := v.checkTimestamps(r, chartLimits)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	if afterTs < chartLimits.MinAllowedTs || beforeTs < chartLimits.MinAllowedTs {
		returnConflict(w, r, fmt.Errorf("requested time range is too old, minimum timestamp for dashboard owner's premium subscription for this aggregation is %v", chartLimits.MinAllowedTs))
		return
	}

	data, err := h.dai.GetValidatorDashboardHeatmap(r.Context(), *dashboardId, protocolModes, aggregation, afterTs, beforeTs)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardHeatmapResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardGroupHeatmap(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	groupId := v.checkExistingGroupId(vars["group_id"])
	requestedTimestamp := v.checkUint(vars["timestamp"], "timestamp")
	protocolModes := v.checkProtocolModes(r.URL.Query().Get("modes"))
	aggregation := checkEnum[enums.ChartAggregation](&v, r.URL.Query().Get("aggregation"), "aggregation")
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	chartLimits, err := h.getCurrentChartTimeLimitsForDashboard(r.Context(), dashboardId, aggregation)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if requestedTimestamp < chartLimits.MinAllowedTs || requestedTimestamp > chartLimits.LatestExportedTs {
		handleErr(w, r, newConflictErr("requested timestamp is outside of allowed chart history for dashboard owner's premium subscription"))
		return
	}

	data, err := h.dai.GetValidatorDashboardGroupHeatmap(r.Context(), *dashboardId, groupId, protocolModes, aggregation, requestedTimestamp)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardGroupHeatmapResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	pagingParams := v.checkPagingParams(r.URL.Query())
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardElDeposits(r.Context(), *dashboardId, pagingParams.cursor, pagingParams.limit)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardExecutionLayerDepositsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	pagingParams := v.checkPagingParams(r.URL.Query())
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardClDeposits(r.Context(), *dashboardId, pagingParams.cursor, pagingParams.limit)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.GetValidatorDashboardConsensusLayerDepositsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardTotalConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardTotalClDeposits(r.Context(), *dashboardId)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.GetValidatorDashboardTotalConsensusDepositsResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardTotalExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	var err error
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	data, err := h.dai.GetValidatorDashboardTotalElDeposits(r.Context(), *dashboardId)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.GetValidatorDashboardTotalExecutionDepositsResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardWithdrawals(w http.ResponseWriter, r *http.Request) {
	var v validationError
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBWithdrawalsColumn](&v, q.Get("sort"))
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardWithdrawals(r.Context(), *dashboardId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit, protocolModes)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardWithdrawalsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardTotalWithdrawals(w http.ResponseWriter, r *http.Request) {
	var v validationError
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	pagingParams := v.checkPagingParams(q)
	protocolModes := v.checkProtocolModes(q.Get("modes"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardTotalWithdrawals(r.Context(), *dashboardId, pagingParams.search, protocolModes)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.GetValidatorDashboardTotalWithdrawalsResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardRocketPool(w http.ResponseWriter, r *http.Request) {
	var v validationError
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBRocketPoolColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardRocketPool(r.Context(), *dashboardId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardRocketPoolResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardTotalRocketPool(w http.ResponseWriter, r *http.Request) {
	var v validationError
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(r.Context(), mux.Vars(r)["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	pagingParams := v.checkPagingParams(q)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardTotalRocketPool(r.Context(), *dashboardId, pagingParams.search)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardTotalRocketPoolResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardNodeRocketPool(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	// support ENS names ?
	nodeAddress := v.checkAddress(vars["node_address"])
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, err := h.dai.GetValidatorDashboardNodeRocketPool(r.Context(), *dashboardId, nodeAddress)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardNodeRocketPoolResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetValidatorDashboardRocketPoolMinipools(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	q := r.URL.Query()
	dashboardId, err := h.handleDashboardId(r.Context(), vars["dashboard_id"])
	if err != nil {
		handleErr(w, r, err)
		return
	}
	// support ENS names ?
	nodeAddress := v.checkAddress(vars["node_address"])
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.VDBRocketPoolMinipoolsColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, paging, err := h.dai.GetValidatorDashboardRocketPoolMinipools(r.Context(), *dashboardId, nodeAddress, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetValidatorDashboardRocketPoolMinipoolsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) PublicGetNetworkValidators(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidator(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorDuties(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkAddressValidators(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkWithdrawalCredentialValidators(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorStatuses(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorLeaderboard(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorQueue(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkEpochs(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkEpoch(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkBlocks(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkBlock(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkSlots(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkSlot(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorBlocks(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkAddressPriorityFeeBlocks(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkAddressProposerRewardBlocks(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkForkedBlocks(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkForkedBlock(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkForkedSlot(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkBlockSizes(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorAttestations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkEpochAttestations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkSlotAttestations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkSlotVotes(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkBlockAttestations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkBlockVotes(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkAggregatedAttestations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkEthStore(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorRewardHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorBalanceHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorPerformanceHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkSlashings(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorSlashings(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkDeposits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorDeposits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkTransactionDeposits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkWithdrawals(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkSlotWithdrawals(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}
func (h *HandlerService) PublicGetNetworkBlockWithdrawals(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorWithdrawals(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkWithdrawalCredentialWithdrawals(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkEpochVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkSlotVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkBlockVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkAddressBalanceHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkAddressTokenSupplyHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkAddressEventLogs(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkTransaction(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkAddressTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkSlotTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkBlockTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkBlockBlobs(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkBlsChanges(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkEpochBlsChanges(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkSlotBlsChanges(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkBlockBlsChanges(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkValidatorBlsChanges(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkAddressEns(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkEns(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkBatches(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkLayer2ToLayer1Transactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkLayer1ToLayer2Transactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicPostNetworkBroadcasts(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, r, nil)
}

func (h *HandlerService) PublicGetEthPriceHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkGasNow(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkAverageGasLimitHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkGasUsedHistory(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetRocketPool(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetRocketPoolNodes(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetRocketPoolMinipools(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetNetworkSyncCommittee(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetMultisigSafe(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetMultisigSafeTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) PublicGetMultisigTransactionConfirmations(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}
