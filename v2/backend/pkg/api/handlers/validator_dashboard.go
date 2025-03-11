package handlers

import (
	"context"
	"io"

	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/types"
)

// PostValidatorDashboardGroups godoc
//
//	@Description	Create a new group in a specified validator dashboard.
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Tags			Validator Dashboard Management
//	@Accept			json
//	@Produce		json
//	@Param			dashboard_id	path		integer												true	"The ID of the dashboard."
//	@Param			request			body		types.PostValidatorDashboardGroupsRequest	true	"request"
//	@Success		201				{object}	types.ApiDataResponse[types.VDBPostCreateGroupData]
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Failure		409				{object}	types.ApiErrorResponse	"Conflict. The request could not be performed by the server because the authenticated user has already reached their group limit."
//	@Router			/validator-dashboards/{dashboard_id}/groups [post]
func (i *inputPostValidatorDashboardGroups) Validate(params map[string]string, body io.ReadCloser) error {
	var v validationError
	var req types.PostValidatorDashboardGroupsRequest
	if err := v.checkBody(&req, body); err != nil {
		return err
	}
	i.dashboardId = v.checkPrimaryDashboardId(params["dashboard_id"])
	i.name = v.checkNameNotEmpty(req.Name)
	return v.AsError()
}

type inputPostValidatorDashboardGroups struct {
	dashboardId types.VDBIdPrimary
	name        string
}

func (h *HandlerService) PostValidatorDashboardGroups(ctx context.Context, input inputPostValidatorDashboardGroups) (types.ApiDataResponse[types.VDBPostCreateGroupData], error) {
	var r types.ApiDataResponse[types.VDBPostCreateGroupData]
	dataAccessor := h.getDataAccessor(ctx)
	userId, err := GetUserIdByContext(ctx)
	if err != nil {
		return r, err
	}
	userInfo, err := dataAccessor.GetUserInfo(ctx, userId)
	if err != nil {
		return r, err
	}
	groupCount, err := dataAccessor.GetValidatorDashboardGroupCount(ctx, input.dashboardId)
	if err != nil {
		return r, err
	}
	if groupCount >= userInfo.PremiumPerks.ValidatorGroupsPerDashboard {
		return r, newConflictErr("maximum number of validator dashboard groups reached")
	}

	data, err := dataAccessor.CreateValidatorDashboardGroup(ctx, input.dashboardId, input.name)
	if err != nil {
		return r, err
	}
	r.Data = *data
	return r, nil
}

// GetValidatorDashboardGroupSummary godoc
//
//	@Description	Get summary information for a specified group in a specified dashboard
//	@Tags			Validator Dashboard
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			group_id		path		integer	true	"The ID of the group."
//	@Param			period			query		string	true	"Time period to get data for."	Enums(all_time, last_30d, last_7d, last_24h, last_1h)
//	@Param			modes			query		string	false	"Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool``."
//	@Success		200				{object}	types.GetValidatorDashboardGroupSummaryResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/groups/{group_id}/summary [get]
func (i *inputGetValidatorDashboardGroupSummary) Validate(params map[string]string, _ io.ReadCloser) error {
	var v validationError
	i.dashboardIdParam = v.checkDashboardId(params["dashboard_id"])
	i.groupId = v.checkGroupId(params["group_id"], forbidEmpty)
	i.period = checkEnum[enums.TimePeriod](&v, params["period"], "period")
	i.protocolModes = v.checkProtocolModes(params["modes"])
	return v.AsError()
}

type inputGetValidatorDashboardGroupSummary struct {
	dashboardIdParam interface{}
	groupId          int64
	protocolModes    types.VDBProtocolModes
	period           enums.TimePeriod
}

func (h *HandlerService) GetValidatorDashboardGroupSummary(ctx context.Context, input inputGetValidatorDashboardGroupSummary) (types.GetValidatorDashboardGroupSummaryResponse, error) {
	var r types.GetValidatorDashboardGroupSummaryResponse
	dashboardId, err := h.getDashboardId(ctx, input.dashboardIdParam)
	if err != nil {
		return r, err
	}
	data, err := h.getDataAccessor(ctx).GetValidatorDashboardGroupSummary(ctx, *dashboardId, input.groupId, input.period, input.protocolModes)
	if err != nil {
		return r, err
	}
	r.Data = *data
	return r, nil
}
