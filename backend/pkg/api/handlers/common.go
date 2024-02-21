package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/invopop/jsonschema"
	"github.com/xeipuuv/gojsonschema"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	types "github.com/gobitfly/beaconchain/pkg/api/types"
)

type HandlerService struct {
	dai dataaccess.DataAccessInterface
}

func NewHandlerService(dataAccessInterface dataaccess.DataAccessInterface) HandlerService {
	return HandlerService{dai: dataAccessInterface}
}

// --------------------------------------

type regexString string

const (
	// Subject to change, just examples
	reName            = regexString(`^[a-zA-Z0-9_\-.\ ]{` + regexString(rune(maxNameLength)) + `}$`)
	reId              = regexString(`^[a-zA-Z0-9_]+$`)
	reNumber          = regexString(`^[0-9]+$`)
	reValidatorPubkey = regexString(`^[0-9a-fA-F]{96}$`)
	reCursor          = regexString(`^[0-9a-fA-F]*$`)
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
)

type RequestError struct {
	StatusCode int

	Err error
}

func (r RequestError) Error() string {
	return fmt.Sprintf("status %d: err %v", r.StatusCode, r.Err)
}

type Paging struct {
	cursor string
	limit  uint64
	order  string
	sort   string
	search string
}

// All changes to common functions MUST NOT break any public handler behavior

// --------------------------------------
//   Validation

func joinErr(err *error, message string) {
	if len(message) > 0 {
		*err = errors.Join(*err, errors.New(message))
	}
}

func regexCheck(handlerErr *error, regex regexString, param string) string {
	if !regexp.MustCompile(string(regex)).MatchString(param) {
		joinErr(handlerErr, fmt.Sprintf(`given value '%s' has incorrect format`, param))
	}
	return param
}

func checkName(handlerErr *error, name string, minLength int) string {
	if len(name) < minLength {
		joinErr(handlerErr, fmt.Sprintf(`given value '%s' for parameter "name" is too short, minimum length is %d`, name, minLength))
	} else if len(name) > 50 {
		joinErr(handlerErr, fmt.Sprintf(`given value '%s' for parameter "name" is too long, maximum length is %d`, name, maxNameLength))
	}
	return regexCheck(handlerErr, reName, name)
}

func regexCheckMultiple(handlerErr *error, regexes []regexString, params []string) []string {
	results := make([]string, len(params))
	for i, param := range params {
		for _, regex := range regexes {
			regexCheck(handlerErr, regex, param)
		}
		// might want to change this later
		results[i] = params[i]
	}
	return results
}

func checkNameNotEmpty(handlerErr *error, name string) string {
	return checkName(handlerErr, name, 1)
}

// check request structure (body contains valid json and all required parameters are present)
func CheckAndGetJson(r io.Reader, data interface{}) error {
	sc := jsonschema.Reflect(data)
	var i interface{}
	if json.NewDecoder(r).Decode(&i) != nil {
		return RequestError{http.StatusBadRequest, errors.New("request is not in JSON format")}
	}
	b, err := json.Marshal(sc)
	if err != nil {
		utils.LogError(err, "error validating json", 0, nil)
		return RequestError{http.StatusInternalServerError, errors.New("can't validate expected format")}
	}
	loader := gojsonschema.NewBytesLoader(b)
	documentLoader, _ := gojsonschema.NewReaderLoader(r)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		utils.LogError(err, "error validating json", 0, nil)
		return RequestError{http.StatusInternalServerError, errors.New("can't create expected format")}
	}
	result, err := schema.Validate(documentLoader)
	if err != nil {
		utils.LogError(err, "error validating json", 0, nil)
		return RequestError{http.StatusInternalServerError, errors.New("couldn't validate JSON request")}
	}
	if !result.Valid() {
		return RequestError{http.StatusBadRequest, errors.New("unexpected JSON format. Check the API documentation for parameter details")}
	}
	if err = json.NewDecoder(r).Decode(data); err != nil {
		// error parsing json; shouldn't happen since we verified it's json in the correct format already
		utils.LogError(err, "error validating json", 0, nil)
		return RequestError{http.StatusInternalServerError, errors.New("couldn't decode JSON request")}
	}
	// could perform data validation checks based on tags here, but might need validation lib for that
	return nil
}

func checkId(handlerErr *error, id string) string {
	return regexCheck(handlerErr, reId, id)
}

func checkUint(handlerErr *error, id string) uint64 {
	id64, err := strconv.ParseUint(id, 10, 64)
	joinErr(handlerErr, err.Error())
	return id64
}

func CheckIdList(handlerErr *error, ids []string) []string {
	return regexCheckMultiple(handlerErr, []regexString{reId}, ids)
}

func checkAndGetPaging(handlerErr *error, r *http.Request) Paging {
	q := r.URL.Query()
	paging := Paging{
		cursor: q.Get("cursor"),
		limit:  defaultReturnLimit,
		order:  defaultSortOrder,
		sort:   q.Get("sort"),
		search: q.Get("search"),
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.ParseUint(limitStr, 10, 64)
		joinErr(handlerErr, err.Error())
		paging.limit = limit
		if limit > maxQueryLimit {
			joinErr(handlerErr, fmt.Sprintf("Paging limit %d is too high, maximum value is %d", paging.limit, maxQueryLimit))
		}
	}

	if order := q.Get("order"); order != "" {
		paging.order = order
	}
	if paging.order != sortOrderAscending && paging.order == sortOrderDescending {
		joinErr(handlerErr, fmt.Sprintf("invalid sorting order: %s", paging.order))
	}
	paging.cursor = regexCheck(handlerErr, reCursor, paging.cursor)
	paging.sort = checkName(handlerErr, paging.sort, 0)
	paging.search = checkName(handlerErr, paging.search, 0)

	return paging
}

func CheckValidatorList(handlerErr *error, validators []string) []string {
	return regexCheckMultiple(handlerErr, []regexString{reNumber, reValidatorPubkey}, validators)
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

func getUser(r *http.Request) (uint64, error) {
	// TODO @LuccaBitfly add real user auth
	userId := r.Header.Get("X-User-Id")
	if userId == "" {
		return 0, errors.New("missing user id")
	}
	id, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		return 0, errors.New("invalid user id")
	}
	return id, nil
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
		// TODO log error
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

func ReturnCreated(w http.ResponseWriter, data interface{}) {
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

func returnInternalServerError(w http.ResponseWriter, err error) {
	returnError(w, http.StatusInternalServerError, err)
}
