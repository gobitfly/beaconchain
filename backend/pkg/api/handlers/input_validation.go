package handlers

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	"github.com/gorilla/mux"
	"github.com/invopop/jsonschema"
	"github.com/shopspring/decimal"
	"github.com/xeipuuv/gojsonschema"
)

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
	reGraffiti                     = regexp.MustCompile(`^.{2,}$`)          // at least 2 characters, so that queries won't time out
	reCursor                       = regexp.MustCompile(`^[A-Za-z0-9-_]+$`) // has to be base64
	reEmail                        = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	rePassword                     = regexp.MustCompile(`^.{5,}$`)
	reEmailUserToken               = regexp.MustCompile(`^[a-z0-9]{40}$`)
	reJsonContentType              = regexp.MustCompile(`^application\/json(;.*)?$`)
)

const (
	maxNameLength                     = 50
	maxValidatorsInList               = 20
	maxQueryLimit              uint64 = 100
	defaultReturnLimit         uint64 = 10
	sortOrderAscending                = "asc"
	sortOrderDescending               = "desc"
	defaultSortOrder                  = sortOrderAscending
	defaultDesc                       = defaultSortOrder == sortOrderDescending
	ethereum                          = "ethereum"
	gnosis                            = "gnosis"
	allowEmpty                        = true
	forbidEmpty                       = false
	MaxArchivedDashboardsCount        = 10
)

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

// --------------------------------------

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
	if contentType := r.Header.Get("Content-Type"); !reJsonContentType.MatchString(contentType) {
		v.add("request body", "'Content-Type' header must be 'application/json'")
	}

	bodyBytes, err := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // unconsume body for error logging
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

func (v *validationError) checkWeiDecimal(param, paramName string) decimal.Decimal {
	dec := decimal.Zero
	// check if only numbers are contained in the string with regex
	if !reInteger.MatchString(param) {
		v.add(paramName, fmt.Sprintf("given value '%s' is not a wei string (must be positive integer)", param))
		return dec
	}
	dec, err := decimal.NewFromString(param)
	if err != nil {
		v.add(paramName, fmt.Sprintf("given value '%s' is not a wei string (must be positive integer)", param))
		return dec
	}
	return dec
}

func (v *validationError) checkWeiMinMax(param, paramName string, min, max decimal.Decimal) decimal.Decimal {
	dec := v.checkWeiDecimal(param, paramName)
	if v.hasErrors() {
		return dec
	}
	if dec.LessThan(min) {
		v.add(paramName, fmt.Sprintf("given value '%s' is too small, minimum value is %s", dec, min))
	}
	if dec.GreaterThan(max) {
		v.add(paramName, fmt.Sprintf("given value '%s' is too large, maximum value is %s", dec, max))
	}
	return dec
}

func (v *validationError) checkBool(param, paramName string) bool {
	if param == "" {
		return false
	}
	boolVar, err := strconv.ParseBool(param)
	if err != nil {
		v.add(paramName, fmt.Sprintf("given value '%s' is not a boolean", param))
	}
	return boolVar
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

func (v *validationError) checkPrimaryDashboardId(param string) types.VDBIdPrimary {
	return types.VDBIdPrimary(v.checkUint(param, "dashboard_id"))
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
			value, err = h.daService.GetLatestBlock()
		} else if paramName == "slot" {
			value, err = h.daService.GetLatestSlot()
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

func checkMinMax[T cmp.Ordered](v *validationError, param T, min T, max T, paramName string) T {
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

type Paging struct {
	cursor string
	limit  uint64
	search string
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

func (v *validationError) parseSortOrder(order string) bool {
	switch order {
	case "":
		return defaultDesc
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
		return &types.Sort[T]{Column: c, Desc: defaultDesc}
	}
	sortSplit := splitParameters(sortString, ':')
	if len(sortSplit) > 2 {
		v.add("sort", fmt.Sprintf("given value '%s' for parameter 'sort' is not valid, expected format is '<column_name>[:(asc|desc)]'", sortString))
		return nil
	}
	var desc bool
	if len(sortSplit) == 1 {
		desc = defaultDesc
	} else {
		desc = v.parseSortOrder(sortSplit[1])
	}
	sortCol := checkEnum[T](v, sortSplit[0], "sort")
	return &types.Sort[T]{Column: sortCol, Desc: desc}
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
		v.add("validators", "list of validators must not be empty")
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

func (v *validationError) checkNetworksParameter(param string) []uint64 {
	var chainIds []uint64
	for _, network := range splitParameters(param, ',') {
		chainIds = append(chainIds, v.checkNetworkParameter(network))
	}
	return chainIds
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
