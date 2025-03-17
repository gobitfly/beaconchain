package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

// getDashboardPremiumPerks gets the premium perks of the dashboard OWNER or if it's a guest dashboard, it returns free tier premium perks
func (h *HandlerService) getDashboardPremiumPerks(ctx context.Context, id types.VDBId) (*types.PremiumPerks, error) {
	// for guest dashboards, return free tier perks
	if id.Validators != nil {
		perk, err := h.daService.GetFreeTierPerks(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting free tier perks: %w", err)
		}
		return perk, nil
	}
	// could be made into a single query if needed
	dashboardUser, err := h.daService.GetValidatorDashboardUser(ctx, id.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting dashboard owner: %w", err)
	}
	userInfo, err := h.daService.GetUserInfo(ctx, dashboardUser.UserId)
	if err != nil {
		if errors.Is(err, dataaccess.ErrNotFound) {
			log.Warn("user not found for dashboard owner, returning free tier perks", log.Fields{"dashboard_id": id.Id, "user_id_of_dashboard": dashboardUser.UserId})
			perk, err := h.daService.GetFreeTierPerks(ctx)
			if err != nil {
				return nil, fmt.Errorf("error getting free tier perks after user not found: %w", err)
			}
			return perk, nil
		}
		return nil, fmt.Errorf("error getting user info for dashboard owner: %w", err)
	}

	return &userInfo.PremiumPerks, nil
}

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

// PostValidatorDashboardValidators godoc
//
//	@Description	Add new validators to a specified dashboard or update the group of already-added validators. This endpoint will add all possible validators or return an error if the subscription plan limits are exceeded. The response will contain a list of added validators.
//	@Security		ApiKeyInHeader || ApiKeyInQuery
//	@Tags			Validator Dashboard Management
//	@Accept			json
//	@Produce		json
//	@Param			dashboard_id	path		integer													true	"The ID of the dashboard."
//	@Param			request			body		types.PostValidatorDashboardValidatorsRequest	true	"`group_id`: (optional) Provide a single group id, to which all validators get added to. If omitted, the default group will be used.<br><br>To add validators or update their group, only one of the following fields can be set:<ul><li>`validators`: Provide a list of validator indices or public keys.</li><li>`deposit_address`: (limited to subscription tiers with 'Bulk adding') Provide a deposit address from which all validators will be added to the dashboard, if possible.</li><li>`withdrawal_credential`: (limited to subscription tiers with 'Bulk adding') Provide a withdrawal credential from which all validators will be added to the dashboard, if possible.</li><li>`graffiti`: (limited to subscription tiers with 'Bulk adding') Provide a graffiti string from which all validators will be added to the dashboard, if possible.</li></ul>"
//	@Success		201				{object}	types.PostValidatorDashboardValidatorsResponse	"Returns a list of added validators."
//	@Failure		400				{object}	types.ApiErrorResponse
//	@Router			/validator-dashboards/{dashboard_id}/validators [post]
func (i *inputPostValidatorDashboardValidators) Validate(params map[string]string, body io.ReadCloser) error {
	var v validationError
	i.dashboardId = v.checkPrimaryDashboardId(params["dashboard_id"])
	req := types.PostValidatorDashboardValidatorsRequest{
		GroupId: types.DefaultGroupId, // default value
	}
	if err := v.checkBody(&req, body); err != nil {
		return err
	}
	i.groupId = req.GroupId
	// make sure exactly one of validators, deposit_address, withdrawal_credential, graffiti is set
	var setCount int

	if req.Validators != nil {
		setCount++
		i.validators = new(validatorsParam)
		i.validators.indices, i.validators.publicKeys = v.checkValidators(req.Validators, forbidEmpty)
	}
	if req.DepositAddress != "" {
		setCount++
		i.depositAddress = v.checkRegex(reEthereumAddress, req.DepositAddress, "deposit_address")
	}
	if req.WithdrawalCredential != "" {
		setCount++
		i.withdrawalCredential = v.checkRegex(reWithdrawalCredential, req.WithdrawalCredential, "withdrawal_credential")
	}
	if req.Graffiti != "" {
		setCount++
		i.graffiti = v.checkRegex(reGraffiti, req.Graffiti, "graffiti")
	}

	if setCount != 1 {
		v.add("body", "exactly one of `validators`, `deposit_address`, `withdrawal_credential`, `graffiti` must be set. Please check the API documentation for more information.")
	}

	return v.AsError()
}

type validatorsParam struct {
	indices    []types.VDBValidator
	publicKeys []string
}
type inputPostValidatorDashboardValidators struct {
	dashboardId types.VDBIdPrimary
	groupId     uint64
	// only one of the following fields can be set
	validators           *validatorsParam
	depositAddress       string
	withdrawalCredential string
	graffiti             string
}

func (h *HandlerService) PostValidatorDashboardValidators(ctx context.Context, input inputPostValidatorDashboardValidators) (types.PostValidatorDashboardValidatorsResponse, error) {
	var r types.PostValidatorDashboardValidatorsResponse

	// check if group exists
	groupExists, err := h.getDataAccessor(ctx).GetValidatorDashboardGroupExists(ctx, input.dashboardId, input.groupId)
	if err != nil {
		return r, fmt.Errorf("error checking group exists: %w", err)
	}
	if !groupExists {
		return r, newNotFoundErr("group not found")
	}

	// check if user has bulk adding enabled
	userId, err := GetUserIdByContext(ctx)
	if err != nil {
		return r, fmt.Errorf("error getting user id: %w", err)
	}
	userInfo, err := h.getDataAccessor(ctx).GetUserInfo(ctx, userId)
	if err != nil {
		return r, fmt.Errorf("error getting user info: %w", err)
	}
	if input.validators == nil && !userInfo.PremiumPerks.BulkAdding {
		return r, newForbiddenErr("bulk adding is not available for current subscription plan")
	}

	// get requested validators
	var requestedValidators []types.VDBValidator
	switch {
	case input.validators != nil:
		requestedValidators, err = h.getDataAccessor(ctx).GetValidatorsFromSlices(ctx, input.validators.indices, input.validators.publicKeys)
	case input.depositAddress != "":
		requestedValidators, err = h.getDataAccessor(ctx).GetValidatorsByDepositAddress(ctx, input.depositAddress)
	case input.withdrawalCredential != "":
		requestedValidators, err = h.getDataAccessor(ctx).GetValidatorsByWithdrawalCredentials(ctx, input.withdrawalCredential)
	case input.graffiti != "":
		requestedValidators, err = h.getDataAccessor(ctx).GetValidatorsByGraffiti(ctx, input.graffiti)
	}
	if err != nil {
		return r, fmt.Errorf("error getting requested validators: %w", err)
	}

	// get current EB space left for dashboard
	limitEBWei := userInfo.PremiumPerks.EffectiveBalancePerDashboard
	ebLimit := utils.GWeiToEther(limitEBWei.BigInt()).BigInt().Uint64()
	allExistingValidators, err := h.getDataAccessor(ctx).GetValidatorDashboardValidatorsOfList(ctx, input.dashboardId, nil /* fetches all validators */)
	if err != nil {
		return r, fmt.Errorf("error getting existing validators: %w", err)
	}
	existingEBs, err := h.getDataAccessor(ctx).GetValidatorsEffectiveBalances(ctx, allExistingValidators, false /* onlyActive */)
	if err != nil {
		return r, fmt.Errorf("error getting existing validators' EBs: %w", err)
	}
	var totalExistingEb uint64
	for _, eb := range existingEBs {
		totalExistingEb += eb
	}
	var ebSpaceLeft uint64
	if ebLimit > totalExistingEb {
		ebSpaceLeft = ebLimit - totalExistingEb
	}

	// get EB of new validators
	newValidators := make([]types.VDBValidator, 0, len(requestedValidators))
	for _, validator := range requestedValidators {
		if _, ok := existingEBs[validator]; ok {
			continue
		}
		newValidators = append(newValidators, validator)
	}
	requestedEbs, err := h.getDataAccessor(ctx).GetValidatorsEffectiveBalances(ctx, newValidators, false /* onlyActive */)
	if err != nil {
		return r, fmt.Errorf("error getting new validators' EBs: %w", err)
	}

	// determine if new validators exceed eb limit
	var totalNewEb uint64
	for _, validator := range newValidators {
		eb, ok := requestedEbs[validator]
		if !ok {
			return r, fmt.Errorf("effective balance not found for validator %d", validator)
		}
		if totalNewEb += eb; totalNewEb > ebSpaceLeft {
			return r, newConflictErr("validator addition exceeds dashboard's effective balance limit of current subscription plan")
		}
	}

	// insert validators / update groups
	insertedValidators, err := h.getDataAccessor(ctx).AddValidatorDashboardValidators(ctx, input.dashboardId, input.groupId, requestedValidators)
	if err != nil {
		return r, fmt.Errorf("error adding validators: %w", err)
	}
	r.Data = insertedValidators
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
func (h *HandlerService) GetValidatorDashboardSummaryChart(ctx context.Context, input inputGetValidatorDashboardSummaryChart) (types.GetValidatorDashboardSummaryChartResponse, error) {
	var r types.GetValidatorDashboardSummaryChartResponse
	dashboardId, err := h.getDashboardId(ctx, input.dashboardId)
	if err != nil {
		return r, err
	}

	dashboardPerks, err := h.getDashboardPremiumPerks(ctx, *dashboardId)
	if err != nil {
		return r, err
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
		return r, newForbiddenErr("requested aggregation is not available for dashboard owner's premium subscription")
	}

	latestExportedTs, err := h.getDataAccessor(ctx).GetLatestExportedChartTs(ctx, input.aggregation)
	if err != nil {
		return r, err
	}

	afterTs, beforeTs, err := resolveAndValidateTimestamps(
		input.afterTs,
		input.beforeTs,
		chartSeconds,
		input.aggregation.Duration(h.cfg.ClConfig.SecondsPerSlot*h.cfg.ClConfig.SlotsPerEpoch),
		latestExportedTs,
	)
	if err != nil {
		return r, err
	}

	data, err := h.getDataAccessor(ctx).GetValidatorDashboardSummaryChart(ctx, *dashboardId, input.groupIds, input.efficiencyType, input.aggregation, afterTs, beforeTs)
	if err != nil {
		return r, err
	}

	r.Data = *data
	return r, nil
}
