package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"

	"github.com/alexedwards/scs/v2"
	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/services"
	types "github.com/gobitfly/beaconchain/pkg/api/types"
	commontypes "github.com/gobitfly/beaconchain/pkg/commons/types"
)

type HandlerService struct {
	daService dataaccess.DataAccessor
	daDummy   dataaccess.DataAccessor
	scs       *scs.SessionManager
	cfg       *commontypes.Config
}

func NewHandlerService(dataAccessor dataaccess.DataAccessor, dummy dataaccess.DataAccessor, sessionManager *scs.SessionManager, cfg *commontypes.Config) *HandlerService {
	if allNetworks == nil {
		networks, err := dataAccessor.GetAllNetworks()
		if err != nil {
			log.Fatal(err, "error getting networks for handler", 0, nil)
		}
		allNetworks = networks
	}

	return &HandlerService{
		daService: dataAccessor,
		daDummy:   dummy,
		scs:       sessionManager,
		cfg:       cfg,
	}
}

// getDataAccessor returns the correct data accessor based on the request context.
// if the request is mocked, the data access dummy is returned; otherwise the data access service.
// should only be used if getting mocked data for the endpoint is appropriate
func (h *HandlerService) getDataAccessor(ctx context.Context) dataaccess.DataAccessor {
	isMocked, isMockedOk := ctx.Value(types.CtxIsMockedKey).(bool)                         // set in StoreIsMockedFlagMiddleware
	isMockingAllowed, isMockingAllowedOk := ctx.Value(types.CtxIsMockingAllowedKey).(bool) // set in Handle function
	if isMockedOk && isMocked && isMockingAllowedOk && isMockingAllowed {
		return h.daDummy
	}
	return h.daService
}

// all networks available in the system, filled on startup in NewHandlerService
var allNetworks []types.NetworkInfo

type InputValidator[T any] interface {
	Validate(params map[string]string, payload io.ReadCloser) error
	*T
}

type BusinessLogicFunc[Input any, Response any] func(ctx context.Context, input Input) (Response, error)

func Handle[Input InputValidator[Value], Value, Response any](defaultCode int, logicFunc BusinessLogicFunc[Value, Response], isMockingAllowed bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// prepare input
		vars := mux.Vars(r)
		if vars == nil {
			vars = make(map[string]string)
		}
		q := r.URL.Query()
		for k, v := range q {
			if _, ok := vars[k]; ok || len(v) == 0 {
				continue
			}
			vars[k] = v[0]
		}
		// input validation
		var input Input = new(Value)
		err := input.Validate(vars, r.Body)
		if err != nil {
			handleErr(w, r, err)
			return
		}
		ctx := r.Context()
		if isMockingAllowed {
			ctx = context.WithValue(ctx, types.CtxIsMockingAllowedKey, true)
		}
		// business logic
		response, err := logicFunc(ctx, *input)
		if err != nil {
			handleErr(w, r, err)
			return
		}

		writeResponse(w, r, defaultCode, response)
	}
}

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
		validators, err := h.daService.GetValidatorsFromSlices(ctx, dashboardId.Indexes, dashboardId.PublicKeys)
		if err != nil {
			return nil, err
		}
		if len(validators) == 0 {
			return nil, newNotFoundErr("no validators found for given id")
		}
		validatorEb, err := h.daService.GetValidatorDashboardEffectiveBalanceTotal(ctx, types.VDBId{Validators: validators}, false)
		if err != nil {
			return nil, err
		}
		// TODO check if we also need a count limit because of cf url length limits
		perks, err := h.daService.GetFreeTierPerks(ctx)
		if err != nil {
			return nil, err
		}
		if utils.GWeiToWei(big.NewInt(int64(validatorEb))).GreaterThan(perks.EffectiveBalancePerDashboard) {
			return nil, newBadRequestErr("effective balance of validators in list is too high, maximum is %d", perks.EffectiveBalancePerDashboard.Div(decimal.NewFromInt(1e9)).IntPart())
		}
		return &types.VDBId{Validators: validators}, nil
	}
	return nil, errMsgParsingId
}

// handleDashboardId is a helper function to both validate the dashboard id param and convert it to a VDBId.
// it should be used as the last validation step for all internal dashboard GET-handlers.
// Modifying handlers (POST, PUT, DELETE) should only accept primary dashboard ids and just use checkPrimaryDashboardId.
func (h *HandlerService) handleDashboardId(ctx context.Context, param string) (*types.VDBId, error) {
	//check if dashboard id is stored in context
	if dashboardId, ok := ctx.Value(types.CtxDashboardIdKey).(*types.VDBId); ok {
		return dashboardId, nil
	}
	// validate dashboard id param
	var v validationError
	dashboardIdParam := v.checkDashboardId(param)
	if err := v.AsError(); err != nil {
		return nil, err
	}
	// convert to VDBId
	dashboardId, err := h.getDashboardId(ctx, dashboardIdParam)
	if err != nil {
		return nil, err
	}

	return dashboardId, nil
}

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

func isUserAdmin(user *types.UserInfo) bool {
	if user == nil { // can happen for guest or shared dashboards
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
	if userId, _ := GetUserIdByContext(r.Context()); userId != 0 {
		requestFields["request_user_id"] = userId
	}
	log.Error(err, "error handling request", callerSkip+1, append(additionalInfos, requestFields)...)
}

func handleErr(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errBadRequest):
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
			mapIndexSlotsSlice("proposal_proposed", v.Proposed),
			mapIndexSlotsSlice("proposal_missed", v.Missed),
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

func mapIndexSlotsSlice(category string, validators []types.IndexSlots) types.VDBSummaryValidatorsData {
	return mapSlice(category, validators,
		func(v types.IndexSlots) (uint64, []uint64) { return v.Index, v.Slots },
	)
}

// --------------------------------------
// notification event mapping

var dbEventToResponse = map[string]string{
	string(commontypes.ValidatorIsOfflineEventName):                "validator_offline",
	string(commontypes.ValidatorIsOnlineEventName):                 "validator_online",
	string(commontypes.ValidatorMissedAttestationEventName):        "attestation_missed",
	string(commontypes.ValidatorMissedProposalEventName):           "proposal_missed",
	string(commontypes.ValidatorExecutedProposalEventName):         "proposal_success",
	string(commontypes.ValidatorUpcomingProposalEventName):         "proposal_upcoming",
	string(commontypes.SyncCommitteeSoonEventName):                 "sync",
	string(commontypes.ValidatorReceivedWithdrawalEventName):       "withdrawal",
	string(commontypes.ValidatorGotSlashedEventName):               "validator_got_slashed",
	string(commontypes.ValidatorDidSlashEventName):                 "validator_has_slashed",
	string(commontypes.ValidatorGroupEfficiencyEventName):          "group_efficiency_below",
	string(commontypes.RocketpoolCollateralMinReachedEventName):    "min_collateral",
	string(commontypes.RocketpoolCollateralMaxReachedEventName):    "max_collateral",
	string(commontypes.IncomingTransactionEventName):               "incoming_tx",
	string(commontypes.OutgoingTransactionEventName):               "outgoing_tx",
	string(commontypes.ERC20TokenTransferEventName):                "transfer_erc20",
	string(commontypes.ERC721TokenTransferEventName):               "transfer_erc721",
	string(commontypes.ERC1155TokenTransferEventName):              "transfer_erc1155",
	string(commontypes.MonitoringMachineOfflineEventName):          "offline",
	string(commontypes.MonitoringMachineDiskAlmostFullEventName):   "storage",
	string(commontypes.MonitoringMachineCpuLoadEventName):          "cpu",
	string(commontypes.MonitoringMachineMemoryUsageEventName):      "memory",
	string(commontypes.RocketpoolNewClaimRoundStartedEventName):    "new_reward_round",
	string(commontypes.NetworkGasBelowThresholdEventName):          "gas_below",
	string(commontypes.NetworkGasAboveThresholdEventName):          "gas_above",
	string(commontypes.NetworkParticipationRateThresholdEventName): "participation_rate",
}

func mapNotificationEventName(event string) string {
	if name, ok := dbEventToResponse[event]; ok {
		return name
	}
	log.Warn("unknown notification event", log.Fields{"event": event})
	return event
}

func mapDashboardNotificationEvents(data []types.NotificationDashboardsTableRow) []types.NotificationDashboardsTableRow {
	for _, row := range data {
		for eventIndex := range row.EventTypes {
			row.EventTypes[eventIndex] = mapNotificationEventName(row.EventTypes[eventIndex])
		}
	}
	return data
}

func mapMachineNotificationEventNames(data []types.NotificationMachinesTableRow) []types.NotificationMachinesTableRow {
	for rowIndex, row := range data {
		data[rowIndex].EventType = mapNotificationEventName(row.EventType)
	}
	return data
}

func mapNetworkNotificationEventNames(data []types.NotificationNetworksTableRow) []types.NotificationNetworksTableRow {
	for rowIndex, row := range data {
		data[rowIndex].EventType = mapNotificationEventName(row.EventType)
	}
	return data
}
