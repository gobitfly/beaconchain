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

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/invopop/jsonschema"
	"github.com/xeipuuv/gojsonschema"

	"github.com/alexedwards/scs/v2"
	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/enums"
	types "github.com/gobitfly/beaconchain/pkg/api/types"
)

type HandlerService struct {
	dai dataaccess.DataAccessor
	scs *scs.SessionManager
}

func NewHandlerService(dataAccessor dataaccess.DataAccessor, sessionManager *scs.SessionManager) *HandlerService {
	return &HandlerService{
		dai: dataAccessor,
		scs: sessionManager,
	}
}

// --------------------------------------

var (
	// Subject to change, just examples
	reName                       = regexp.MustCompile(`^[a-zA-Z0-9_\-.\ ]+$`)
	reNumber                     = regexp.MustCompile(`^[0-9]+$`)
	reValidatorDashboardPublicId = regexp.MustCompile(`^v-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	//reAccountDashboardPublicId   = regexp.MustCompile(`^a-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	reValidatorPubkey = regexp.MustCompile(`^0x[0-9a-fA-F]{96}$`)
	reCursor          = regexp.MustCompile(`^[A-Za-z0-9-_]+$`) // has to be base64
	reEmail           = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
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
)

type Paging struct {
	cursor string
	limit  uint64
	search string
}

// All changes to common functions MUST NOT break any public handler behavior (not in effect yet)

// --------------------------------------
//   Validation

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
	return sb.String()
}

func (v validationError) add(paramName, problem string) {
	v[paramName] = problem
}

func (v validationError) hasErrors() bool {
	return len(v) > 0
}

func (v validationError) checkRegex(regex *regexp.Regexp, param, paramName string) string {
	if !regex.MatchString(param) {
		v.add(paramName, fmt.Sprintf(`given value '%s' has incorrect format`, param))
	}
	return param
}

func (v validationError) checkName(name string, minLength int) string {
	if len(name) < minLength {
		v.add("name", fmt.Sprintf(`given value '%s' is too short, minimum length is %d`, name, minLength))
		return name
	} else if len(name) > maxNameLength {
		v.add("name", fmt.Sprintf(`given value '%s' is too long, maximum length is %d`, name, maxNameLength))
		return name
	}
	return v.checkRegex(reName, name, "name")
}

func (v validationError) checkNameNotEmpty(name string) string {
	return v.checkName(name, 1)
}

func (v validationError) checkEmail(email string) string {
	return v.checkRegex(reEmail, email, "email")
}

// check request structure (body contains valid json and all required parameters are present)
// return error only if internal error occurs, otherwise join error to handlerErr and/or return nil
func (v validationError) checkBody(data interface{}, r *http.Request) error {
	// check if content type is application/json
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		v.add("request body", "'Content-Type' header must be 'application/json'")
	}
	body := r.Body

	// Read the entire request body (this consumes the request body)
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		log.Error(err, "error reading request body", 0, nil)
		return errors.New("can't read request body")
	}

	// Use bytes.NewReader to create an io.Reader for the body bytes, so it can be reused
	bodyReader := bytes.NewReader(bodyBytes)

	// First check: Decode into an empty interface to check JSON format
	var i interface{}
	if err := json.NewDecoder(bodyReader).Decode(&i); err != nil {
		v.add("request body", "not in JSON format")
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
		v.add("request body", "invalid schema, check the API documentation for the expected format")
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

func (v validationError) checkInt(param, paramName string) int64 {
	num, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		v.add(paramName, fmt.Sprintf("given value '%s' is not an integer", param))
	}
	return num
}

func (v validationError) checkUint(param, paramName string) uint64 {
	num, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		v.add(paramName, fmt.Sprintf("given value %s is not a positive integer", param))
	}
	return num
}

type validatorSet struct {
	Indexes    []uint64
	PublicKeys []string
}

// parseDashboardId is a helper function to validate the string dashboard id param.
func parseDashboardId(id string) (interface{}, error) {
	var v validationError
	if reNumber.MatchString(id) {
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
		return nil, newBadRequestErr("given value %s is not a valid dashboard id", id)
	}
	indexes, publicKeys := v.checkValidatorList(string(decodedId))
	if v.hasErrors() {
		return nil, v
	}
	return validatorSet{Indexes: indexes, PublicKeys: publicKeys}, nil
}

// getDashboardId is a helper function to convert the dashboard id param to a VDBId.
// precondition: dashboardIdParam must be a valid dashboard id and either a primary id, public id, or list of validators.
func (h *HandlerService) getDashboardId(dashboardIdParam interface{}) (*types.VDBId, error) {
	switch dashboardId := dashboardIdParam.(type) {
	case types.VDBIdPrimary:
		return &types.VDBId{Id: dashboardId, Validators: nil}, nil
	case types.VDBIdPublic:
		dashboardInfo, err := h.dai.GetValidatorDashboardInfoByPublicId(dashboardId)
		if err != nil {
			return nil, err
		}
		return &types.VDBId{Id: dashboardInfo.Id, Validators: nil}, nil
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
func (h *HandlerService) handleDashboardId(param string) (*types.VDBId, error) {
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

func (v validationError) checkPrimaryDashboardId(param string) types.VDBIdPrimary {
	return types.VDBIdPrimary(v.checkUint(param, "dashboard_id"))
}

// checkGroupId validates the given group id and returns it as an int64.
// If the given group id is empty and allowEmpty is true, it returns -1 (all groups).
func (v validationError) checkGroupId(param string, allowEmpty bool) int64 {
	if param == "" && allowEmpty {
		return types.AllGroups
	}
	return v.checkInt(param, "group_id")
}

// checkExistingGroupId validates if the given group id is not empty and a positive integer.
func (v validationError) checkExistingGroupId(param string) int64 {
	id := v.checkGroupId(param, forbidEmpty)
	if id < 0 {
		v.add("group_id", fmt.Sprintf("given value '%s' is not a valid group id", param))
	}
	return id
}

func (v validationError) checkValidatorDashboardPublicId(publicId string) types.VDBIdPublic {
	return types.VDBIdPublic(v.checkRegex(reValidatorDashboardPublicId, publicId, "public_dashboard_id"))
}

func (v validationError) checkPagingParams(q url.Values) Paging {
	paging := Paging{
		cursor: q.Get("cursor"),
		limit:  defaultReturnLimit,
		search: q.Get("search"),
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.ParseUint(limitStr, 10, 64)
		if err != nil {
			v.add("limit", fmt.Sprintf("given value '%s' is not a valid positive integer", limitStr))
			return paging
		}
		if limit > maxQueryLimit {
			v.add("limit", fmt.Sprintf("given value '%d' is too large, maximum limit is %d", limit, maxQueryLimit))
			return paging
		}
		paging.limit = limit
	}

	if paging.cursor != "" {
		paging.cursor = v.checkRegex(reCursor, paging.cursor, "cursor")
	}

	return paging
}

func checkEnum[T enums.EnumFactory[T]](v validationError, enum string, name string) T {
	var c T
	col := c.NewFromString(enum)
	if col.Int() == -1 {
		v.add(name, fmt.Sprintf("given value '%s' for parameter '%s' is not valid", enum, name))
		return c
	}
	return col
}

func (v validationError) parseSortOrder(order string) bool {
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

func checkSort[T enums.EnumFactory[T]](v validationError, sortString string) *types.Sort[T] {
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

func (v validationError) checkValidatorList(validators string) ([]uint64, []string) {
	return v.checkValidatorArray(strings.Split(validators, ","))
}

func (v validationError) checkValidatorArray(validators []string) ([]uint64, []string) {
	var indexes []uint64
	var publicKeys []string
	for _, validator := range validators {
		if reNumber.MatchString(validator) {
			indexes = append(indexes, v.checkUint(validator, "validators"))
		} else if reValidatorPubkey.MatchString(validator) {
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

func (v validationError) checkNetwork(network string) uint64 {
	// try parsing as uint64
	networkId, err := strconv.ParseUint(network, 10, 64)
	if err != nil {
		// TODO try to match string with network name
		v.add("network", fmt.Sprintf("given value '%s'is not a valid network id", network))
	}
	return networkId
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

/* func returnConflict(w http.ResponseWriter, err error) {
	returnError(w, http.StatusConflict, err)
} */

func returnInternalServerError(w http.ResponseWriter, err error) {
	log.Error(err, "internal server error", 0, nil)
	returnError(w, http.StatusInternalServerError, err)
}

func handleErr(w http.ResponseWriter, err error) {
	if _, ok := err.(validationError); ok || errors.Is(err, errBadRequest) {
		returnBadRequest(w, err)
		return
	} else if errors.Is(err, dataaccess.ErrNotFound) {
		returnNotFound(w, err)
		return
	} else if errors.Is(err, errUnauthorized) {
		returnUnauthorized(w, err)
		return
	}
	returnInternalServerError(w, err)
}

// --------------------------------------
//  Error Helpers

func errWithMsg(err error, format string, args ...interface{}) error {
	return fmt.Errorf("%w: %s", err, fmt.Sprintf(format, args...))
}

func newBadRequestErr(format string, args ...interface{}) error {
	return errWithMsg(errBadRequest, format, args...)
}

func newUnauthorizedErr(format string, args ...interface{}) error {
	return errWithMsg(errUnauthorized, format, args...)
}

func newNotFoundErr(format string, args ...interface{}) error {
	return errWithMsg(dataaccess.ErrNotFound, format, args...)
}
