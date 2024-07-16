package handlers

import (
	"errors"
	"net/http"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gorilla/mux"
)

// TODO move to internal.go
// TODO make sure middleware sets the user id in the context

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
	data, paging, err := h.dai.GetNotificationDashboards(r.Context(), userId, chainId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationDashboards{
		Data:   data,
		Paging: *paging,
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
	data, paging, err := h.dai.GetNotificationMachines(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationMachines{
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
	data, paging, err := h.dai.GetNotificationClients(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationClients{
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
	data, paging, err := h.dai.GetNotificationRocketPool(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationRocketPool{
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
	data, paging, err := h.dai.GetNotificationNetworks(r.Context(), userId, pagingParams.cursor, *sort, pagingParams.search, pagingParams.limit)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalGetUserNotificationNetworks{
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
	// TODO validate the request
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
	chainId := v.checkNetworkParameter(mux.Vars(r)["network"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	// TODO validate the request
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
	pairedDeviceId := v.checkNameNotEmpty(mux.Vars(r)["paired_device_id"])
	// TODO validate the request
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
	pairedDeviceId := v.checkNameNotEmpty(mux.Vars(r)["paired_device_id"])
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
	// TODO validate the request
	vars := mux.Vars(r)
	dashboardId := v.checkPrimaryDashboardId(vars["dashboard_id"])
	groupId := v.checkExistingGroupId(vars["group_id"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	// TODO check if the dashboard belongs to the user
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
	var req types.NotificationSettingsAccountDashboard
	if err := v.checkBody(&req, r); err != nil {
		handleErr(w, err)
		return
	}
	// TODO validate the request
	vars := mux.Vars(r)
	dashboardId := v.checkPrimaryDashboardId(vars["dashboard_id"])
	groupId := v.checkExistingGroupId(vars["group_id"])
	if v.hasErrors() {
		handleErr(w, v)
		return
	}
	// TODO check if the dashboard belongs to the user
	err := h.dai.UpdateNotificationSettingsAccountDashboard(r.Context(), dashboardId, groupId, req)
	if err != nil {
		handleErr(w, err)
		return
	}
	response := types.InternalPutUserNotificationSettingsAccountDashboardResponse{
		Data: req,
	}
	returnOk(w, response)
}
