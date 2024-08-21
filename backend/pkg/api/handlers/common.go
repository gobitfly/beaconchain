package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gorilla/mux"
	"github.com/invopop/jsonschema"
	"github.com/xeipuuv/gojsonschema"

	"github.com/alexedwards/scs/v2"
	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/services"
	types "github.com/gobitfly/beaconchain/pkg/api/types"
)

type HandlerService struct {
	dai dataaccess.DataAccessor
	scs *scs.SessionManager
}

func NewHandlerService(dataAccessor dataaccess.DataAccessor, sessionManager *scs.SessionManager) *HandlerService {
	if allNetworks == nil {
		networks, err := dataAccessor.GetAllNetworks()
		if err != nil {
			log.Fatal(err, "error getting networks for handler", 0, nil)
		}
		allNetworks = networks
	}

	return &HandlerService{
		dai: dataAccessor,
		scs: sessionManager,
	}
}

// all networks available in the system, filled on startup in NewHandlerService
var allNetworks []types.NetworkInfo

// --------------------------------------

var (
	// Subject to change, just examples
	reName                         = regexp.MustCompile(`^[a-zA-Z0-9_\-.\ ]*$`)
	reInteger                      = regexp.MustCompile(`^[0-9]+$`)
	reValidatorDashboardPublicId   = regexp.MustCompile(`^v-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	reValidatorPublicKeyWithPrefix = regexp.MustCompile(`^0x[0-9a-fA-F]{96}$`)
	reValidatorPublicKey           = regexp.MustCompile(`^(0x)?[0-9a-fA-F]{96}$`)
	reEthereumAddress              = regexp.MustCompile(`^(0x)?[0-9a-fA-F]{40}$`)
	reWithdrawalCredential         = regexp.MustCompile(`^(0x0[01])?[0-9a-fA-F]{62}$`)
	reEnsName                      = regexp.MustCompile(`^.+\.eth$`)
	reNonEmpty                     = regexp.MustCompile(`^\s*\S.*$`)
	reCursor                       = regexp.MustCompile(`^[A-Za-z0-9-_]+$`) // has to be base64
	reEmail                        = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	rePassword                     = regexp.MustCompile(`^.{5,}$`)
	reEmailUserToken               = regexp.MustCompile(`^[a-z0-9]{40}$`)
)

const (
	maxNameLength                     = 50
	maxValidatorsInList               = 20
	maxQueryLimit              uint64 = 100
	defaultReturnLimit         uint64 = 10
	sortOrderAscending                = "asc"
	sortOrderDescending               = "desc"
	defaultSortOrder                  = sortOrderAscending
	ethereum                          = "ethereum"
	gnosis                            = "gnosis"
	allowEmpty                        = true
	forbidEmpty                       = false
	maxArchivedDashboardsCount        = 10
)

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

type Paging struct {
	cursor string
	limit  uint64
	search string
}

// All changes to common functions MUST NOT break any public handler behavior (not in effect yet)

// --------------------------------------
//   Input Validation

// validationError is a map of parameter names to error messages.
// It is used to collect multiple validation errors before returning them to the user.
type validationError map[string]string

func (v validationError) Error() string {
	//iterate over map and create a string
	var sb strings.Builder
	for k, v := range v {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(v)
		sb.WriteString("\n")
	}
	return sb.String()[:sb.Len()-1]
}

func (v *validationError) add(paramName, problem string) {
	if *v == nil {
		*v = make(validationError)
	}
	validationMap := *v
	if _, ok := validationMap[paramName]; ok {
		problem = validationMap[paramName] + "; " + problem
	}
	validationMap[paramName] = problem
}

func (v *validationError) hasErrors() bool {
	return v != nil && len(*v) > 0
}

func (v *validationError) checkRegex(regex *regexp.Regexp, param, paramName string) string {
	if !regex.MatchString(param) {
		v.add(paramName, fmt.Sprintf(`given value '%s' has incorrect format`, param))
	}
	return param
}

func (v *validationError) checkLength(name, paramName string, minLength int) string {
	if len(name) < minLength {
		v.add(paramName, fmt.Sprintf(`given value '%s' is too short, minimum length is %d`, name, minLength))
	}
	if len(name) > maxNameLength {
		v.add(paramName, fmt.Sprintf(`given value '%s' is too long, maximum length is %d`, name, maxNameLength))
	}
	return name
}

func (v *validationError) checkName(name string, minLength int) string {
	name = v.checkLength(name, "name", minLength)
	return v.checkRegex(reName, name, "name")
}

func (v *validationError) checkNameNotEmpty(name string) string {
	return v.checkName(name, 1)
}

func (v *validationError) checkKeyNotEmpty(key string) string {
	key = v.checkLength(key, "key", 1)
	return v.checkRegex(reName, key, "key")
}

func (v *validationError) checkEmail(email string) string {
	return v.checkRegex(reEmail, strings.ToLower(email), "email")
}

func (v *validationError) checkPassword(password string) string {
	return v.checkRegex(rePassword, password, "password")
}

func (v *validationError) checkUserEmailToken(token string) string {
	return v.checkRegex(reEmailUserToken, token, "token")
}

// check request structure (body contains valid json and all required parameters are present)
// return error only if internal error occurs, otherwise add error to validationError and/or return nil
func (v *validationError) checkBody(data interface{}, r *http.Request) error {
	// check if content type is application/json
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		v.add("request body", "'Content-Type' header must be 'application/json'")
	}

	bodyBytes, err := io.ReadAll(r.Body)
	r.Body = io.NopCloser(io.LimitReader(bytes.NewReader(bodyBytes), 1024)) // unconsume body for error logging, but limit body size to 1KB
	if err != nil {
		return newInternalServerErr("error reading request body")
	}

	// First check: Unmarshal into an empty interface to check JSON format
	var i interface{}
	if err := json.Unmarshal(bodyBytes, &i); err != nil {
		v.add("request body", "not in JSON format")
		return nil
	}

	// Second check: Validate against the expected schema
	sc := jsonschema.Reflect(data)
	b, err := json.Marshal(sc)
	if err != nil {
		return newInternalServerErr("error creating expected schema")
	}
	loader := gojsonschema.NewBytesLoader(b)
	documentLoader := gojsonschema.NewBytesLoader(bodyBytes)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		return newInternalServerErr("error creating schema")
	}
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return newInternalServerErr("error validating schema")
	}
	isSchemaValid := result.Valid()
	if !isSchemaValid {
		v.add("request body", "invalid schema, check the API documentation for the expected format")
	}

	// Unmarshal into the target struct, only log error if it's a valid JSON
	if err := json.Unmarshal(bodyBytes, data); err != nil && isSchemaValid {
		return newInternalServerErr("error unmarshalling request body")
	}

	// Proceed with additional validation or processing as necessary
	return nil
}

func (v *validationError) checkInt(param, paramName string) int64 {
	num, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		v.add(paramName, fmt.Sprintf("given value '%s' is not an integer", param))
	}
	return num
}

func (v *validationError) checkUint(param, paramName string) uint64 {
	num, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		v.add(paramName, fmt.Sprintf("given value %s is not a positive integer", param))
	}
	return num
}

func (v *validationError) checkAdConfigurationKeys(keysString string) []string {
	if keysString == "" {
		return []string{}
	}
	var keys []string
	for _, key := range splitParameters(keysString, ',') {
		keys = append(keys, v.checkRegex(reName, key, "keys"))
	}
	return keys
}

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
		dashboardInfo, err := h.dai.GetValidatorDashboardPublicId(ctx, dashboardId)
		if err != nil {
			return nil, err
		}
		return &types.VDBId{Id: types.VDBIdPrimary(dashboardInfo.DashboardId), Validators: nil, AggregateGroups: !dashboardInfo.ShareSettings.ShareGroups}, nil
	case validatorSet:
		validators, err := h.dai.GetValidatorsFromSlices(dashboardId.Indexes, dashboardId.PublicKeys)
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
	limits.LatestExportedTs, err = h.dai.GetLatestExportedChartTs(ctx, aggregation)
	if err != nil {
		return limits, err
	}
	limits.MinAllowedTs = limits.LatestExportedTs - min(maxAge, limits.LatestExportedTs)                        // min to prevent underflow
	secondsPerEpoch := uint64(12 * 32)                                                                          // TODO: fetch dashboards chain id and use correct value for network once available
	limits.MaxAllowedInterval = chartDatapointLimit*uint64(aggregation.Duration(secondsPerEpoch).Seconds()) - 1 // -1 to make sure we don't go over the limit

	return limits, nil
}

func (v *validationError) checkPrimaryDashboardId(param string) types.VDBIdPrimary {
	return types.VDBIdPrimary(v.checkUint(param, "dashboard_id"))
}

// getDashboardPremiumPerks gets the premium perks of the dashboard OWNER or if it's a guest dashboard, it returns free tier premium perks
func (h *HandlerService) getDashboardPremiumPerks(ctx context.Context, id types.VDBId) (*types.PremiumPerks, error) {
	// for guest dashboards, return free tier perks
	if id.Validators != nil {
		perk, err := h.dai.GetFreeTierPerks(ctx)
		if err != nil {
			return nil, err
		}
		return perk, nil
	}
	// could be made into a single query if needed
	dashboardUser, err := h.dai.GetValidatorDashboardUser(ctx, id.Id)
	if err != nil {
		return nil, err
	}
	userInfo, err := h.dai.GetUserInfo(ctx, dashboardUser.UserId)
	if err != nil {
		return nil, err
	}

	return &userInfo.PremiumPerks, nil
}

// helper function to unify handling of block detail request validation
func (h *HandlerService) validateBlockRequest(r *http.Request, paramName string) (uint64, uint64, error) {
	var v validationError
	var err error
	chainId := v.checkNetworkParameter(mux.Vars(r)["network"])
	var value uint64
	switch paramValue := mux.Vars(r)[paramName]; paramValue {
	// possibly add other values like "genesis", "finalized", hardforks etc. later
	case "latest":
		if paramName == "block" {
			value, err = h.dai.GetLatestBlock()
		} else if paramName == "slot" {
			value, err = h.dai.GetLatestSlot()
		}
		if err != nil {
			return 0, 0, err
		}
	default:
		value = v.checkUint(paramValue, paramName)
	}
	if v.hasErrors() {
		return 0, 0, v
	}
	return chainId, value, nil
}

// checkGroupId validates the given group id and returns it as an int64.
// If the given group id is empty and allowEmpty is true, it returns -1 (all groups).
func (v *validationError) checkGroupId(param string, allowEmpty bool) int64 {
	if param == "" && allowEmpty {
		return types.AllGroups
	}
	return v.checkInt(param, "group_id")
}

// checkExistingGroupId validates if the given group id is not empty and a positive integer.
func (v *validationError) checkExistingGroupId(param string) uint64 {
	return v.checkUint(param, "group_id")
}

//nolint:unparam
func splitParameters(params string, delim rune) []string {
	// This splits the string by delim and removes empty strings
	f := func(c rune) bool {
		return c == delim
	}
	return strings.FieldsFunc(params, f)
}

func parseGroupIdList[T any](groupIds string, convert func(string, string) T) []T {
	var ids []T
	for _, id := range splitParameters(groupIds, ',') {
		ids = append(ids, convert(id, "group_ids"))
	}
	return ids
}

func (v *validationError) checkExistingGroupIdList(groupIds string) []uint64 {
	return parseGroupIdList(groupIds, v.checkUint)
}

func (v *validationError) checkGroupIdList(groupIds string) []int64 {
	return parseGroupIdList(groupIds, v.checkInt)
}

func (v *validationError) checkValidatorDashboardPublicId(publicId string) types.VDBIdPublic {
	return types.VDBIdPublic(v.checkRegex(reValidatorDashboardPublicId, publicId, "public_dashboard_id"))
}

type number interface {
	uint64 | int64 | float64
}

func checkMinMax[T number](v *validationError, param T, min T, max T, paramName string) T {
	if param < min {
		v.add(paramName, fmt.Sprintf("given value '%v' is too small, minimum value is %v", param, min))
	}
	if param > max {
		v.add(paramName, fmt.Sprintf("given value '%v' is too large, maximum value is %v", param, max))
	}
	return param
}

func (v *validationError) checkAddress(publicId string) string {
	return v.checkRegex(reEthereumAddress, publicId, "address")
}

func (v *validationError) checkUintMinMax(param string, min uint64, max uint64, paramName string) uint64 {
	return checkMinMax(v, v.checkUint(param, paramName), min, max, paramName)
}

func (v *validationError) checkPagingParams(q url.Values) Paging {
	paging := Paging{
		cursor: q.Get("cursor"),
		limit:  defaultReturnLimit,
		search: q.Get("search"),
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		paging.limit = v.checkUintMinMax(limitStr, 1, maxQueryLimit, "limit")
	}

	if paging.cursor != "" {
		paging.cursor = v.checkRegex(reCursor, paging.cursor, "cursor")
	}

	return paging
}

// checkEnum validates the given enum string and returns the corresponding enum value.
func checkEnum[T enums.EnumFactory[T]](v *validationError, enumString string, name string) T {
	var e T
	enum := e.NewFromString(enumString)
	if enums.IsInvalidEnum(enum) {
		v.add(name, fmt.Sprintf("given value '%s' is not valid", enumString))
		return enum
	}
	return enum
}

// checkEnumIsAllowed checks if the given enum is in the list of allowed enums.
// precondition: the enum is the same type as the allowed enums.
func (v *validationError) checkEnumIsAllowed(enum enums.Enum, allowed []enums.Enum, name string) {
	if enums.IsInvalidEnum(enum) {
		v.add(name, "parameter is missing or invalid, please check the API documentation")
		return
	}
	for _, a := range allowed {
		if enum.Int() == a.Int() {
			return
		}
	}
	v.add(name, "parameter is missing or invalid, please check the API documentation")
}

func (v *validationError) parseSortOrder(order string) bool {
	switch order {
	case "":
		return defaultSortOrder == sortOrderDescending
	case sortOrderAscending:
		return false
	case sortOrderDescending:
		return true
	default:
		v.add("sort", fmt.Sprintf("given value '%s' for parameter 'sort' is not valid, allowed order values are: %s, %s", order, sortOrderAscending, sortOrderDescending))
		return false
	}
}

func checkSort[T enums.EnumFactory[T]](v *validationError, sortString string) *types.Sort[T] {
	var c T
	if sortString == "" {
		return &types.Sort[T]{Column: c, Desc: false}
	}
	sortSplit := strings.Split(sortString, ":")
	if len(sortSplit) > 2 {
		v.add("sort", fmt.Sprintf("given value '%s' for parameter 'sort' is not valid, expected format is '<column_name>[:(asc|desc)]'", sortString))
		return nil
	}
	if len(sortSplit) == 1 {
		sortSplit = append(sortSplit, "")
	}
	sortCol := checkEnum[T](v, sortSplit[0], "sort")
	order := v.parseSortOrder(sortSplit[1])
	return &types.Sort[T]{Column: sortCol, Desc: order}
}

func (v *validationError) checkProtocolModes(protocolModes string) types.VDBProtocolModes {
	var modes types.VDBProtocolModes
	if protocolModes == "" {
		return modes
	}
	protocolsSlice := splitParameters(protocolModes, ',')
	for _, protocolMode := range protocolsSlice {
		switch protocolMode {
		case "rocket_pool":
			modes.RocketPool = true
		default:
			v.add("modes", fmt.Sprintf("given value '%s' is not a valid protocol mode", protocolMode))
		}
	}
	return modes
}

func (v *validationError) checkValidatorList(validators string, allowEmpty bool) ([]types.VDBValidator, []string) {
	if validators == "" && !allowEmpty {
		v.add("validators", "list of validators is must not be empty")
		return nil, nil
	}
	validatorsSlice := splitParameters(validators, ',')
	var indexes []types.VDBValidator
	var publicKeys []string
	for _, validator := range validatorsSlice {
		if reInteger.MatchString(validator) {
			indexes = append(indexes, v.checkUint(validator, "validators"))
		} else if reValidatorPublicKeyWithPrefix.MatchString(validator) {
			_, err := hexutil.Decode(validator)
			if err != nil {
				v.add("validators", fmt.Sprintf("invalid value '%s' in list of validators", v))
			}
			publicKeys = append(publicKeys, validator)
		} else {
			v.add("validators", fmt.Sprintf("invalid value '%s' in list of validators", v))
		}
	}
	return indexes, publicKeys
}

func (v *validationError) checkValidators(validators []intOrString, allowEmpty bool) ([]types.VDBValidator, []string) {
	if len(validators) == 0 && !allowEmpty {
		v.add("validators", "list of validators is empty")
		return nil, nil
	}
	var indexes []types.VDBValidator
	var publicKeys []string
	for _, validator := range validators {
		switch {
		case validator.intValue != nil:
			indexes = append(indexes, *validator.intValue)
		case validator.strValue != nil:
			if !reValidatorPublicKey.MatchString(*validator.strValue) {
				v.add("validators", fmt.Sprintf("given value '%s' is not a valid validator", *validator.strValue))
				continue
			}
			publicKeys = append(publicKeys, *validator.strValue)
		default:
			v.add("validators", "list contains invalid validator")
		}
	}
	return indexes, publicKeys
}

func (v *validationError) checkNetwork(network intOrString) uint64 {
	chainId, ok := isValidNetwork(network)
	if !ok {
		v.add("network", fmt.Sprintf("given value '%s' is not a valid network", network))
	}
	return chainId
}

func (v *validationError) checkNetworkParameter(param string) uint64 {
	if reInteger.MatchString(param) {
		chainId, err := strconv.ParseUint(param, 10, 64)
		if err != nil {
			v.add("network", fmt.Sprintf("given value '%s' is not a valid network", param))
			return 0
		}
		return v.checkNetwork(intOrString{intValue: &chainId})
	}
	return v.checkNetwork(intOrString{strValue: &param})
}

// isValidNetwork checks if the given network is a valid network.
// It returns the chain id of the network and true if it is valid, otherwise 0 and false.
func isValidNetwork(network intOrString) (uint64, bool) {
	for _, realNetwork := range allNetworks {
		if (network.intValue != nil && realNetwork.ChainId == *network.intValue) || (network.strValue != nil && realNetwork.Name == *network.strValue) {
			return realNetwork.ChainId, true
		}
	}
	return 0, false
}

func (v *validationError) checkTimestamps(r *http.Request, chartLimits ChartTimeDashboardLimits) (after uint64, before uint64) {
	afterParam := r.URL.Query().Get("after_ts")
	beforeParam := r.URL.Query().Get("before_ts")
	switch {
	// If both parameters are empty, return the latest data
	case afterParam == "" && beforeParam == "":
		return max(chartLimits.LatestExportedTs-chartLimits.MaxAllowedInterval, chartLimits.MinAllowedTs), chartLimits.LatestExportedTs

	// If only the afterParam is provided
	case afterParam != "" && beforeParam == "":
		afterTs := v.checkUint(afterParam, "after_ts")
		beforeTs := afterTs + chartLimits.MaxAllowedInterval
		return afterTs, beforeTs

	// If only the beforeParam is provided
	case beforeParam != "" && afterParam == "":
		beforeTs := v.checkUint(beforeParam, "before_ts")
		afterTs := max(beforeTs-chartLimits.MaxAllowedInterval, chartLimits.MinAllowedTs)
		return afterTs, beforeTs

	// If both parameters are provided, validate them
	default:
		afterTs := v.checkUint(afterParam, "after_ts")
		beforeTs := v.checkUint(beforeParam, "before_ts")

		if afterTs > beforeTs {
			v.add("after_ts", "parameter `after_ts` must not be greater than `before_ts`")
		}

		if beforeTs-afterTs > chartLimits.MaxAllowedInterval {
			v.add("before_ts", fmt.Sprintf("parameters `after_ts` and `before_ts` must not lie apart more than %d seconds for this aggregation", chartLimits.MaxAllowedInterval))
		}

		return afterTs, beforeTs
	}
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
		logApiError(r, fmt.Errorf("error encoding json data: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		response = types.ApiErrorResponse{
			Error: "error encoding json data",
		}
		if err = json.NewEncoder(w).Encode(response); err != nil {
			// there seems to be an error with the lib
			logApiError(r, fmt.Errorf("error encoding error response after failed encoding: %w", err))
		}
		return
	}
	w.WriteHeader(statusCode)
	if _, err = w.Write(jsonData); err != nil {
		// already returned wrong status code to user, can't prevent that
		logApiError(r, fmt.Errorf("error writing response data: %w", err))
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

func logApiError(r *http.Request, err error) {
	body, _ := io.ReadAll(r.Body)
	log.Error(err, "error handling request", 3, nil,
		map[string]interface{}{
			"endpoint": r.Method + " " + r.URL.Path,
			"query":    r.URL.RawQuery,
			"body":     string(body),
		},
	)
}

func returnInternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	logApiError(r, err)
	// TODO: don't return the error message to the user in production
	returnError(w, r, http.StatusInternalServerError, err)
}

func handleErr(w http.ResponseWriter, r *http.Request, err error) {
	_, isValidationError := err.(validationError)
	switch {
	case isValidationError || errors.Is(err, errBadRequest):
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
	default:
		returnInternalServerError(w, r, err)
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
