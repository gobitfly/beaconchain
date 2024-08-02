package handlers

import (
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
	"time"

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
	reEmailConfirmationHash        = regexp.MustCompile(`^[a-z0-9]{40}$`)
)

const (
	maxNameLength              = 50
	maxValidatorsInList        = 20
	maxQueryLimit       uint64 = 100
	defaultReturnLimit  uint64 = 10
	sortOrderAscending         = "asc"
	sortOrderDescending        = "desc"
	defaultSortOrder           = sortOrderAscending
	ethereum                   = "ethereum"
	gnosis                     = "gnosis"
	allowEmpty                 = true
	forbidEmpty                = false
)

var (
	errMsgParsingId = errors.New("error parsing parameter 'dashboard_id'")
	errBadRequest   = errors.New("bad request")
	errUnauthorized = errors.New("unauthorized")
	errForbidden    = errors.New("forbidden")
	errConflict     = errors.New("conflict")
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

func (v *validationError) checkConfirmationHash(hash string) string {
	return v.checkRegex(reEmailConfirmationHash, hash, "token")
}

// check request structure (body contains valid json and all required parameters are present)
// return error only if internal error occurs, otherwise add error to validationError and/or return nil
func (v *validationError) checkBody(data interface{}, r *http.Request) error {
	// check if content type is application/json
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		v.add("request body", "'Content-Type' header must be 'application/json'")
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err, "error reading request body", 0, nil)
		return errors.New("can't read request body")
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
		log.Error(err, "error marshalling schema", 0, nil)
		return errors.New("can't marshal schema for validation")
	}
	loader := gojsonschema.NewBytesLoader(b)
	documentLoader := gojsonschema.NewBytesLoader(bodyBytes)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		log.Error(err, "error creating schema", 0, nil)
		return errors.New("can't create expected format")
	}
	result, err := schema.Validate(documentLoader)
	if err != nil {
		log.Error(err, "error validating json", 0, nil)
		return errors.New("couldn't validate JSON request")
	}
	if !result.Valid() {
		v.add("request body", "invalid schema, check the API documentation for the expected format")
		return nil
	}

	// Unmarshal into the target struct
	if err := json.Unmarshal(bodyBytes, data); err != nil {
		log.Error(err, "error decoding json into target structure", 0, nil)
		return errors.New("couldn't decode JSON request into target structure")
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
// it should be used as the last validation step for all internal dashboard handlers.
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
	dashboardInfo, err := h.dai.GetValidatorDashboardInfo(ctx, id.Id)
	if err != nil {
		return nil, err
	}
	userInfo, err := h.dai.GetUserInfo(ctx, dashboardInfo.UserId)
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
//
//nolint:unparam
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

func (v *validationError) checkDate(dateString string) time.Time {
	// expecting date in format "YYYY-MM-DD"
	date, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		v.add("date", fmt.Sprintf("given value '%s' is not a valid date", dateString))
	}
	return date
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

func (v *validationError) checkTimestamps(afterParam string, beforeParam string, latestExportedTs uint64, minAllowedTs uint64, maxAllowedInterval uint64) (after uint64, before uint64) {
	switch {
	// If both parameters are empty, return the latest data
	case afterParam == "" && beforeParam == "":
		return max(latestExportedTs-maxAllowedInterval, minAllowedTs), latestExportedTs

	// If only the afterParam is provided
	case afterParam != "" && beforeParam == "":
		afterTs := v.checkUint(afterParam, "after_ts")
		beforeTs := afterTs + maxAllowedInterval
		return afterTs, beforeTs

	// If only the beforeParam is provided
	case beforeParam != "" && afterParam == "":
		beforeTs := v.checkUint(beforeParam, "before_ts")
		afterTs := max(beforeTs-maxAllowedInterval, minAllowedTs)
		return afterTs, beforeTs

	// If both parameters are provided, validate them
	default:
		afterTs := v.checkUint(afterParam, "after_ts")
		beforeTs := v.checkUint(beforeParam, "before_ts")

		if afterTs > beforeTs {
			v.add("after_ts", "parameter `after_ts` must not be greater than `before_ts`")
		}

		if beforeTs-afterTs > maxAllowedInterval {
			v.add("before_ts", fmt.Sprintf("parameters `after_ts` and `before_ts` must not lie apart more than %d seconds for this aggregation", maxAllowedInterval))
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

// --------------------------------------
//   Response handling

func writeResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if response == nil {
		w.WriteHeader(statusCode)
		return
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Error(err, "error encoding json data", 2, nil)
		w.WriteHeader(http.StatusInternalServerError)
		response = types.ApiErrorResponse{
			Error: "error encoding json data",
		}
		if err = json.NewEncoder(w).Encode(response); err != nil {
			// there seems to be an error with the lib
			log.Error(err, "error writing response", 0, nil)
		}
		return
	}
	w.WriteHeader(statusCode)
	if _, err = w.Write(jsonData); err != nil {
		// already returned wrong status code to user, can't prevent that
		log.Error(err, "error writing response", 0, nil)
	}
}

func returnError(w http.ResponseWriter, code int, err error) {
	response := types.ApiErrorResponse{
		Error: err.Error(),
	}
	writeResponse(w, code, response)
}

func returnOk(w http.ResponseWriter, data interface{}) {
	writeResponse(w, http.StatusOK, data)
}

func returnCreated(w http.ResponseWriter, data interface{}) {
	writeResponse(w, http.StatusCreated, data)
}

func returnNoContent(w http.ResponseWriter) {
	writeResponse(w, http.StatusNoContent, nil)
}

// Errors

func returnBadRequest(w http.ResponseWriter, err error) {
	returnError(w, http.StatusBadRequest, err)
}

func returnUnauthorized(w http.ResponseWriter, err error) {
	returnError(w, http.StatusUnauthorized, err)
}

func returnNotFound(w http.ResponseWriter, err error) {
	returnError(w, http.StatusNotFound, err)
}

func returnConflict(w http.ResponseWriter, err error) {
	returnError(w, http.StatusConflict, err)
}

func returnForbidden(w http.ResponseWriter, err error) {
	returnError(w, http.StatusForbidden, err)
}

func returnInternalServerError(w http.ResponseWriter, err error) {
	log.Error(err, "internal server error", 2, nil)
	// TODO: don't return the error message to the user in production
	returnError(w, http.StatusInternalServerError, err)
}

func handleErr(w http.ResponseWriter, err error) {
	_, isValidationError := err.(validationError)
	switch {
	case isValidationError || errors.Is(err, errBadRequest):
		returnBadRequest(w, err)
	case errors.Is(err, dataaccess.ErrNotFound):
		returnNotFound(w, err)
	case errors.Is(err, errUnauthorized):
		returnUnauthorized(w, err)
	case errors.Is(err, errForbidden):
		returnForbidden(w, err)
	case errors.Is(err, errConflict):
		returnConflict(w, err)
	case errors.Is(err, services.ErrWaiting):
		returnError(w, http.StatusServiceUnavailable, err)
	default:
		returnInternalServerError(w, err)
	}
}

// --------------------------------------
//  Error Helpers

func errWithMsg(err error, format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", err, fmt.Sprintf(format, args...))
}

func newBadRequestErr(format string, args ...interface{}) error {
	return errWithMsg(errBadRequest, format, args...)
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

func newNotFoundErr(format string, args ...interface{}) error {
	return errWithMsg(dataaccess.ErrNotFound, format, args...)
}

// --------------------------------------
// misc. helper functions

// maps different types of validator dashboard summary validators to a common format
func mapVDBIndices(indices interface{}) ([]types.VDBSummaryValidatorsData, error) {
	if indices == nil {
		return nil, errors.New("no data found when mapping")
	}

	var data []types.VDBSummaryValidatorsData
	// Helper function to create a VDBValidatorIndices and append to data

	switch v := indices.(type) {
	case *types.VDBGeneralSummaryValidators:
		// deposited, online, offline, slashing, slashed, exited, withdrawn, pending, exiting, withdrawing
		data = append(data,
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
		)
		return data, nil

	case *types.VDBSyncSummaryValidators:
		data = append(data,
			mapUintSlice("sync_current", v.Current),
			mapUintSlice("sync_upcoming", v.Current),
		)
		pastValidators := make([]types.VDBSummaryValidator, len(v.Past))
		for i, validator := range v.Past {
			pastValidators[i] = types.VDBSummaryValidator{Index: validator.Index, DutyObjects: []uint64{validator.Count}}
		}
		data = append(data, types.VDBSummaryValidatorsData{
			Category:   "sync_past",
			Validators: pastValidators,
		})
		return data, nil

	case *types.VDBSlashingsSummaryValidators:
		return mapVDBSummarySlashings(v), nil

	case *types.VDBProposalSummaryValidators:
		return mapVDBSummaryProposals(v), nil

	default:
		return nil, fmt.Errorf("unsupported indices type")
	}
}
func mapUintSlice(category string, validators []uint64) types.VDBSummaryValidatorsData {
	validatorsData := make([]types.VDBSummaryValidator, len(validators))
	for i, validatorIndex := range validators {
		validatorsData[i] = types.VDBSummaryValidator{Index: validatorIndex}
	}
	return types.VDBSummaryValidatorsData{
		Category:   category,
		Validators: validatorsData,
	}
}

func mapIndexTimestampSlice(category string, validators []types.IndexTimestamp) types.VDBSummaryValidatorsData {
	validatorsData := make([]types.VDBSummaryValidator, len(validators))
	for i, validator := range validators {
		validatorsData[i] = types.VDBSummaryValidator{Index: validator.Index, DutyObjects: []uint64{validator.Timestamp}}
	}
	return types.VDBSummaryValidatorsData{
		Category:   category,
		Validators: validatorsData,
	}
}

func mapVDBSummarySlashings(v *types.VDBSlashingsSummaryValidators) []types.VDBSummaryValidatorsData {
	gotSlashedValidators := make([]types.VDBSummaryValidator, len(v.GotSlashed))
	for i, gotSlashed := range v.GotSlashed {
		gotSlashedValidators[i] = types.VDBSummaryValidator{Index: gotSlashed.Index, DutyObjects: []uint64{gotSlashed.SlashedBy}}
	}

	hasSlashedValidators := make([]types.VDBSummaryValidator, len(v.HasSlashed))
	for i, hasSlashed := range v.HasSlashed {
		hasSlashedValidators[i] = types.VDBSummaryValidator{Index: hasSlashed.Index, DutyObjects: hasSlashed.SlashedIndices}
	}

	return []types.VDBSummaryValidatorsData{
		{
			Category:   "got_slashed",
			Validators: gotSlashedValidators,
		},
		{
			Category:   "has_slashed",
			Validators: hasSlashedValidators,
		},
	}
}

func mapVDBSummaryProposals(v *types.VDBProposalSummaryValidators) []types.VDBSummaryValidatorsData {
	proposedValidators := make([]types.VDBSummaryValidator, len(v.Proposed))
	for i, proposed := range v.Proposed {
		proposedValidators[i] = types.VDBSummaryValidator{Index: proposed.Index, DutyObjects: proposed.Blocks}
	}

	missedValidators := make([]types.VDBSummaryValidator, len(v.Missed))
	for i, missed := range v.Missed {
		missedValidators[i] = types.VDBSummaryValidator{Index: missed.Index, DutyObjects: missed.Blocks}
	}

	return []types.VDBSummaryValidatorsData{
		{
			Category:   "proposal_proposed",
			Validators: proposedValidators,
		},
		{
			Category:   "proposal_missed",
			Validators: missedValidators,
		},
	}
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
