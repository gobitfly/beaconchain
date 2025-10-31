package apputils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/gobitfly/beaconchain-backend/api/external/model"

	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/log"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

func ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	// Default response
	status := http.StatusInternalServerError
	msg := common.GenericErrMsg

	if errors.Is(err, io.EOF) { // gen-server returns an untyped EOF error on empty body
		err = common.NewAPIUserFacingError(http.StatusBadRequest, "empty request body")
	}

	// Handle our custom API errors (user facing and internal)
	var apiErr *common.APIError
	if errors.As(err, &apiErr) {
		status = apiErr.Status
		if apiErr.Visibility == common.ErrorVisibilityUser {
			msg = apiErr.Message
		}
		log.WarnWithFields(apiErr.Extras, fmt.Sprintf("API error: %s | Status: %d | Path: %s ", apiErr.Message, apiErr.Status, r.URL.Path))
	} else {
		// Bad request errors as thrown by generated server
		switch e := err.(type) {
		case *model.InvalidParamFormatError, *model.RequiredParamError:
			status = http.StatusBadRequest
			msg = e.Error()
		default:
			log.Error(fmt.Sprintf("Internal error: %v | Path: %s", err, r.URL.Path))
		}
	}

	w.WriteHeader(status)
	if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": msg}); encodeErr != nil {
		log.Error(fmt.Sprintf("failed to encode error response: %v", encodeErr))
	}
}

func ValidationErrorHandler(_ context.Context, err error, w http.ResponseWriter, r *http.Request, opts nethttpmiddleware.ErrorHandlerOpts) {
	// we only expect request errors from validation
	switch v := err.(type) {
	case *routers.RouteError:
		// middleware defaults to 404, but we want to distinguish between 404 and 405
		switch v {
		case routers.ErrPathNotFound:
			err = common.NewAPIUserFacingError(http.StatusNotFound, fmt.Sprintf("path not found: %s", r.URL.Path))
		case routers.ErrMethodNotAllowed:
			err = common.NewAPIUserFacingError(http.StatusMethodNotAllowed, fmt.Sprintf("method not allowed: %s", r.Method))
		}
	case *openapi3filter.RequestError, *openapi3filter.ValidationError:
		// openapi errors seem to be multi-line with a decent message on the first
		// Split up the verbose error by lines and return the first one
		errorLines := strings.Split(v.Error(), "\n")
		err = common.NewAPIUserFacingError(opts.StatusCode, errorLines[0])
	default:
		err = common.NewAPIInternalError(http.StatusInternalServerError, fmt.Sprintf("unexpected err in validation middleware: %v", err))
	}
	ErrorHandler(w, r, err)
}
