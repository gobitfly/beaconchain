package handlers

import (
	"context"
	"io"
	"time"

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
//	@Param			modes			query		string	false	"Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool`."
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

// GetValidatorDashboardSummaryChart godoc
//
//	@Description	Get summary chart data for a specified dashboard
//	@Tags			Validator Dashboard
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			group_ids		query		string	false	"Provide a comma separated list of group IDs to filter the results by."
//	@Param			efficiency_type	query		string	false	"Efficiency type to get data for."	Enums(all, attestation, sync, proposal)
//	@Param			aggregation		query		string	false	"Aggregation type to get data for."	Enums(epoch, hourly, daily, weekly)	Default(hourly)
//	@Param			after_ts		query		string	false	"Return data after this timestamp."
//	@Param			before_ts		query		string	false	"Return data before this timestamp."
//	@Success		200				{object}	types.GetValidatorDashboardSummaryChartResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/summary-chart [get]
func (i *inputGetValidatorDashboardSummaryChart) Validate(params map[string]string, _ io.ReadCloser) error {
	var v validationError
	i.dashboardId = v.checkDashboardId(params["dashboard_id"])
	i.groupIds = v.checkGroupIdList(params["group_ids"])
	i.efficiencyType = checkEnum[enums.VDBSummaryChartEfficiencyType](&v, params["efficiency_type"], "efficiency_type")
	i.aggregation = checkEnum[enums.ChartAggregation](&v, params["aggregation"], "aggregation")
	afterTsParam := params["after_ts"]
	beforeTsParam := params["before_ts"]
	if afterTsParam != "" {
		afterTs := v.checkUint(afterTsParam, "after_ts")
		i.afterTs = &afterTs
	}
	if beforeTsParam != "" {
		beforeTs := v.checkUint(beforeTsParam, "before_ts")
		i.beforeTs = &beforeTs
	}
	if i.afterTs != nil && i.beforeTs != nil && *i.afterTs >= *i.beforeTs {
		v.add("after_ts", "after_ts must be less than before_ts")
	}
	return v.AsError()
}

type inputGetValidatorDashboardSummaryChart struct {
	dashboardId    interface{}
	groupIds       []int64
	efficiencyType enums.VDBSummaryChartEfficiencyType
	aggregation    enums.ChartAggregation
	afterTs        *uint64
	beforeTs       *uint64
}

const chartDatapointLimit uint64 = 200

// resolveAndValidateTimestamps calculates the missing timestamps for the chart data and/or validates them.
func resolveAndValidateTimestamps(
	afterTs *uint64,
	beforeTs *uint64,
	chartSeconds uint64,
	aggregationDuration time.Duration,
	latestExportedTs uint64,
) (uint64, uint64, error) {
	maxAllowedInterval := chartDatapointLimit * uint64(aggregationDuration.Seconds())
	minAllowedTs := latestExportedTs - min(chartSeconds, latestExportedTs)
	// Resolve missing timestamps based on the provided input.
	var resolvedAfterTs, resolvedBeforeTs uint64
	switch {
	case afterTs == nil && beforeTs == nil:
		intervalLookback := latestExportedTs - min(maxAllowedInterval, latestExportedTs)
		resolvedAfterTs = max(minAllowedTs, intervalLookback)
		resolvedBeforeTs = latestExportedTs
	case afterTs == nil && beforeTs != nil: // beforeTs is provided
		intervalLookback := *beforeTs - min(maxAllowedInterval, *beforeTs)
		resolvedAfterTs = max(minAllowedTs, intervalLookback)
		resolvedBeforeTs = *beforeTs
	case afterTs != nil && beforeTs == nil: // afterTs is provided
		resolvedAfterTs = *afterTs
		resolvedBeforeTs = *afterTs + maxAllowedInterval
	case afterTs != nil && beforeTs != nil: // both are provided
		resolvedAfterTs = *afterTs
		resolvedBeforeTs = *beforeTs
	}

	// Validate the resolved timestamps.
	if resolvedAfterTs < minAllowedTs {
		return 0, 0, newConflictErr("`after_ts` must be greater or equal to %d", minAllowedTs)
	}
	if resolvedBeforeTs < minAllowedTs {
		return 0, 0, newConflictErr("`before_ts` must be greater or equal to %d", minAllowedTs)
	}
	if resolvedBeforeTs-resolvedAfterTs > maxAllowedInterval {
		return 0, 0, newBadRequestErr("difference between `before_ts` and `after_ts` must be smaller or equal to %d", maxAllowedInterval)
	}
	return resolvedAfterTs, resolvedBeforeTs, nil
}

// Main handler using the simplified timestamp resolution.
func (h *HandlerService) GetValidatorDashboardSummaryChart(ctx context.Context, input inputGetValidatorDashboardSummaryChart) (*types.GetValidatorDashboardSummaryChartResponse, error) {
	dashboardId, err := h.getDashboardId(ctx, input.dashboardId)
	if err != nil {
		return nil, err
	}

	dashboardPerks, err := h.getDashboardPremiumPerks(ctx, *dashboardId)
	if err != nil {
		return nil, err
	}
	perkSeconds := dashboardPerks.ChartHistorySeconds
	aggregations := enums.ChartAggregations
	var chartSeconds uint64
	switch input.aggregation {
	case aggregations.Epoch:
		chartSeconds = perkSeconds.Epoch
	case aggregations.Hourly:
		chartSeconds = perkSeconds.Hourly
	case aggregations.Daily:
		chartSeconds = perkSeconds.Daily
	case aggregations.Weekly:
		chartSeconds = perkSeconds.Weekly
	}
	if chartSeconds == 0 {
		return nil, newForbiddenErr("requested aggregation is not available for dashboard owner's premium subscription")
	}

	latestExportedTs, err := h.getDataAccessor(ctx).GetLatestExportedChartTs(ctx, input.aggregation)
	if err != nil {
		return nil, err
	}

	afterTs, beforeTs, err := resolveAndValidateTimestamps(
		input.afterTs,
		input.beforeTs,
		chartSeconds,
		input.aggregation.Duration(h.cfg.ClConfig.SecondsPerSlot*h.cfg.ClConfig.SlotsPerEpoch),
		latestExportedTs,
	)
	if err != nil {
		return nil, err
	}

	data, err := h.getDataAccessor(ctx).GetValidatorDashboardSummaryChart(ctx, *dashboardId, input.groupIds, input.efficiencyType, input.aggregation, afterTs, beforeTs)
	if err != nil {
		return nil, err
	}

	return &types.GetValidatorDashboardSummaryChartResponse{
		Data: *data,
	}, nil
}

// GetValidatorDashboardExecutionLayerWithdrawals godoc
//
//	@Description	Get withdrawals information (EL) for a specified dashboard
//	@Tags			Validator Dashboard
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			cursor			query		string	false	"Return data for the given cursor value. Pass the `paging.next_cursor` value of the previous response to navigate to forward, or pass the `paging.prev_cursor` value of the previous response to navigate to backward."
//	@Param			limit			query		string	false	"The maximum number of results that may be returned."
//	@Param			sort			query		string	false	"The field you want to sort by. Append with `:desc` for descending order."	Enums(epoch, slot, index, recipient, amount)
//	@Param			search			query		string	false	"Search for Index, Block, Address, Group, Public Key, Transaction Hash."
//	@Param			modes			query		string	false	"Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool`."
//	@Success		200				{object}	types.GetValidatorDashboardExecutionLayerWithdrawalsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/execution-layer-withdrawals [get]
func (i *inputGetValidatorDashboardExecutionLayerWithdrawals) Validate(params map[string]string, _ io.ReadCloser) error {
	var v validationError
	i.dashboardId = v.checkDashboardId(params["dashboard_id"])
	i.protocolModes = v.checkProtocolModes(params["modes"])
	i.sort = checkSort[enums.VDBWithdrawalsElColumn](&v, params["sort"])
	return v.AsError()
}

type inputGetValidatorDashboardExecutionLayerWithdrawals struct {
	Paging
	protocolModes types.VDBProtocolModes
	sort          types.Sort[enums.VDBWithdrawalsElColumn]
	dashboardId   interface{}
}

func (h *HandlerService) GetValidatorDashboardExecutionLayerWithdrawals(ctx context.Context, input inputGetValidatorDashboardExecutionLayerWithdrawals) (types.GetValidatorDashboardExecutionLayerWithdrawalsResponse, error) {
	var r types.GetValidatorDashboardExecutionLayerWithdrawalsResponse
	dashboardId, err := h.getDashboardId(ctx, input.dashboardId)
	if err != nil {
		return r, err
	}
	data, paging, err := h.getDataAccessor(ctx).GetValidatorDashboardElWithdrawals(ctx, *dashboardId, input.cursor, input.sort, input.search, input.limit, input.protocolModes)
	if err != nil {
		return r, err
	}
	r.Data = data
	r.Paging = *paging
	return r, nil
}

// PublicGetValidatorDashboardConsolidations godoc
//
//	@Description	Get consolidations information for a specified dashboard
//	@Tags			Validator Dashboard
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			cursor			query		string	false	"Return data for the given cursor value. Pass the `paging.next_cursor`` value of the previous response to navigate to forward, or pass the `paging.prev_cursor`` value of the previous response to navigate to backward."
//	@Param			limit			query		string	false	"The maximum number of results that may be returned."
//	@Param			sort			query		string	false	"The field you want to sort by. Append with `:desc` for descending order."	Enums(epoch, slot, index, recipient, amount)
//	@Param			search			query		string	false	"Search for Index, Public Key, Address."
//	@Success		200				{object}	types.GetValidatorDashboardConsolidationsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/consolidations [get]
func (i *inputGetValidatorDashboardConsolidations) Validate(params map[string]string, body io.ReadCloser) error {
	var v validationError
	i.Paging = v.checkPagingMap(params)
	i.sort = checkSort[enums.VDBConsolidationsColumn](&v, params["sort"])
	i.dashboardId = v.checkDashboardId(params["dashboard_id"])
	return v.AsError()
}

type inputGetValidatorDashboardConsolidations struct {
	Paging
	sort        types.Sort[enums.VDBConsolidationsColumn]
	dashboardId interface{}
}

func (h *HandlerService) GetValidatorDashboardConsolidations(ctx context.Context, input inputGetValidatorDashboardConsolidations) (types.GetValidatorDashboardConsolidationsResponse, error) {
	var r types.GetValidatorDashboardConsolidationsResponse
	dashboardId, err := h.getDashboardId(ctx, input.dashboardId)
	if err != nil {
		return r, err
	}
	data, paging, err := h.getDataAccessor(ctx).GetValidatorDashboardConsolidations(ctx, *dashboardId, input.cursor, input.sort, input.search, input.limit)
	if err != nil {
		return r, err
	}
	r.Data = data
	r.Paging = *paging
	return r, nil
}

// GetValidatorDashboardConsensusLayerWithdrawals godoc
//
//	@Description	Get withdrawals information (CL) for a specified dashboard
//	@Tags			Validator Dashboard
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			cursor			query		string	false	"Return data for the given cursor value. Pass the `paging.next_cursor` value of the previous response to navigate to forward, or pass the `paging.prev_cursor` value of the previous response to navigate to backward."
//	@Param			limit			query		string	false	"The maximum number of results that may be returned."
//	@Param			sort			query		string	false	"The field you want to sort by. Append with `:desc` for descending order."	Enums(epoch, slot, index, recipient, amount)
//	@Param			search			query		string	false	"Search for Index, Slot, Group, Public Key, Recipient."
//	@Param			modes			query		string	false	"Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool`."
//	@Success		200				{object}	types.GetValidatorDashboardConsensusLayerWithdrawalsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/consensus-layer-withdrawals [get]
func (i *inputGetValidatorDashboardConsensusLayerWithdrawals) Validate(params map[string]string, _ io.ReadCloser) error {
	var v validationError
	i.dashboardId = v.checkDashboardId(params["dashboard_id"])
	i.protocolModes = v.checkProtocolModes(params["modes"])
	i.sort = checkSort[enums.VDBWithdrawalsClColumn](&v, params["sort"])
	return v.AsError()
}

type inputGetValidatorDashboardConsensusLayerWithdrawals struct {
	Paging
	protocolModes types.VDBProtocolModes
	sort          types.Sort[enums.VDBWithdrawalsClColumn]
	dashboardId   interface{}
}

func (h *HandlerService) GetValidatorDashboardConsensusLayerWithdrawals(ctx context.Context, input inputGetValidatorDashboardConsensusLayerWithdrawals) (types.GetValidatorDashboardConsensusLayerWithdrawalsResponse, error) {
	var r types.GetValidatorDashboardConsensusLayerWithdrawalsResponse
	dashboardId, err := h.getDashboardId(ctx, input.dashboardId)
	if err != nil {
		return r, err
	}
	data, paging, err := h.getDataAccessor(ctx).GetValidatorDashboardClWithdrawals(ctx, *dashboardId, input.cursor, input.sort, input.search, input.limit, input.protocolModes)
	if err != nil {
		return r, err
	}
	r.Data = data
	r.Paging = *paging
	return r, nil
}

// GetValidatorDashboardTotalExecutionLayerWithdrawals godoc
//
//	@Description	Get total withdrawals information (EL) for a specified dashboard
//	@Tags			Validator Dashboard
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			search			query		string	false	"Search for Index, Block, Address, Group, Public Key, Transaction Hash."
//	@Param			modes			query		string	false	"Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool`."
//	@Success		200				{object}	types.GetValidatorDashboardTotalExecutionWithdrawalsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/total-execution-layer-withdrawals [get]
func (i *inputGetValidatorDashboardTotalExecutionLayerWithdrawals) Validate(params map[string]string, _ io.ReadCloser) error {
	var v validationError
	i.dashboardId = v.checkDashboardId(params["dashboard_id"])
	i.protocolModes = v.checkProtocolModes(params["modes"])
	i.search = params["search"]
	return v.AsError()
}

type inputGetValidatorDashboardTotalExecutionLayerWithdrawals struct {
	protocolModes types.VDBProtocolModes
	search        string
	dashboardId   interface{}
}

func (h *HandlerService) GetValidatorDashboardTotalExecutionLayerWithdrawals(ctx context.Context, input inputGetValidatorDashboardTotalExecutionLayerWithdrawals) (types.GetValidatorDashboardTotalExecutionWithdrawalsResponse, error) {
	var r types.GetValidatorDashboardTotalExecutionWithdrawalsResponse
	dashboardId, err := h.getDashboardId(ctx, input.dashboardId)
	if err != nil {
		return r, err
	}

	data, err := h.getDataAccessor(ctx).GetValidatorDashboardTotalElWithdrawals(ctx, *dashboardId, input.search, input.protocolModes)
	if err != nil {
		return r, err
	}
	r.Data = *data
	return r, nil
}

// GetValidatorDashboardTotalConsensusLayerWithdrawals godoc
//
//	@Description	Get total withdrawals information (CL) for a specified dashboard
//	@Tags			Validator Dashboard
//	@Produce		json
//	@Param			dashboard_id	path		string	true	"The ID of the dashboard."
//	@Param			search			query		string	false	"Search for Index, Slot, Group, Public Key, Recipient."
//	@Param			modes			query		string	false	"Provide a comma separated list of protocol modes which should be respected for validator calculations. Possible values are `rocket_pool`."
//	@Success		200				{object}	types.GetValidatorDashboardTotalConsensusWithdrawalsResponse
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/total-consensus-layer-withdrawals [get]
func (i *inputGetValidatorDashboardTotalConsensusLayerWithdrawals) Validate(params map[string]string, _ io.ReadCloser) error {
	var v validationError
	i.dashboardId = v.checkDashboardId(params["dashboard_id"])
	i.protocolModes = v.checkProtocolModes(params["modes"])
	i.search = params["search"]
	return v.AsError()
}

type inputGetValidatorDashboardTotalConsensusLayerWithdrawals struct {
	protocolModes types.VDBProtocolModes
	search        string
	dashboardId   interface{}
}

func (h *HandlerService) GetValidatorDashboardTotalConsensusLayerWithdrawals(ctx context.Context, input inputGetValidatorDashboardTotalConsensusLayerWithdrawals) (types.GetValidatorDashboardTotalConsensusWithdrawalsResponse, error) {
	var r types.GetValidatorDashboardTotalConsensusWithdrawalsResponse
	dashboardId, err := h.getDashboardId(ctx, input.dashboardId)
	if err != nil {
		return r, err
	}

	data, err := h.getDataAccessor(ctx).GetValidatorDashboardTotalClWithdrawals(ctx, *dashboardId, input.search, input.protocolModes)
	if err != nil {
		return r, err
	}
	r.Data = *data
	return r, nil
}
