package handlers

import (
	"bytes"
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

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/invopop/jsonschema"
	"github.com/xeipuuv/gojsonschema"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	types "github.com/gobitfly/beaconchain/pkg/api/types"
)

type HandlerService struct {
	dai dataaccess.DataAccessor
}

func NewHandlerService(DataAccessor dataaccess.DataAccessor) HandlerService {
	return HandlerService{dai: DataAccessor}
}

// --------------------------------------

var (
	// Subject to change, just examples
	reName                       = regexp.MustCompile(`^[a-zA-Z0-9_\-.\ ]+$`)
	reNumber                     = regexp.MustCompile(`^[0-9]+$`)
	reValidatorDashboardPublicId = regexp.MustCompile(`^v-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	//reAccountDashboardPublicId   = regexp.MustCompile(`^a-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	reValidatorPubkey = regexp.MustCompile(`^0x[0-9a-fA-F]{96}$`)
	reCursor          = regexp.MustCompile(`^[0-9a-fA-F]*$`)
)

const (
	maxNameLength       uint64 = 50
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

type Paging struct {
	cursor string
	limit  uint64
	search string
}

// All changes to common functions MUST NOT break any public handler behavior (not in effect yet)

// --------------------------------------
//   Validation

func joinErr(err *error, message string) {
	if len(message) > 0 {
		*err = errors.Join(*err, errors.New(message))
	}
}

func checkRegex(handlerErr *error, regex *regexp.Regexp, param, paramName string) string {
	if !regex.MatchString(param) {
		joinErr(handlerErr, fmt.Sprintf(`given value '%s' for parameter ' `+paramName+` has incorrect format`, param))
	}
	return param
}

func checkName(handlerErr *error, name string, minLength int) string {
	if len(name) < minLength {
		joinErr(handlerErr, fmt.Sprintf(`given value '%s' for parameter 'name' is too short, minimum length is %d`, name, minLength))
	} else if len(name) > 50 {
		joinErr(handlerErr, fmt.Sprintf(`given value '%s' for parameter 'name' is too long, maximum length is %d`, name, maxNameLength))
	}
	return checkRegex(handlerErr, reName, name, "name")
}

func checkMultipleRegex(handlerErr *error, regexes []*regexp.Regexp, params []string, paramName string) []string {
OUTER:
	for _, param := range params {
		for _, regex := range regexes {
			if regex.MatchString(param) {
				continue OUTER
			}
		}
		joinErr(handlerErr, fmt.Sprintf("given value '%s' for parameter '%s' has incorrect format", param, paramName))
	}
	return params
}

func checkNameNotEmpty(handlerErr *error, name string) string {
	return checkName(handlerErr, name, 1)
}

// check request structure (body contains valid json and all required parameters are present)
// return error only if internal error occurs, otherwise join error to handlerErr and/or return nil
func checkBody(handlerErr *error, data interface{}, r io.Reader) error {
	// Read the entire request body (this consumes the request body)
	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		log.Error(err, "error reading request body", 0, nil)
		return errors.New("can't read request body")
	}

	// Use bytes.NewReader to create an io.Reader for the body bytes, so it can be reused
	bodyReader := bytes.NewReader(bodyBytes)

	// First check: Decode into an empty interface to check JSON format
	var i interface{}
	if err := json.NewDecoder(bodyReader).Decode(&i); err != nil {
		joinErr(handlerErr, "request body is not in JSON format")
		return nil
	}

	// Reset the reader for the next use
	_, err = bodyReader.Seek(0, io.SeekStart)
	if err != nil {
		return errors.New("couldn't seek to start of the body")
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
		joinErr(handlerErr, "error reading request body due to invalid schema, check the API documentation for the expected format")
		return nil
	}

	// Decode into the target data structure
	// Reset the reader again for the final decode
	_, err = bodyReader.Seek(0, io.SeekStart)
	if err != nil {
		return errors.New("couldn't seek to start of the body")
	}
	if err := json.NewDecoder(bodyReader).Decode(data); err != nil {
		log.Error(err, "error decoding json into target structure", 0, nil)
		return errors.New("couldn't decode JSON request into target structure")
	}

	// Proceed with additional validation or processing as necessary
	return nil
}

func checkInt(handlerErr *error, param, paramName string) int64 {
	num, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		joinErr(handlerErr, fmt.Sprintf("given value '"+param+"' for parameter '"+paramName+"' is not an integer"))
	}
	return num
}

func checkUint(handlerErr *error, param, paramName string) uint64 {
	num, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		joinErr(handlerErr, fmt.Sprintf("given value '"+param+"' for parameter '"+paramName+"' is not a positive integer"))
	}
	return num
}

// parseDashboardId is a helper function to validate the string dashboard id param.
func parseDashboardId(id string) (interface{}, error) {
	var err error
	if reNumber.MatchString(id) {
		// given id is a normal id
		id := checkUint(&err, id, "dashboard_id")
		return types.VDBIdPrimary(id), err
	}
	if reValidatorDashboardPublicId.MatchString(id) {
		// given id is a public id
		return types.VDBIdPublic(id), nil
	}
	// given id must be an encoded set of validators
	decodedId, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		return nil, errors.New("invalid format for parameter 'dashboard_id'")
	}
	validators := checkValidatorList(&err, string(decodedId))
	if len(validators) > 20 {
		return nil, errors.New("too many validators in the list, maximum is 20")
	}
	return validators, nil
}

// getDashboardId is a helper function to convert the dashboard id param to a VDBId.
// precondition: dashboardIdParam must be a valid dashboard id and either a primary id, public id, or list of validators.
func (h HandlerService) getDashboardId(dashboardIdParam interface{}) (*types.VDBId, error) {
	switch dashboardId := dashboardIdParam.(type) {
	case types.VDBIdPrimary:
		return &types.VDBId{Id: dashboardId, Validators: nil}, nil
	case types.VDBIdPublic:
		dashboardInfo, err := h.dai.GetValidatorDashboardInfoByPublicId(dashboardId)
		if err != nil {
			return nil, err
		}
		return &types.VDBId{Id: dashboardInfo.Id, Validators: nil}, nil
	case []string:
		validators, err := h.dai.GetValidatorsFromStrings(dashboardId)
		if err != nil {
			return nil, err
		}
		return &types.VDBId{Validators: validators}, nil
	}
	return nil, errors.New(errorMsgParsingId)
}

// handleDashboardId is a helper function to both validate the dashboard id param and convert it to a VDBId.
// it should be used as the last validation step for all internal dashboard handlers.
func (h HandlerService) handleDashboardId(param string) (*types.VDBId, error) {
	// validate dashboard id param
	dashboardIdParam, err := parseDashboardId(param)
	if err != nil {
		return nil, err
	}
	// convert to VDBId
	dashboardId, err := h.getDashboardId(dashboardIdParam)
	if err != nil {
		return nil, err
	}
	return dashboardId, nil
}

func checkDashboardPrimaryId(handlerErr *error, param string) types.VDBIdPrimary {
	return types.VDBIdPrimary(checkUint(handlerErr, param, "dashboard_id"))
}

// checkGroupId validates the given group id and returns it as an int64.
// If the given group id is empty and allowEmpty is true, it returns -1 (all groups).
func checkGroupId(handlerErr *error, param string, allowEmpty bool) int64 {
	if param == "" && allowEmpty {
		return types.AllGroups
	}
	return checkInt(handlerErr, param, "group_id")
}

// checkExistingGroupId validates if the given group id is not empty and a positive integer.
func checkExistingGroupId(handlerErr *error, param string) int64 {
	id := checkGroupId(handlerErr, param, forbidEmpty)
	if id < 0 {
		joinErr(handlerErr, "given value '"+param+"' for parameter 'group_id' is not a valid group id")
	}
	return id
}

func checkValidatorDashboardPublicId(handlerErr *error, publicId string) string {
	return checkRegex(handlerErr, reValidatorDashboardPublicId, publicId, "public_dashboard_id")
}

func checkPagingParams(handlerErr *error, q url.Values) Paging {
	paging := Paging{
		cursor: q.Get("cursor"),
		limit:  defaultReturnLimit,
		search: q.Get("search"),
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.ParseUint(limitStr, 10, 64)
		joinErr(handlerErr, err.Error())
		paging.limit = min(limit, maxQueryLimit)
	}

	if paging.cursor != "" {
		paging.cursor = checkRegex(handlerErr, reCursor, paging.cursor, "cursor")
	}

	return paging
}

func parseSortColumn[T enums.EnumFactory[T]](column string) (T, error) {
	var c T
	col := c.NewFromString(column)
	if col.Int() == -1 {
		return col, errors.New("given value '" + column + "' for parameter 'sort' is not a valid column name for sorting")
	}
	return col, nil
}

func parseSortOrder(order string) (bool, error) {
	switch order {
	case "":
		return defaultSortOrder == sortOrderDescending, nil
	case sortOrderAscending:
		return false, nil
	case sortOrderDescending:
		return true, nil
	default:
		return false, errors.New("given value '" + order + "' for parameter 'sort' is not valid, allowed order values are: " + sortOrderAscending + ", " + sortOrderDescending + "")
	}
}

func checkSort[T enums.EnumFactory[T]](handlerErr *error, sort string) []types.Sort[T] {
	if sort == "" {
		var c T
		return []types.Sort[T]{{Column: c, Desc: false}}
	}
	sortQueries := strings.Split(sort, ",")
	sorts := make([]types.Sort[T], 0, len(sortQueries))
	for _, v := range sortQueries {
		sortSplit := strings.Split(v, ":")
		if len(sortSplit) > 2 {
			joinErr(handlerErr, "given value '"+v+"' for parameter 'sort' is not valid, expected format is '<column_name>[:(asc|desc)]'")
			return sorts
		}
		if len(sortSplit) == 1 {
			sortSplit = append(sortSplit, "")
		}
		sort, err := parseSortColumn[T](sortSplit[0])
		if err != nil {
			joinErr(handlerErr, err.Error())
			return sorts
		}
		order, err := parseSortOrder(sortSplit[1])
		if err != nil {
			joinErr(handlerErr, err.Error())
		}
		sorts = append(sorts, types.Sort[T]{Column: sort, Desc: order})
	}
	return sorts
}

func checkValidatorList(handlerErr *error, validators string) []string {
	return checkValidatorArray(handlerErr, strings.Split(validators, ","))
}

func checkValidatorArray(handlerErr *error, validators []string) []string {
	return checkMultipleRegex(handlerErr, []*regexp.Regexp{reNumber, reValidatorPubkey}, validators, "validators")
}

func checkNetwork(handlerErr *error, network string) uint64 {
	// try parsing as uint64
	networkId, err := strconv.ParseUint(network, 10, 64)
	if err != nil {
		// TODO string try to match network name
		joinErr(handlerErr, fmt.Sprintf("invalid network id: %s", network))
	}
	return networkId
}

// --------------------------------------
// Authorization

type User struct {
	Id uint64
	// TODO add more user fields, e.g. subscription tier
}

func getUser(r *http.Request) (User, error) {
	// TODO @LuccaBitfly add real user auth
	userId := r.Header.Get("X-User-Id")
	if userId == "" {
		return User{}, errors.New("missing user id, please set the X-User-Id header")
	}
	id, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		return User{}, errors.New("invalid user id, must be a positive integer")
	}
	// TODO if api key is used, fetch user id from the database

	// TODO if access token is used, verify the token and get user id from the token

	return User{
		Id: id,
	}, nil
}

// --------------------------------------
//   Response handling

func writeResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	if response == nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
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

//nolint:unused
func returnForbidden(w http.ResponseWriter, err error) {
	returnError(w, http.StatusForbidden, err)
}

//nolint:unused
func returnNotFound(w http.ResponseWriter, err error) {
	returnError(w, http.StatusNotFound, err)
}

//nolint:unused
func returnConflict(w http.ResponseWriter, err error) {
	returnError(w, http.StatusConflict, err)
}

func returnInternalServerError(w http.ResponseWriter, err error) {
	returnError(w, http.StatusInternalServerError, err)
}

func handleError(w http.ResponseWriter, err error) {
	// TODO @recy21 define error types in data access package
	// TODO @LuccaBitfly handle specific data access errors
	returnInternalServerError(w, err)
}
