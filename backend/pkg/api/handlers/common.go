package apihandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"

	apitypes "github.com/gobitfly/beaconchain/pkg/types/api"
	"github.com/invopop/jsonschema"
	"github.com/xeipuuv/gojsonschema"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
)

type HandlerService struct {
	dai dataaccess.DataAccessInterface
}

func NewHandlerService(das dataaccess.DataAccessInterface) HandlerService {
	return HandlerService{dai: das}
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
	maxNameLength       = 50
	maxQueryLimit       = 100
	defaultReturnLimit  = 10
	sortOrderAscending  = "asc"
	sortOrderDescending = "desc"
	defaultSortOrder    = sortOrderAscending
	ethereum            = "ethereum"
	gnosis              = "gnosis"
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
	limit  int
	order  string
	sort   string
	search string
}

// All changes to common functions MUST NOT break any public handler behavior

// --------------------------------------
//   Validation

func joinErr(err *error, message string) {
	*err = errors.Join(*err, errors.New(message))
}

func regexCheck(regex regexString, param string) error {
	if !regexp.MustCompile(string(regex)).MatchString(param) {
		return fmt.Errorf(`given value '%s' has incorrect format`, param)
	}
	return nil
}

func checkName(name string, minLength int) error {
	if len(name) < minLength {
		return fmt.Errorf(`given value '%s' for parameter "name" is too short, minimum length is %d`, name, minLength)
	} else if len(name) > 50 {
		return fmt.Errorf(`given value '%s' for parameter "name" is too long, maximum length is %d`, name, maxNameLength)
	}
	return regexCheck(reName, name)
}

func regexCheckMultiple(regexes []regexString, params []string) error {
	var err error
	for _, param := range params {
		for _, regex := range regexes {
			err = errors.Join(err, regexCheck(regex, param))
		}
	}
	return err
}

func CheckNameNotEmpty(name string) error {
	return checkName(name, 1)
	// return name
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
		fmt.Printf("error validating json: %s\n", err.Error())
		return RequestError{http.StatusInternalServerError, errors.New("can't validate expected format")}
	}
	loader := gojsonschema.NewBytesLoader(b)
	documentLoader, _ := gojsonschema.NewReaderLoader(r)
	schema, err := gojsonschema.NewSchema(loader)
	if err != nil {
		fmt.Printf("error validating json: %s\n", err.Error())
		return RequestError{http.StatusInternalServerError, errors.New("can't create expected format")}
	}
	result, err := schema.Validate(documentLoader)
	if err != nil {
		fmt.Printf("error validating json: %s\n", err.Error())
		return RequestError{http.StatusInternalServerError, errors.New("couldn't validate JSON request")}
	}
	if !result.Valid() {
		return RequestError{http.StatusBadRequest, errors.New("unexpected JSON format. Check the API documentation for parameter details")}
	}
	if err = json.NewDecoder(r).Decode(data); err != nil {
		// error parsing json; shouldn't happen since we verified it's json in the correct format already
		fmt.Printf("error validating json: %s\n", err.Error())
		return RequestError{http.StatusInternalServerError, errors.New("couldn't decode JSON request")}
	}
	// could perform data validation checks based on tags here, but might need validation lib for that
	return nil
}

func CheckId(id string) error {
	return regexCheck(reId, id)
}

func CheckIdList(ids []string) error {
	return regexCheckMultiple([]regexString{reId}, ids)
}

func CheckAndGetPaging(r *http.Request) (Paging, error) {
	q := r.URL.Query()
	paging := Paging{
		cursor: q.Get("cursor"),
		limit:  defaultReturnLimit,
		order:  defaultSortOrder,
		sort:   q.Get("sort"),
		search: q.Get("search"),
	}

	var paging_limit_error error
	if limit_str := q.Get("limit"); limit_str != "" {
		paging.limit, paging_limit_error = strconv.Atoi(limit_str)
		if paging.limit > maxQueryLimit {
			paging_limit_error = fmt.Errorf("Paging limit %d is too high, maximum value is %d", paging.limit, maxQueryLimit)
		}
	}

	var paging_order_error error
	if order := q.Get("order"); order != "" {
		paging.order = order
	}
	if paging.order != sortOrderAscending && paging.order == sortOrderDescending {
		paging_order_error = fmt.Errorf("invalid sorting order: %s", paging.order)
	}
	return paging,
		errors.Join(
			regexCheck(reCursor, paging.cursor),
			paging_order_error,
			paging_limit_error,
			checkName(paging.sort, 0),
			checkName(paging.search, 0),
		)
}

func CheckValidatorList(validators []string) error {
	return regexCheckMultiple([]regexString{reNumber, reValidatorPubkey}, validators)
}

func CheckNetwork(network string) error {
	if network != ethereum && network != gnosis {
		return fmt.Errorf(`given parameter '%s' for "network" isn't valid, allowed values are: %s, %s`, network, ethereum, gnosis)
	}
	return nil
}

// --------------------------------------
//   Response handling

func writeResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.WriteHeader(statusCode)

	if response != nil {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error encoding json data"))
		}
	}
}

func returnError(w http.ResponseWriter, code int, err error) {
	response := apitypes.ApiErrorResponse{
		Error: err.Error(),
	}
	writeResponse(w, code, response)
}

func ReturnOk(w http.ResponseWriter, data interface{}) {
	writeResponse(w, http.StatusOK, data)
}

func ReturnCreated(w http.ResponseWriter, data interface{}) {
	writeResponse(w, http.StatusCreated, data)
}

func ReturnNoContent(w http.ResponseWriter) {
	writeResponse(w, http.StatusNoContent, nil)
}

// Errors

func ReturnBadRequest(w http.ResponseWriter, err error) {
	returnError(w, http.StatusBadRequest, err)
}

func ReturnNotFound(w http.ResponseWriter, err error) {
	returnError(w, http.StatusNotFound, err)
}

func ReturnConflict(w http.ResponseWriter, err error) {
	returnError(w, http.StatusConflict, err)
}

func ReturnInternalServerError(w http.ResponseWriter, err error) {
	returnError(w, http.StatusInternalServerError, err)
}
