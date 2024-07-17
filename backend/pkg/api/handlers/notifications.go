package handlers

import (
	"errors"
	"math"
	"net/http"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gorilla/mux"
)

// TODO move to internal.go before merging

func (h *HandlerService) InternalGetUserNotifications(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(ctxUserIdKey).(uint64)
	if !ok {
		handleErr(w, errors.New("error getting user id from context"))
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
	userId, ok := r.Context().Value(ctxUserIdKey).(uint64)
	if !ok {
		handleErr(w, errors.New("error getting user id from context"))
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
	userId, ok := r.Context().Value(ctxUserIdKey).(uint64)
	if !ok {
		handleErr(w, errors.New("error getting user id from context"))
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
	userId, ok := r.Context().Value(ctxUserIdKey).(uint64)
	if !ok {
		handleErr(w, errors.New("error getting user id from context"))
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
	userId, ok := r.Context().Value(ctxUserIdKey).(uint64)
	if !ok {
		handleErr(w, errors.New("error getting user id from context"))
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
	userId, ok := r.Context().Value(ctxUserIdKey).(uint64)
	if !ok {
		handleErr(w, errors.New("error getting user id from context"))
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
	userId, ok := r.Context().Value(ctxUserIdKey).(uint64)
	if !ok {
		handleErr(w, errors.New("error getting user id from context"))
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
	userId, ok := r.Context().Value(ctxUserIdKey).(uint64)
	if !ok {
		handleErr(w, errors.New("error getting user id from context"))
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
	err := h.dai.UpdateNotificationSettingsGeneral(r.Context(), userId, req)
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
	userId, ok := r.Context().Value(ctxUserIdKey).(uint64)
	if !ok {
		handleErr(w, errors.New("error getting user id from context"))
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
	err := h.dai.UpdateNotificationSettingsNetworks(r.Context(), userId, chainId, req)
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
	v.checkNameNotEmpty(req.Name)
	err := h.dai.UpdateNotificationSettingsPairedDevice(r.Context(), pairedDeviceId, req.Name, req.IsNotificationsEnabled)
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
	err := h.dai.DeleteNotificationSettingsPairedDevice(r.Context(), pairedDeviceId)
	if err != nil {
		handleErr(w, err)
		return
	}
	returnNoContent(w)
}

func (h *HandlerService) InternalGetUserNotificationSettingsDashboards(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, ok := r.Context().Value(ctxUserIdKey).(uint64)
	if !ok {
		handleErr(w, errors.New("error getting user id from context"))
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
		SubscribedChainIds              []intOrString `json:"subscribed_networks"`

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
