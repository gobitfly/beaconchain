package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/invopop/jsonschema"

	"github.com/alexedwards/scs/v2"
	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/services"
	types "github.com/gobitfly/beaconchain/pkg/api/types"
)

type HandlerService struct {
	daService                   dataaccess.DataAccessor
	daDummy                     dataaccess.DataAccessor
	scs                         *scs.SessionManager
	isPostMachineMetricsEnabled bool // if more config options are needed, consider having the whole config in here
}

func NewHandlerService(dataAccessor dataaccess.DataAccessor, dummy dataaccess.DataAccessor, sessionManager *scs.SessionManager, enablePostMachineMetrics bool) *HandlerService {
	if allNetworks == nil {
		networks, err := dataAccessor.GetAllNetworks()
		if err != nil {
			log.Fatal(err, "error getting networks for handler", 0, nil)
		}
		allNetworks = networks
	}

	return &HandlerService{
		daService:                   dataAccessor,
		daDummy:                     dummy,
		scs:                         sessionManager,
		isPostMachineMetricsEnabled: enablePostMachineMetrics,
	}
}

// getDataAccessor returns the correct data accessor based on the request context.
// if the request is mocked, the data access dummy is returned; otherwise the data access service.
// should only be used if getting mocked data for the endpoint is appropriate
func (h *HandlerService) getDataAccessor(r *http.Request) dataaccess.DataAccessor {
	if isMocked(r) {
		return h.daDummy
	}
	return h.daService
}

// all networks available in the system, filled on startup in NewHandlerService
var allNetworks []types.NetworkInfo

// --------------------------------------
// errors

var (
	errMsgParsingId    = errors.New("error parsing parameter 'dashboard_id'")
	errBadRequest      = errors.New("bad request")
	errInternalServer  = errors.New("internal server error")
	errUnauthorized    = errors.New("unauthorized")
	errForbidden       = errors.New("forbidden")
	errConflict        = errors.New("conflict")
	errTooManyRequests = errors.New("too many requests")
	errGone            = errors.New("gone")
)

// --------------------------------------
// utility functions

type validatorSet struct {
	Indexes    []types.VDBValidator
	PublicKeys []string
}

// parseDashboardId is a helper function to validate the string dashboard id param.
func parseDashboardId(id string) (interface{}, error) {
	var v validationError
	if reInteger.MatchString(id) {
		// given id is a normal id
		id := v.checkUint(id, "dashboard_id")
		if v.hasErrors() {
			return nil, v
		}
		return types.VDBIdPrimary(id), nil
	}
	if reValidatorDashboardPublicId.MatchString(id) {
		// given id is a public id
		return types.VDBIdPublic(id), nil
	}
	// given id must be an encoded set of validators
	decodedId, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		return nil, newBadRequestErr("given value '%s' is not a valid dashboard id", id)
	}
	indexes, publicKeys := v.checkValidatorList(string(decodedId), forbidEmpty)
	if v.hasErrors() {
		return nil, newBadRequestErr("given value '%s' is not a valid dashboard id", id)
	}
	return validatorSet{Indexes: indexes, PublicKeys: publicKeys}, nil
}

// getDashboardId is a helper function to convert the dashboard id param to a VDBId.
// precondition: dashboardIdParam must be a valid dashboard id and either a primary id, public id, or list of validators.
func (h *HandlerService) getDashboardId(ctx context.Context, dashboardIdParam interface{}) (*types.VDBId, error) {
	switch dashboardId := dashboardIdParam.(type) {
	case types.VDBIdPrimary:
		return &types.VDBId{Id: dashboardId, Validators: nil}, nil
	case types.VDBIdPublic:
		dashboardInfo, err := h.daService.GetValidatorDashboardPublicId(ctx, dashboardId)
		if err != nil {
			return nil, err
		}
		return &types.VDBId{Id: types.VDBIdPrimary(dashboardInfo.DashboardId), Validators: nil, AggregateGroups: !dashboardInfo.ShareSettings.ShareGroups}, nil
	case validatorSet:
		validators, err := h.daService.GetValidatorsFromSlices(dashboardId.Indexes, dashboardId.PublicKeys)
		if err != nil {
			return nil, err
		}
		if len(validators) == 0 {
			return nil, newNotFoundErr("no validators found for given id")
		}
		if len(validators) > maxValidatorsInList {
			return nil, newBadRequestErr("too many validators in list, maximum is %d", maxValidatorsInList)
		}
		return &types.VDBId{Validators: validators}, nil
	}
	return nil, errMsgParsingId
}

// handleDashboardId is a helper function to both validate the dashboard id param and convert it to a VDBId.
// it should be used as the last validation step for all internal dashboard GET-handlers.
// Modifying handlers (POST, PUT, DELETE) should only accept primary dashboard ids and just use checkPrimaryDashboardId.
func (h *HandlerService) handleDashboardId(ctx context.Context, param string) (*types.VDBId, error) {
	// validate dashboard id param
	dashboardIdParam, err := parseDashboardId(param)
	if err != nil {
		return nil, err
	}
	// convert to VDBId
	dashboardId, err := h.getDashboardId(ctx, dashboardIdParam)
	if err != nil {
		return nil, err
	}

	return dashboardId, nil
}

const chartDatapointLimit uint64 = 200

type ChartTimeDashboardLimits struct {
	MinAllowedTs       uint64
	LatestExportedTs   uint64
	MaxAllowedInterval uint64
}

// helper function to retrieve allowed chart timestamp boundaries according to the users premium perks at the current point in time
func (h *HandlerService) getCurrentChartTimeLimitsForDashboard(ctx context.Context, dashboardId *types.VDBId, aggregation enums.ChartAggregation) (ChartTimeDashboardLimits, error) {
	limits := ChartTimeDashboardLimits{}
	var err error
	premiumPerks, err := h.getDashboardPremiumPerks(ctx, *dashboardId)
	if err != nil {
		return limits, err
	}

	maxAge := getMaxChartAge(aggregation, premiumPerks.ChartHistorySeconds) // can be max int for unlimited, always check for underflows
	if maxAge == 0 {
		return limits, newConflictErr("requested aggregation is not available for dashboard owner's premium subscription")
	}
	limits.LatestExportedTs, err = h.daService.GetLatestExportedChartTs(ctx, aggregation)
	if err != nil {
		return limits, err
	}
	limits.MinAllowedTs = limits.LatestExportedTs - min(maxAge, limits.LatestExportedTs)                        // min to prevent underflow
	secondsPerEpoch := uint64(12 * 32)                                                                          // TODO: fetch dashboards chain id and use correct value for network once available
	limits.MaxAllowedInterval = chartDatapointLimit*uint64(aggregation.Duration(secondsPerEpoch).Seconds()) - 1 // -1 to make sure we don't go over the limit

	return limits, nil
}

// getDashboardPremiumPerks gets the premium perks of the dashboard OWNER or if it's a guest dashboard, it returns free tier premium perks
func (h *HandlerService) getDashboardPremiumPerks(ctx context.Context, id types.VDBId) (*types.PremiumPerks, error) {
	// for guest dashboards, return free tier perks
	if id.Validators != nil {
		perk, err := h.daService.GetFreeTierPerks(ctx)
		if err != nil {
			return nil, err
		}
		return perk, nil
	}
	// could be made into a single query if needed
	dashboardUser, err := h.daService.GetValidatorDashboardUser(ctx, id.Id)
	if err != nil {
		return nil, err
	}
	userInfo, err := h.daService.GetUserInfo(ctx, dashboardUser.UserId)
	if err != nil {
		return nil, err
	}

	return &userInfo.PremiumPerks, nil
}

// getMaxChartAge returns the maximum age of a chart in seconds based on the given aggregation type and premium perks
func getMaxChartAge(aggregation enums.ChartAggregation, perkSeconds types.ChartHistorySeconds) uint64 {
	aggregations := enums.ChartAggregations
	switch aggregation {
	case aggregations.Epoch:
		return perkSeconds.Epoch
	case aggregations.Hourly:
		return perkSeconds.Hourly
	case aggregations.Daily:
		return perkSeconds.Daily
	case aggregations.Weekly:
		return perkSeconds.Weekly
	default:
		return 0
	}
}

func isUserAdmin(user *types.UserInfo) bool {
	if user == nil {
		return false
	}
	return user.UserGroup == types.UserGroupAdmin
}

// --------------------------------------
//   Response handling

func writeResponse(w http.ResponseWriter, r *http.Request, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if response == nil {
		w.WriteHeader(statusCode)
		return
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		logApiError(r, fmt.Errorf("error encoding json data: %w", err), 0,
			log.Fields{
				"data": fmt.Sprintf("%+v", response),
			})
		w.WriteHeader(http.StatusInternalServerError)
		response = types.ApiErrorResponse{
			Error: "error encoding json data",
		}
		if err = json.NewEncoder(w).Encode(response); err != nil {
			// there seems to be an error with the lib
			logApiError(r, fmt.Errorf("error encoding error response after failed encoding: %w", err), 0)
		}
		return
	}
	w.WriteHeader(statusCode)
	if _, err = w.Write(jsonData); err != nil {
		// already returned wrong status code to user, can't prevent that
		logApiError(r, fmt.Errorf("error writing response data: %w", err), 0)
	}
}

func returnError(w http.ResponseWriter, r *http.Request, code int, err error) {
	response := types.ApiErrorResponse{
		Error: err.Error(),
	}
	writeResponse(w, r, code, response)
}

func returnOk(w http.ResponseWriter, r *http.Request, data interface{}) {
	writeResponse(w, r, http.StatusOK, data)
}

func returnCreated(w http.ResponseWriter, r *http.Request, data interface{}) {
	writeResponse(w, r, http.StatusCreated, data)
}

func returnNoContent(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, r, http.StatusNoContent, nil)
}

// Errors

func returnBadRequest(w http.ResponseWriter, r *http.Request, err error) {
	returnError(w, r, http.StatusBadRequest, err)
}

func returnUnauthorized(w http.ResponseWriter, r *http.Request, err error) {
	returnError(w, r, http.StatusUnauthorized, err)
}

func returnNotFound(w http.ResponseWriter, r *http.Request, err error) {
	returnError(w, r, http.StatusNotFound, err)
}

func returnConflict(w http.ResponseWriter, r *http.Request, err error) {
	returnError(w, r, http.StatusConflict, err)
}

func returnForbidden(w http.ResponseWriter, r *http.Request, err error) {
	returnError(w, r, http.StatusForbidden, err)
}

func returnTooManyRequests(w http.ResponseWriter, r *http.Request, err error) {
	returnError(w, r, http.StatusTooManyRequests, err)
}

func returnGone(w http.ResponseWriter, r *http.Request, err error) {
	returnError(w, r, http.StatusGone, err)
}

const maxBodySize = 10 * 1024

func logApiError(r *http.Request, err error, callerSkip int, additionalInfos ...log.Fields) {
	requestFields := log.Fields{
		"request_endpoint": r.Method + " " + r.URL.Path,
	}
	if len(r.URL.RawQuery) > 0 {
		requestFields["request_query"] = r.URL.RawQuery
	}
	if body, _ := io.ReadAll(io.LimitReader(r.Body, maxBodySize)); len(body) > 0 {
		requestFields["request_body"] = string(body)
	}
	if userId, _ := GetUserIdByContext(r); userId != 0 {
		requestFields["request_user_id"] = userId
	}
	log.Error(err, "error handling request", callerSkip+1, append(additionalInfos, requestFields)...)
}

func handleErr(w http.ResponseWriter, r *http.Request, err error) {
	_, isValidationError := err.(validationError)
	switch {
	case isValidationError, errors.Is(err, errBadRequest):
		returnBadRequest(w, r, err)
	case errors.Is(err, dataaccess.ErrNotFound):
		returnNotFound(w, r, err)
	case errors.Is(err, errUnauthorized):
		returnUnauthorized(w, r, err)
	case errors.Is(err, errForbidden):
		returnForbidden(w, r, err)
	case errors.Is(err, errConflict):
		returnConflict(w, r, err)
	case errors.Is(err, services.ErrWaiting):
		returnError(w, r, http.StatusServiceUnavailable, err)
	case errors.Is(err, errTooManyRequests):
		returnTooManyRequests(w, r, err)
	case errors.Is(err, errGone):
		returnGone(w, r, err)
	case errors.Is(err, context.Canceled):
		if r.Context().Err() != context.Canceled { // only return error if the request context was canceled
			logApiError(r, err, 1)
			returnError(w, r, http.StatusInternalServerError, err)
		}
	default:
		logApiError(r, err, 1)
		// TODO: don't return the error message to the user in production
		returnError(w, r, http.StatusInternalServerError, err)
	}
}

// --------------------------------------
//  Error Helpers

func errWithMsg(err error, format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", err, fmt.Sprintf(format, args...))
}

//nolint:nolintlint
//nolint:unparam
func newBadRequestErr(format string, args ...interface{}) error {
	return errWithMsg(errBadRequest, format, args...)
}

//nolint:unparam
func newInternalServerErr(format string, args ...interface{}) error {
	return errWithMsg(errInternalServer, format, args...)
}

//nolint:unparam
func newUnauthorizedErr(format string, args ...interface{}) error {
	return errWithMsg(errUnauthorized, format, args...)
}

func newForbiddenErr(format string, args ...interface{}) error {
	return errWithMsg(errForbidden, format, args...)
}

//nolint:unparam
func newConflictErr(format string, args ...interface{}) error {
	return errWithMsg(errConflict, format, args...)
}

//nolint:nolintlint
//nolint:unparam
func newNotFoundErr(format string, args ...interface{}) error {
	return errWithMsg(dataaccess.ErrNotFound, format, args...)
}

func newTooManyRequestsErr(format string, args ...interface{}) error {
	return errWithMsg(errTooManyRequests, format, args...)
}

func newGoneErr(format string, args ...interface{}) error {
	return errWithMsg(errGone, format, args...)
}

// --------------------------------------
// misc. helper functions

// maps different types of validator dashboard summary validators to a common format
func mapVDBIndices(indices interface{}) ([]types.VDBSummaryValidatorsData, error) {
	if indices == nil {
		return nil, errors.New("no data found when mapping")
	}

	switch v := indices.(type) {
	case *types.VDBGeneralSummaryValidators:
		// deposited, online, offline, slashing, slashed, exited, withdrawn, pending, exiting, withdrawing
		return []types.VDBSummaryValidatorsData{
			mapUintSlice("deposited", v.Deposited),
			mapUintSlice("online", v.Online),
			mapUintSlice("offline", v.Offline),
			mapUintSlice("slashing", v.Slashing),
			mapUintSlice("slashed", v.Slashed),
			mapUintSlice("exited", v.Exited),
			mapUintSlice("withdrawn", v.Withdrawn),
			mapIndexTimestampSlice("pending", v.Pending),
			mapIndexTimestampSlice("exiting", v.Exiting),
			mapIndexTimestampSlice("withdrawing", v.Withdrawing),
		}, nil

	case *types.VDBSyncSummaryValidators:
		return []types.VDBSummaryValidatorsData{
			mapUintSlice("sync_current", v.Current),
			mapUintSlice("sync_upcoming", v.Upcoming),
			mapSlice("sync_past", v.Past,
				func(v types.VDBValidatorSyncPast) (uint64, []uint64) { return v.Index, []uint64{v.Count} },
			),
		}, nil

	case *types.VDBSlashingsSummaryValidators:
		return []types.VDBSummaryValidatorsData{
			mapSlice("got_slashed", v.GotSlashed,
				func(v types.VDBValidatorGotSlashed) (uint64, []uint64) { return v.Index, []uint64{v.SlashedBy} },
			),
			mapSlice("has_slashed", v.HasSlashed,
				func(v types.VDBValidatorHasSlashed) (uint64, []uint64) { return v.Index, v.SlashedIndices },
			),
		}, nil

	case *types.VDBProposalSummaryValidators:
		return []types.VDBSummaryValidatorsData{
			mapIndexBlocksSlice("proposal_proposed", v.Proposed),
			mapIndexBlocksSlice("proposal_missed", v.Missed),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported indices type")
	}
}

// maps different types of validator dashboard summary validators to a common format
func mapSlice[T any](category string, validators []T, getIndexAndDutyObjects func(validator T) (index uint64, dutyObjects []uint64)) types.VDBSummaryValidatorsData {
	validatorsData := make([]types.VDBSummaryValidator, len(validators))
	for i, validator := range validators {
		index, dutyObjects := getIndexAndDutyObjects(validator)
		validatorsData[i] = types.VDBSummaryValidator{Index: index, DutyObjects: dutyObjects}
	}
	return types.VDBSummaryValidatorsData{
		Category:   category,
		Validators: validatorsData,
	}
}
func mapUintSlice(category string, validators []uint64) types.VDBSummaryValidatorsData {
	return mapSlice(category, validators,
		func(v uint64) (uint64, []uint64) { return v, nil },
	)
}

func mapIndexTimestampSlice(category string, validators []types.IndexTimestamp) types.VDBSummaryValidatorsData {
	return mapSlice(category, validators,
		func(v types.IndexTimestamp) (uint64, []uint64) { return v.Index, []uint64{v.Timestamp} },
	)
}

func mapIndexBlocksSlice(category string, validators []types.IndexBlocks) types.VDBSummaryValidatorsData {
	return mapSlice(category, validators,
		func(v types.IndexBlocks) (uint64, []uint64) { return v.Index, v.Blocks },
	)
}

// --------------------------------------
// intOrString is a custom type that can be unmarshalled from either an int or a string (strings will also be parsed to int if possible).
// if unmarshaling throws no errors one of the two fields will be set, the other will be nil.
type intOrString struct {
	intValue *uint64
	strValue *string
}

func (v *intOrString) UnmarshalJSON(data []byte) error {
	// Attempt to unmarshal as uint64 first
	var intValue uint64
	if err := json.Unmarshal(data, &intValue); err == nil {
		v.intValue = &intValue
		return nil
	}

	// If unmarshalling as uint64 fails, try to unmarshal as string
	var strValue string
	if err := json.Unmarshal(data, &strValue); err == nil {
		if parsedInt, err := strconv.ParseUint(strValue, 10, 64); err == nil {
			v.intValue = &parsedInt
		} else {
			v.strValue = &strValue
		}
		return nil
	}

	// If both unmarshalling attempts fail, return an error
	return fmt.Errorf("failed to unmarshal intOrString from json: %s", string(data))
}

func (v intOrString) String() string {
	if v.intValue != nil {
		return strconv.FormatUint(*v.intValue, 10)
	}
	if v.strValue != nil {
		return *v.strValue
	}
	return ""
}

func (intOrString) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{Type: "string"}, {Type: "integer"},
		},
	}
}

func isMocked(r *http.Request) bool {
	isMocked, ok := r.Context().Value(ctxIsMockedKey).(bool)
	return ok && isMocked
}
