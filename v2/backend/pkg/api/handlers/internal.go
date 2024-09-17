package handlers

import (
	"errors"
	"math"
	"net/http"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	types "github.com/gobitfly/beaconchain/pkg/api/types"

	"github.com/gorilla/mux"
)

// --------------------------------------
// Premium Plans

func (h *HandlerService) InternalGetProductSummary(w http.ResponseWriter, r *http.Request) {
	data, err := h.dai.GetProductSummary(r.Context())
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetProductSummaryResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

// --------------------------------------
// API Ratelimit Weights

func (h *HandlerService) InternalGetRatelimitWeights(w http.ResponseWriter, r *http.Request) {
	data, err := h.dai.GetApiWeights(r.Context())
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetRatelimitWeightsResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

// --------------------------------------
// Latest State

func (h *HandlerService) InternalGetLatestState(w http.ResponseWriter, r *http.Request) {
	latestSlot, err := h.dai.GetLatestSlot()
	if err != nil {
		handleErr(w, r, err)
		return
	}

	finalizedEpoch, err := h.dai.GetLatestFinalizedEpoch()
	if err != nil {
		handleErr(w, r, err)
		return
	}

	exchangeRates, err := h.dai.GetLatestExchangeRates()
	if err != nil {
		handleErr(w, r, err)
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
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetRocketPool(w http.ResponseWriter, r *http.Request) {
	data, err := h.dai.GetRocketPoolOverview(r.Context())
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetRocketPoolResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

// All handler function names must include the HTTP method and the path they handle
// Internal handlers may only be authenticated by an OAuth token

// --------------------------------------
// Ad Configurations

func (h *HandlerService) InternalPostAdConfigurations(w http.ResponseWriter, r *http.Request) {
	var v validationError
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if user.UserGroup != types.UserGroupAdmin {
		returnForbidden(w, r, errors.New("user is not an admin"))
		return
	}

	var req types.AdConfigurationData
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
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
		returnBadRequest(w, r, errors.New("provide either banner_id or html_content"))
		return
	}
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	err = h.dai.CreateAdConfiguration(r.Context(), key, req.JQuerySelector, insertMode, req.RefreshInterval, req.ForAllUsers, req.BannerId, req.HtmlContent, req.Enabled)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.ApiDataResponse[types.AdConfigurationData]{
		Data: req,
	}
	returnCreated(w, r, response)
}

func (h *HandlerService) InternalGetAdConfigurations(w http.ResponseWriter, r *http.Request) {
	var v validationError
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if user.UserGroup != types.UserGroupAdmin {
		returnForbidden(w, r, errors.New("user is not an admin"))
		return
	}

	keys := v.checkAdConfigurationKeys(r.URL.Query().Get("keys"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	data, err := h.dai.GetAdConfigurations(r.Context(), keys)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.ApiDataResponse[[]types.AdConfigurationData]{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalPutAdConfiguration(w http.ResponseWriter, r *http.Request) {
	var v validationError
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if user.UserGroup != types.UserGroupAdmin {
		returnForbidden(w, r, errors.New("user is not an admin"))
		return
	}

	key := v.checkKeyNotEmpty(mux.Vars(r)["key"])
	var req types.AdConfigurationUpdateData
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
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
		returnConflict(w, r, errors.New("provide either banner_id or html_content"))
		return
	}
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	err = h.dai.UpdateAdConfiguration(r.Context(), key, req.JQuerySelector, insertMode, req.RefreshInterval, req.ForAllUsers, req.BannerId, req.HtmlContent, req.Enabled)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.ApiDataResponse[types.AdConfigurationData]{
		Data: types.AdConfigurationData{Key: key, AdConfigurationUpdateData: &req},
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalDeleteAdConfiguration(w http.ResponseWriter, r *http.Request) {
	var v validationError
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	if user.UserGroup != types.UserGroupAdmin {
		returnForbidden(w, r, errors.New("user is not an admin"))
		return
	}

	key := v.checkKeyNotEmpty(mux.Vars(r)["key"])
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}

	err = h.dai.RemoveAdConfiguration(r.Context(), key)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	returnNoContent(w, r)
}

// --------------------------------------
// User

func (h *HandlerService) InternalGetUserInfo(w http.ResponseWriter, r *http.Request) {
	// TODO patrick
	user, err := h.getUserBySession(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	userInfo, err := h.dai.GetUserInfo(r.Context(), user.Id)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetUserInfoResponse{
		Data: *userInfo,
	}
	returnOk(w, r, response)
}

// --------------------------------------
// Dashboards

func (h *HandlerService) InternalGetUserDashboards(w http.ResponseWriter, r *http.Request) {
	h.PublicGetUserDashboards(w, r)
}

// --------------------------------------
// Account Dashboards

func (h *HandlerService) InternalPostAccountDashboards(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, r, nil)
}

func (h *HandlerService) InternalGetAccountDashboard(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) InternalDeleteAccountDashboard(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w, r)
}

func (h *HandlerService) InternalPostAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, r, nil)
}

func (h *HandlerService) InternalDeleteAccountDashboardGroups(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w, r)
}

func (h *HandlerService) InternalPostAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, r, nil)
}

func (h *HandlerService) InternalGetAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) InternalDeleteAccountDashboardAccounts(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w, r)
}

func (h *HandlerService) InternalPutAccountDashboardAccount(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) InternalPostAccountDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	returnCreated(w, r, nil)
}

func (h *HandlerService) InternalPutAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) InternalDeleteAccountDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	returnNoContent(w, r)
}

func (h *HandlerService) InternalGetAccountDashboardTransactions(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

func (h *HandlerService) InternalPutAccountDashboardTransactionsSettings(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}

// --------------------------------------
// Validator Dashboards

func (h *HandlerService) InternalPostValidatorDashboards(w http.ResponseWriter, r *http.Request) {
	h.PublicPostValidatorDashboards(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboard(w, r)
}

func (h *HandlerService) InternalDeleteValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	h.PublicDeleteValidatorDashboard(w, r)
}

func (h *HandlerService) InternalPutValidatorDashboardName(w http.ResponseWriter, r *http.Request) {
	h.PublicPutValidatorDashboardName(w, r)
}

func (h *HandlerService) InternalPostValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	h.PublicPostValidatorDashboardGroups(w, r)
}

func (h *HandlerService) InternalPutValidatorDashboardGroups(w http.ResponseWriter, r *http.Request) {
	h.PublicPutValidatorDashboardGroups(w, r)
}

func (h *HandlerService) InternalDeleteValidatorDashboardGroup(w http.ResponseWriter, r *http.Request) {
	h.PublicDeleteValidatorDashboardGroup(w, r)
}

func (h *HandlerService) InternalPostValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	h.PublicPostValidatorDashboardValidators(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardValidators(w, r)
}

func (h *HandlerService) InternalDeleteValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	h.PublicDeleteValidatorDashboardValidators(w, r)
}

func (h *HandlerService) InternalPostValidatorDashboardPublicIds(w http.ResponseWriter, r *http.Request) {
	h.PublicPostValidatorDashboardPublicIds(w, r)
}

func (h *HandlerService) InternalPutValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	h.PublicPutValidatorDashboardPublicId(w, r)
}

func (h *HandlerService) InternalDeleteValidatorDashboardPublicId(w http.ResponseWriter, r *http.Request) {
	h.PublicDeleteValidatorDashboardPublicId(w, r)
}

func (h *HandlerService) InternalPutValidatorDashboardArchiving(w http.ResponseWriter, r *http.Request) {
	h.PublicPutValidatorDashboardArchiving(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardSlotViz(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardSlotViz(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardSummary(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardSummary(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupSummary(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardGroupSummary(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardSummaryChart(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardSummaryChart(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardSummaryValidators(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardSummaryValidators(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardRewards(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardRewards(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupRewards(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardGroupRewards(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardRewardsChart(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardRewardsChart(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardDuties(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardDuties(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardBlocks(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardBlocks(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardHeatmap(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardHeatmap(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardGroupHeatmap(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardGroupHeatmap(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardExecutionLayerDeposits(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardConsensusLayerDeposits(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardTotalConsensusLayerDeposits(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardTotalConsensusLayerDeposits(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardTotalExecutionLayerDeposits(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardTotalExecutionLayerDeposits(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardWithdrawals(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardWithdrawals(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardTotalWithdrawals(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardTotalWithdrawals(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardRocketPool(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardRocketPool(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardTotalRocketPool(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardTotalRocketPool(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardNodeRocketPool(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardNodeRocketPool(w, r)
}

func (h *HandlerService) InternalGetValidatorDashboardRocketPoolMinipools(w http.ResponseWriter, r *http.Request) {
	h.PublicGetValidatorDashboardRocketPoolMinipools(w, r)
}

// --------------------------------------
// Mobile

func (h *HandlerService) InternalGetMobileLatestBundle(w http.ResponseWriter, r *http.Request) {
	var v validationError
	q := r.URL.Query()
	force := v.checkBool(q.Get("force"), "force")
	bundleVersion := v.checkUint(q.Get("bundle_version"), "bundle_version")
	nativeVersion := v.checkUint(q.Get("native_version"), "native_version")
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	stats, err := h.dai.GetLatestBundleForNativeVersion(r.Context(), nativeVersion)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	var data types.MobileBundleData
	data.HasNativeUpdateAvailable = stats.MaxNativeVersion > nativeVersion
	// if given bundle version is smaller than the latest and delivery count is less than target count, return the latest bundle
	if force || (bundleVersion < stats.LatestBundleVersion && (stats.TargetCount == 0 || stats.DeliveryCount < stats.TargetCount)) {
		data.BundleUrl = stats.BundleUrl
	}
	response := types.GetMobileLatestBundleResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalPostMobileBundleDeliveries(w http.ResponseWriter, r *http.Request) {
	var v validationError
	vars := mux.Vars(r)
	bundleVersion := v.checkUint(vars["bundle_version"], "bundle_version")
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	err := h.dai.IncrementBundleDeliveryCount(r.Context(), bundleVersion)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	returnNoContent(w, r)
}

// --------------------------------------
// Notifications

func (h *HandlerService) InternalGetUserNotifications(w http.ResponseWriter, r *http.Request) {
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	data, err := h.dai.GetNotificationOverview(r.Context(), userId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetUserNotificationsResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetUserNotificationDashboards(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.NotificationDashboardsColumn](&v, q.Get("sort"))
	chainId := v.checkNetworkParameter(q.Get("network"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	data, paging, err := h.dai.GetDashboardNotifications(r.Context(), userId, chainId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetUserNotificationDashboardsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetUserNotificationsValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	var v validationError
	notificationId := v.checkRegex(reNonEmpty, mux.Vars(r)["notification_id"], "notification_id")
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	data, err := h.dai.GetValidatorDashboardNotificationDetails(r.Context(), notificationId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetUserNotificationsValidatorDashboardResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetUserNotificationsAccountDashboard(w http.ResponseWriter, r *http.Request) {
	var v validationError
	notificationId := v.checkRegex(reNonEmpty, mux.Vars(r)["notification_id"], "notification_id")
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	data, err := h.dai.GetAccountDashboardNotificationDetails(r.Context(), notificationId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetUserNotificationsAccountDashboardResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetUserNotificationMachines(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.NotificationMachinesColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	data, paging, err := h.dai.GetMachineNotifications(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetUserNotificationMachinesResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetUserNotificationClients(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.NotificationClientsColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	data, paging, err := h.dai.GetClientNotifications(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetUserNotificationClientsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetUserNotificationRocketPool(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.NotificationRocketPoolColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	data, paging, err := h.dai.GetRocketPoolNotifications(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetUserNotificationRocketPoolResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetUserNotificationNetworks(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.NotificationNetworksColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	data, paging, err := h.dai.GetNetworkNotifications(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetUserNotificationNetworksResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetUserNotificationSettings(w http.ResponseWriter, r *http.Request) {
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	data, err := h.dai.GetNotificationSettings(r.Context(), userId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetUserNotificationSettingsResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalPutUserNotificationSettingsGeneral(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	var req types.NotificationSettingsGeneral
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	checkMinMax(&v, req.MachineStorageUsageThreshold, 0, 1, "machine_storage_usage_threshold")
	checkMinMax(&v, req.MachineCpuUsageThreshold, 0, 1, "machine_cpu_usage_threshold")
	checkMinMax(&v, req.MachineMemoryUsageThreshold, 0, 1, "machine_memory_usage_threshold")
	checkMinMax(&v, req.RocketPoolMaxCollateralThreshold, 0, 1, "rocket_pool_max_collateral_threshold")
	checkMinMax(&v, req.RocketPoolMinCollateralThreshold, 0, 1, "rocket_pool_min_collateral_threshold")
	// TODO: check validity of clients
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	err = h.dai.UpdateNotificationSettingsGeneral(r.Context(), userId, req)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalPutUserNotificationSettingsGeneralResponse{
		Data: req,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalPutUserNotificationSettingsNetworks(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	var req types.NotificationSettingsNetwork
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	checkMinMax(&v, req.ParticipationRateThreshold, 0, 1, "participation_rate_threshold")

	chainId := v.checkNetworkParameter(mux.Vars(r)["network"])
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	err = h.dai.UpdateNotificationSettingsNetworks(r.Context(), userId, chainId, req)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalPutUserNotificationSettingsNetworksResponse{
		Data: types.NotificationNetwork{
			ChainId:  chainId,
			Settings: req,
		},
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalPutUserNotificationSettingsPairedDevices(w http.ResponseWriter, r *http.Request) {
	var v validationError
	req := struct {
		Name                   string `json:"name,omitempty"`
		IsNotificationsEnabled bool   `json:"is_notifications_enabled"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	// TODO use a better way to validate the paired device id
	pairedDeviceId := v.checkRegex(reNonEmpty, mux.Vars(r)["paired_device_id"], "paired_device_id")
	name := v.checkNameNotEmpty(req.Name)
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	err := h.dai.UpdateNotificationSettingsPairedDevice(r.Context(), pairedDeviceId, name, req.IsNotificationsEnabled)
	if err != nil {
		handleErr(w, r, err)
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

	returnOk(w, r, response)
}

func (h *HandlerService) InternalDeleteUserNotificationSettingsPairedDevices(w http.ResponseWriter, r *http.Request) {
	var v validationError
	// TODO use a better way to validate the paired device id
	pairedDeviceId := v.checkRegex(reNonEmpty, mux.Vars(r)["paired_device_id"], "paired_device_id")
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	err := h.dai.DeleteNotificationSettingsPairedDevice(r.Context(), pairedDeviceId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	returnNoContent(w, r)
}

func (h *HandlerService) InternalGetUserNotificationSettingsDashboards(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	q := r.URL.Query()
	pagingParams := v.checkPagingParams(q)
	sort := checkSort[enums.NotificationSettingsDashboardColumn](&v, q.Get("sort"))
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	data, paging, err := h.dai.GetNotificationSettingsDashboards(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetUserNotificationSettingsDashboardsResponse{
		Data:   data,
		Paging: *paging,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalPutUserNotificationSettingsValidatorDashboard(w http.ResponseWriter, r *http.Request) {
	var v validationError
	var req types.NotificationSettingsValidatorDashboard
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	checkMinMax(&v, req.GroupOfflineThreshold, 0, 1, "group_offline_threshold")
	vars := mux.Vars(r)
	dashboardId := v.checkPrimaryDashboardId(vars["dashboard_id"])
	groupId := v.checkExistingGroupId(vars["group_id"])
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	err := h.dai.UpdateNotificationSettingsValidatorDashboard(r.Context(), dashboardId, groupId, req)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalPutUserNotificationSettingsValidatorDashboardResponse{
		Data: req,
	}
	returnOk(w, r, response)
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
		handleErr(w, r, err)
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
		handleErr(w, r, v)
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
		handleErr(w, r, err)
		return
	}
	response := types.InternalPutUserNotificationSettingsAccountDashboardResponse{
		Data: settings,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalPostUserNotificationsTestEmail(w http.ResponseWriter, r *http.Request) {
	// TODO
	returnOk(w, r, nil)
}

func (h *HandlerService) InternalPostUserNotificationsTestPush(w http.ResponseWriter, r *http.Request) {
	// TODO
	returnOk(w, r, nil)
}

func (h *HandlerService) InternalPostUserNotificationsTestWebhook(w http.ResponseWriter, r *http.Request) {
	var v validationError
	req := struct {
		WebhookUrl              string `json:"webhook_url"`
		IsDiscordWebhookEnabled bool   `json:"is_discord_webhook_enabled,omitempty"`
	}{}
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, r, err)
		return
	}
	if v.hasErrors() {
		handleErr(w, r, v)
		return
	}
	// TODO
	returnOk(w, r, nil)
}

// --------------------------------------
// Blocks

func (h *HandlerService) InternalGetBlock(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetBlock(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetBlockOverview(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetBlockOverview(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetBlockOverviewResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetBlockTransactions(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetBlockTransactions(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockTransactionsResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetBlockVotes(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetBlockVotes(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockVotesResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetBlockAttestations(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetBlockAttestations(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockAttestationsResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetBlockWithdrawals(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetBlockWithdrawals(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockWtihdrawalsResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetBlockBlsChanges(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetBlockBlsChanges(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockBlsChangesResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetBlockVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetBlockVoluntaryExits(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockVoluntaryExitsResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetBlockBlobs(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "block")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetBlockBlobs(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockBlobsResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

// --------------------------------------
// Slots

func (h *HandlerService) InternalGetSlot(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetSlot(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetSlotOverview(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetSlotOverview(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.InternalGetBlockOverviewResponse{
		Data: *data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetSlotTransactions(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetSlotTransactions(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockTransactionsResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetSlotVotes(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetSlotVotes(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockVotesResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetSlotAttestations(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetSlotAttestations(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockAttestationsResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetSlotWithdrawals(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetSlotWithdrawals(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockWtihdrawalsResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetSlotBlsChanges(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetSlotBlsChanges(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockBlsChangesResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetSlotVoluntaryExits(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetSlotVoluntaryExits(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockVoluntaryExitsResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) InternalGetSlotBlobs(w http.ResponseWriter, r *http.Request) {
	chainId, block, err := h.validateBlockRequest(r, "slot")
	if err != nil {
		handleErr(w, r, err)
		return
	}

	data, err := h.dai.GetSlotBlobs(r.Context(), chainId, block)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	response := types.InternalGetBlockBlobsResponse{
		Data: data,
	}
	returnOk(w, r, response)
}

func (h *HandlerService) ReturnOk(w http.ResponseWriter, r *http.Request) {
	returnOk(w, r, nil)
}
