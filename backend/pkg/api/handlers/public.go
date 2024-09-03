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

//	@title			beaconcha.in API
//	@version		2.0
//	@description	To authenticate your API request beaconcha.in uses API Keys. Set your API Key either by:
//	@description	- Setting the `Authorization` header in the following format: `Authorization: Bearer <your-api-key>`. (recommended)
//	@description	- Setting the URL query parameter in the following format: `api_key={your_api_key}`.\
//	@description	Example: `https://beaconcha.in/api/v2/example?field=value&api_key={your_api_key}`

//	@host		beaconcha.in
//	@BasePath	/api/v2

//	@securitydefinitions.apikey	ApiKeyInHeader
//	@in							header
//	@name						Authorization
//	@description				Use your API key as a Bearer token, e.g. `Bearer <your-api-key>`

//	@securitydefinitions.apikey	ApiKeyInQuery
//	@in							query
//	@name						api_key

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

// PublicPostValidatorDashboards godoc
//
//	@Description	Create a new validator dashboard. **Note**: New dashboards will automatically have a default group created.
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Tags			Validator Dashboard Management
//	@Accept			json
//	@Produce		json
//	@Param			request	body		handlers.PublicPostValidatorDashboards.request	true	"`name`: Specify the name of the dashboard.<br>`network`: Specify the network for the dashboard. Possible options are:<ul><li>`ethereum`</li><li>`gnosis`</li></ul>"
//	@Success		201		{object}	types.PostValidatorDashboardsResponse
//	@Failure		400		{object}	types.ApiErrorResponse
//	@Failure		409		{object}	types.ApiErrorResponse	"Conflict. The request could not be performed by the server because the authenticated user has already reached their dashboard limit."
//	@Router			/validator-dashboards [post]
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
	if dashboardCount >= userInfo.PremiumPerks.ValidatorDasboards && !isUserAdmin(userInfo) {
		returnConflict(w, r, errors.New("maximum number of validator dashboards reached"))
		return
	}

	data, err := h.dai.CreateValidatorDashboard(r.Context(), userId, name, chainId)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.PostValidatorDashboardsResponse{
		Data: *data,
	}
	returnCreated(w, r, response)
}

// PublicGetValidatorDashboards godoc
//
//	@Description	Get overview information for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			modes			query		string	false	"Provide a comma seperated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``."
//	@Success		200				{object}	types.GetValidatorDashboardResponse
//	@Failure		400				{object}	types.ApiErrorResponse	"Bad Request"
//	@Router			/validator-dashboards/{dashboard_id} [get]
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

// PublicPutValidatorDashboard godoc
//
//	@Description	Delete a specified validator dashboard.
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Tags			Validator Dashboard Management
//	@Produce		json
//	@Param			dashboard_id	path	string	true	"The ID of the dashboard."
//	@Success		204				"Dashboard deleted successfully."
//	@Failure		400				{object}	types.ApiErrorResponse	"Bad Request"
//	@Router			/validator-dashboards/{dashboard_id} [delete]
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

// PublicPutValidatorDashboard godoc
//
//	@Description	Update the name of a specified validator dashboard.
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Tags			Validator Dashboard Management
//	@Accept			json
//	@Produce		json
//	@Param			dashboard_id	path		string												true	"The ID of the dashboard."
//	@Param			request			body		handlers.PublicPutValidatorDashboardName.request	true	"request"
//	@Success		200				{object}	types.ApiDataResponse[types.VDBPostReturnData]
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/name [put]
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

// PublicPostValidatorDashboardGroups godoc
//
//	@Description	Create a new group in a specified validator dashboard.
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Tags			Validator Dashboard Management
//	@Accept			json
//	@Produce		json
//	@Param			dashboard_id	path		string												true	"The ID of the dashboard."
//	@Param			request			body		handlers.PublicPostValidatorDashboardGroups.request	true	"request"
//	@Success		201				{object}	types.ApiDataResponse[types.VDBPostCreateGroupData]
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Failure		409				{object}	types.ApiErrorResponse	"Conflict. The request could not be performed by the server because the authenticated user has already reached their group limit."
//	@Router			/validator-dashboards/{dashboard_id}/groups [post]
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

// PublicGetValidatorDashboardGroups godoc
//
//	@Description	Update a groups name in a specified validator dashboard.
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Tags			Validator Dashboard Management
//	@Accept			json
//	@Produce		json
//	@Param			dashboard_id	path		string												true	"The ID of the dashboard."
//	@Param			group_id		path		string												true	"The ID of the group."
//	@Param			request			body		handlers.PublicPutValidatorDashboardGroups.request	true	"request"
//	@Success		200				{object}	types.ApiDataResponse[types.VDBPostCreateGroupData]
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/groups/{group_id} [put]
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

// PublicDeleteValidatorDashboardGroups godoc
//
//	@Description	Delete a group in a specified validator dashboard.
//	@Tags			Validator Dashboard Management
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Accept			json
//	@Produce		json
//	@Param			dashboard_id	path	string	true	"The ID of the dashboard."
//	@Param			group_id		path	string	true	"The ID of the group."
//	@Success		204				"Group deleted successfully."
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/groups/{group_id} [delete]
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

// PublicGetValidatorDashboardGroups godoc
//
//	@Description	Add new validators to a specified dashboard or update the group of already-added validators.
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Tags			Validator Dashboard Management
//	@Accept			json
//	@Produce		json
//	@Param			dashboard_id	path		string													true	"The ID of the dashboard."
//	@Param			request			body		handlers.PublicPostValidatorDashboardValidators.request	true	"`group_id`: (optional) Provide a single group id, to which all validators get added to. If ommited, the default group will be used.<br><br>To add validators, only one of the following fields can be set:<ul><li>`validators`: Provide a list of validator indices or public keys to add to the dashboard.</li><li>`deposit_address`: (limited to subscription tiers with 'Bulk adding') Provide a deposit address from which as many validators as possible will be added to the dashboard.</li><li>`withdrawal_address`: (limited to subscription tiers with 'Bulk adding') Provide a withdrawal address from which as many validators as possible will be added to the dashboard.</li><li>`graffiti`: (limited to subscription tiers with 'Bulk adding') Provide a graffiti string from which as many validators as possible will be added to the dashboard.</li></ul>"
//	@Success		201				{object}	types.ApiDataResponse[[]types.VDBPostValidatorsData]	"Returns a list of added validators."
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Failure		409				{object}	types.ApiErrorResponse	"Conflict. The request could not be performed by the server because the authenticated user has already reached their validator limit."
//	@Router			/validator-dashboards/{dashboard_id}/validators [post]
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

// PublicGetValidatorDashboardValidators godoc
//
//	@Description	Get a list of groups in a specified validator dashboard.
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			group_id		query		string	false	"The ID of the group."
//	@Param			cursor			query		string	false	"Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward."
//	@Param			limit			query		string	false	"The maximum number of results that may be returned."
//	@Param			sort			query		string	false	"The field you want to sort by. Append with `:desc` for descending order." Enums(index, public_key, balance, status, withdrawal_credentials)
//	@Param			search			query		string	false	"Search for Address, ENS."
//	@Success		200				{object}	types.GetValidatorDashboardValidatorsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/groups [get]
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

// PublicDeleteValidatorDashboardValidators godoc
//
//	@Description	Remove validators from a specified dashboard.
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Tags			Validator Dashboard Management
//	@Produce		json
//	@Param			dashboard_id	path	string	true	"The ID of the dashboard."
//	@Param			validators		query	string	false	"Provide a comma seperated list of validator indices or public keys to remove from the dashboard."
//	@Success		204				"Validators removed successfully."
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/validators [delete]
func (h *HandlerService) PublicDeleteValidatorDashboardValidators(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	indices, publicKeys := v.checkValidatorList(r.URL.Query().Get("validators"), forbidEmpty)
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

// PublicPostValidatorDashboardPublicIds godoc
//
//	@Description	Create a new public ID for a specified dashboard. This can be used as an ID by other users for non-modyfing (i.e. GET) endpoints only. Currently limited to one per dashboard.
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Tags			Validator Dashboard Management
//	@Accept			json
//	@Produce		json
//	@Param			dashboard_id	path		string													true	"The ID of the dashboard."
//	@Param			request			body		handlers.PublicPostValidatorDashboardPublicIds.request	true	"`name`: Provide a public name for the dashboard<br>`share_settings`:<ul><li>`share_groups`: If set to `true`, accessing the dashboard through the public ID will not reveal any group information.</li></ul>"
//	@Success		201				{object}	types.ApiDataResponse[types.VDBPublicId]
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Failure		409				{object}	types.ApiErrorResponse	"Conflict. The request could not be performed by the server because the authenticated user has already reached their public ID limit."
//	@Router			/validator-dashboards/{dashboard_id}/public-ids [post]
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
	response := types.ApiDataResponse[types.VDBPublicId]{
		Data: *data,
	}

	returnCreated(w, r, response)
}

// PublicPutValidatorDashboardPublicId godoc
//
//	@Description	Update a specified public ID for a specified dashboard.
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Tags			Validator Dashboard Management
//	@Accept			json
//	@Produce		json
//	@Param			dashboard_id	path		string													true	"The ID of the dashboard."
//	@Param			public_id		path		string													true	"The ID of the public ID."
//	@Param			request			body		handlers.PublicPutValidatorDashboardPublicId.request	true	"`name`: Provide a public name for the dashboard<br>`share_settings`:<ul><li>`share_groups`: If set to `true`, accessing the dashboard through the public ID will not reveal any group information.</li></ul>"
//	@Success		200				{object}	types.ApiDataResponse[types.VDBPublicId]
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/public-ids/{public_id} [put]
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
	response := types.ApiDataResponse[types.VDBPublicId]{
		Data: *data,
	}

	returnOk(w, r, response)
}

// PublicDeleteValidatorDashboardPublicId godoc
//
//	@Description	Delete a specified public ID for a specified dashboard.
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Tags			Validator Dashboard Management
//	@Produce		json
//	@Param			dashboard_id	path	string	true	"The ID of the dashboard."
//	@Param			public_id		path	string	true	"The ID of the public ID."
//	@Success		204				"Public ID deleted successfully."
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/public-ids/{public_id} [delete]
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

// PublicPutValidatorDashboardArchiving godoc
//
//	@Description	Archive or unarchive a specified validator dashboard. Archived dashboards cannot be accessed by other endpoints. Archiving happens automatically if the number of dashboards, validators, or groups exceeds the limit allowed by your subscription plan. For example, this might occur if you downgrade your subscription to a lower tier.
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Tags			Validator Dashboard Management
//	@Accept			json
//	@Produce		json
//	@Param			dashboard_id	path		string													true	"The ID of the dashboard."
//	@Param			request			body		handlers.PublicPutValidatorDashboardArchiving.request	true	"request"
//	@Success		200				{object}	types.ApiDataResponse[types.VDBPostArchivingReturnData]
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Conflict		409																			{object}	types.ApiErrorResponse	"Conflict. The request could not be performed by the server because the authenticated user has already reached their subscription limit."
//	@Router			/validator-dashboards/{dashboard_id}/archiving [put]
func (h *HandlerService) PublicPutValidatorDashboardArchiving(w http.ResponseWriter, r *http.Request) {
	var v validationError
	dashboardId := v.checkPrimaryDashboardId(mux.Vars(r)["dashboard_id"])
	type request struct {
		IsArchived bool `json:"is_archived"`
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
			if dashboardCount >= maxArchivedDashboardsCount {
				returnConflict(w, r, errors.New("maximum number of archived validator dashboards reached"))
				return
			}
		} else {
			if dashboardCount >= userInfo.PremiumPerks.ValidatorDasboards {
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

	data, err := h.dai.UpdateValidatorDashboardArchiving(r.Context(), dashboardId, req.IsArchived)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.ApiDataResponse[types.VDBPostArchivingReturnData]{
		Data: *data,
	}
	returnOk(w, r, response)
}

// PublicGetValidatorDashboardSlotViz godoc
//
//	@Description	Get slot viz information for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			group_ids		query		string	false	"Provide a comma seperated list of group IDs to filter the results by. If ommited, all groups will be included."
//	@Success		200				{object}	types.GetValidatorDashboardSlotVizResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/slot-viz [get]
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

var summaryAllowedPeriods = []enums.Enum{enums.TimePeriods.AllTime, enums.TimePeriods.Last30d, enums.TimePeriods.Last7d, enums.TimePeriods.Last24h, enums.TimePeriods.Last1h}

// PublicGetValidatorDashboardSummary godoc
//
//	@Description	Get summary information for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			period			query		string	true	"Time period to get data for."	Enums(all_time, last_30d, last_7d, last_24h, last_1h)
//	@Param			cursor			query		string	false	"Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward."
//	@Param			limit			query		string	false	"The maximum number of results that may be returned."
//	@Param			sort			query		string	false	"The field you want to sort by. Append with `:desc` for descending order." Enums(group_id, validators, efficiency, attestations, proposals, reward)
//	@Param			search			query		string	false	"Search for Index, Public Key, Group."
//	@Param			modes			query		string	false	"Provide a comma seperated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``."
//	@Success		200				{object}	types.GetValidatorDashboardSummaryResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/summary [get]
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
	v.checkEnumIsAllowed(period, summaryAllowedPeriods, "period")
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

// PublicGetValidatorDashboardGroupSummary godoc
//
//	@Description	Get summary information for a specified group in a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			group_id		path		string	true	"The ID of the group."
//	@Param			period			query		string	true	"Time period to get data for."	Enums(all_time, last_30d, last_7d, last_24h, last_1h)
//	@Param			modes			query		string	false	"Provide a comma seperated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``."
//	@Success		200				{object}	types.GetValidatorDashboardGroupSummaryResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/groups/{group_id}/summary [get]
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
	v.checkEnumIsAllowed(period, summaryAllowedPeriods, "period")
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

// PublicGetValidatorDashboardSummaryChart godoc
//
//	@Description	Get summary chart data for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			group_ids		query		string	false	"Provide a comma seperated list of group IDs to filter the results by."
//	@Param			efficiency_type	query		string	false	"Efficiency type to get data for."	Enums(all, attestation, sync, proposal)
//	@Param			aggregation		query		string	false	"Aggregation type to get data for."	Enums(epoch, hourly, daily, weekly) Default(hourly)
//	@Param			after_ts 		query		string	false	"Return data after this timestamp."
//	@Param			before_ts 		query		string	false	"Return data before this timestamp."
//	@Success		200				{object}	types.GetValidatorDashboardSummaryChartResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/summary-chart [get]
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

// PublicGetValidatorDashboardSummaryValidators godoc
//
//	@Description	Get summary information for validators in a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			group_id		query		string	false	"The ID of the group."
//	@Param			duty			query		string	false	"Validator duty to get data for."	Enums(none, sync, slashed, proposal) Default(none)
//	@Param			period			query		string	true	"Time period to get data for."	Enums(all_time, last_30d, last_7d, last_24h, last_1h)
//	@Success		200				{object}	types.GetValidatorDashboardSummaryValidatorsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/summary/validators [get]
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

// PublicGetValidatorDashboardRewards godoc
//
//	@Description	Get rewards information for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			cursor			query		string	false	"Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward."
//	@Param			limit			query		string	false	"The maximum number of results that may be returned."
//	@Param			sort			query		string	false	"The field you want to sort by. Append with `:desc` for descending order." Enums(epoch)
//	@Param			search			query		string	false	"Search for Epoch, Index, Public Key, Group."
//	@Param			modes			query		string	false	"Provide a comma seperated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``."
//	@Success		200				{object}	types.GetValidatorDashboardRewardsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/rewards [get]
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

// PublicGetValidatorDashboardGroupRewards godoc
//
//	@Description	Get rewards information for a specified group in a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			group_id		path		string	true	"The ID of the group."
//	@Param			epoch			path		string	true	"The epoch to get data for."
//	@Param			modes			query		string	false	"Provide a comma seperated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``."
//	@Success		200				{object}	types.GetValidatorDashboardGroupRewardsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/groups/{group_id}/rewards/{epoch} [get]
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

// PublicGetValidatorDashboardRewardsChart godoc
//
//	@Description	Get rewards chart data for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			modes			query		string	false	"Provide a comma seperated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``."
//	@Success		200				{object}	types.GetValidatorDashboardRewardsChartResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/rewards-chart [get]
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

// PublicGetValidatorDashboardDuties godoc
//
//	@Description	Get duties information for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			epoch			path		string	true	"The epoch to get data for."
//	@Param			group_id		query		string	false	"The ID of the group."
//	@Param			cursor			query		string	false	"Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward."
//	@Param			limit			query		string	false	"The maximum number of results that may be returned."
//	@Param			sort			query		string	false	"The field you want to sort by. Append with `:desc` for descending order." Enums(validator, reward)
//	@Param			search			query		string	false	"Search for Index, Public Key."
//	@Param			modes			query		string	false	"Provide a comma seperated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``."
//	@Success		200				{object}	types.GetValidatorDashboardDutiesResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/duties/{epoch} [get]
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

// PublicGetValidatorDashboardBlocks godoc
//
//	@Description	Get blocks information for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			cursor			query		string	false	"Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward."
//	@Param			limit			query		string	false	"The maximum number of results that may be returned."
//	@Param			sort			query		string	false	"The field you want to sort by. Append with `:desc` for descending order." Enums(proposer, slot, block, status, reward)
//	@Param			search			query		string	false	"Search for Index, Public Key, Group."
//	@Param			modes			query		string	false	"Provide a comma seperated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``."
//	@Success		200				{object}	types.GetValidatorDashboardBlocksResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/blocks [get]
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

// PublicGetValidatorDashboardHeatmap godoc
//
//	@Description	Get heatmap information for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			aggregation		query		string	false	"Aggregation type to get data for."	Enums(epoch, hourly, daily, weekly) Default(hourly)
//	@Param			after_ts 		query		string	false	"Return data after this timestamp."
//	@Param			before_ts 		query		string	false	"Return data before this timestamp."
//	@Param			modes			query		string	false	"Provide a comma seperated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``."
//	@Success		200				{object}	types.GetValidatorDashboardHeatmapResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/heatmap [get]
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

// PublicGetValidatorDashboardGroupHeatmap godoc
//
//	@Description	Get heatmap information for a specified group in a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			group_id		path		string	true	"The ID of the group."
//	@Param			timestamp		path		string	true	"The timestamp to get data for."
//	@Param			modes			query		string	false	"Provide a comma seperated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``."
//	@Param			aggregation		query		string	false	"Aggregation type to get data for."	Enums(epoch, hourly, daily, weekly) Default(hourly)
//	@Success		200				{object}	types.GetValidatorDashboardGroupHeatmapResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/groups/{group_id}/heatmap/{timestamp} [get]
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

// PublicGetValidatorDashboardExecutionLayerDeposits godoc
//
//	@Description	Get execution layer deposits information for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			cursor			query		string	false	"Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward."
//	@Param			limit			query		string	false	"The maximum number of results that may be returned."
//	@Success		200				{object}	types.GetValidatorDashboardExecutionLayerDepositsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/execution-layer-deposits [get]
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

	data, paging, err := h.dai.GetValidatorDashboardElDeposits(r.Context(), *dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
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

// PublicGetValidatorDashboardConsensusLayerDeposits godoc
//
//	@Description	Get consensus layer deposits information for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			cursor			query		string	false	"Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward."
//	@Param			limit			query		string	false	"The maximum number of results that may be returned."
//	@Success		200				{object}	types.GetValidatorDashboardConsensusLayerDepositsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/consensus-layer-deposits [get]
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

	data, paging, err := h.dai.GetValidatorDashboardClDeposits(r.Context(), *dashboardId, pagingParams.cursor, pagingParams.search, pagingParams.limit)
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

// PublicGetValidatorDashboardTotalConsensusLayerDeposits godoc
//
//	@Description	Get total consensus layer deposits information for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Success		200				{object}	types.GetValidatorDashboardTotalConsensusDepositsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/total-consensus-layer-deposits [get]
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

// PublicGetValidatorDashboardTotalExecutionLayerDeposits godoc
//
//	@Description	Get total execution layer deposits information for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Success		200				{object}	types.GetValidatorDashboardTotalExecutionDepositsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/total-execution-layer-deposits [get]
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

// PublicGetValidatorDashboardWithdrawals godoc
//
//	@Description	Get withdrawals information for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			cursor			query		string	false	"Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward."
//	@Param			limit			query		string	false	"The maximum number of results that may be returned."
//	@Param			sort			query		string	false	"The field you want to sort by. Append with `:desc` for descending order." Enums(epoch, slot, index, recipient, amount)
//	@Param			search			query		string	false	"Search for Index, Public Key, Address."
//	@Param			modes			query		string	false	"Provide a comma seperated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``."
//	@Success		200				{object}	types.GetValidatorDashboardWithdrawalsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/withdrawals [get]
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

// PublicGetValidatorDashboardTotalWithdrawals godoc
//
//	@Description	Get total withdrawals information for a specified dashboard
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			modes			query		string	false	"Provide a comma seperated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``."
//	@Success		200				{object}	types.GetValidatorDashboardTotalWithdrawalsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/total-withdrawals [get]
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

// PublicGetValidatorDashboardRocketPool godoc
//
//	@Description	Get an aggregated list of the Rocket Pool nodes details associated with a specified dashboard.
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			cursor			query		string	false	"Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward."
//	@Param			limit			query		string	false	"The maximum number of results that may be returned."
//	@Param			sort			query		string	false	"The field you want to sort by. Append with `:desc` for descending order." Enums(node, minipools, collateral, rpl, effective_rpl, rpl_apr, smoothing_pool)
//	@Param			search			query		string	false	"Search for Node address."
//	@Success		200				{object}	types.GetValidatorDashboardRocketPoolResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/rocket-pool [get]
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

// PublicGetValidatorDashboardTotalRocketPool godoc
//
//	@Description	Get a summary of all Rocket Pool nodes details associated with a specified dashboard.
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Success		200				{object}	types.GetValidatorDashboardTotalRocketPoolResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/total-rocket-pool [get]
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

// PublicGetValidatorDashboardNodeRocketPool godoc
//
//	@Description	Get details for a specific Rocket Pool node associated with a specified dashboard.
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			node_address	path		string	true	"The address of the node."
//	@Success		200				{object}	types.GetValidatorDashboardNodeRocketPoolResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/rocket-pool/{node_address} [get]
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

// PublicGetValidatorDashboardRocketPoolMinipools godoc
//
//	@Description	Get minipools information for a specified Rocket Pool node associated with a specified dashboard.
//	@Tags			Validator Dashboards
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			node_address	path		string	true	"The address of the node."
//	@Param			cursor			query		string	false	"Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward."
//	@Param			limit			query		string	false	"The maximum number of results that may be returned."
//	@Param			sort			query		string	false	"The field you want to sort by. Append with `:desc` for descending order. Possible values are TODO."
//	@Param			search			query		string	false	"Search for Index, Node."
//	@Success		200				{object}	types.GetValidatorDashboardRocketPoolMinipoolsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/rocket-pool/{node_address}/minipools [get]
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
