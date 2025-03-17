package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
	"github.com/gobitfly/beaconchain/pkg/api/services"
	"github.com/gobitfly/beaconchain/pkg/api/types"
	commontypes "github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func handlerTestSetup(da dataaccess.DataAccessor) (context.Context, *HandlerService) {
	ctx := context.WithValue(context.Background(), types.CtxUserIdKey, uint64(1))
	cfg := &commontypes.Config{
		Chain: commontypes.Chain{
			ClConfig: commontypes.ClChainConfig{
				SecondsPerSlot: 12,
				SlotsPerEpoch:  32,
			},
		},
	}

	return ctx, NewHandlerService(da, da, nil, cfg)
}

func stringAsBody(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

type inputMock struct {
	shouldFail     bool
	errorType      string
	successMessage string
}

func (i *inputMock) Validate(params map[string]string, payload io.ReadCloser) error {
	var v validationError

	i.shouldFail = v.checkBool(params["should_fail"], "should_fail")
	i.errorType = params["error_type"]
	if !i.shouldFail {
		type request struct {
			SuccessMessage string `json:"success_message"`
		}
		var req request
		if err := v.checkBody(&req, payload); err != nil {
			return err
		}
		i.successMessage = req.SuccessMessage
	}
	return v.AsError()
}

func logicMock(ctx context.Context, input inputMock) (string, error) {
	if !input.shouldFail {
		return input.successMessage, nil
	}
	switch input.errorType {
	case "bad_request":
		return "", newBadRequestErr("test 400 error")
	case "forbidden":
		return "", newForbiddenErr("test 403 error")
	default:
		return "", newInternalServerErr("test 500 error")
	}
}
func TestHandle(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		vars         map[string]string
		bodyStr      string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Success",
			url:          "/test",
			bodyStr:      `{"success_message":"success"}`,
			expectedCode: http.StatusOK,
			expectedBody: "success",
		},
		{
			name:         "QueryParamFailure",
			url:          "/test?should_fail=true",
			expectedCode: http.StatusInternalServerError,
			expectedBody: "test 500 error",
		},
		{
			name:         "VarsParamFailure",
			url:          "/test",
			vars:         map[string]string{"should_fail": "true"},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "test 500 error",
		},
		{
			name:         "Vars Overwrites Query - Success",
			url:          "/test?should_fail=true",
			vars:         map[string]string{"should_fail": "false"},
			bodyStr:      `{"success_message":"successful overwrite"}`,
			expectedCode: http.StatusOK,
			expectedBody: "successful overwrite",
		},
		{
			name:         "Vars Overwrites Query - Failure",
			url:          "/test?should_fail=false",
			vars:         map[string]string{"should_fail": "true"},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "test 500 error",
		},
		{
			name:         "BadRequest",
			url:          "/test",
			vars:         map[string]string{"should_fail": "true", "error_type": "bad_request"},
			expectedCode: http.StatusBadRequest,
			expectedBody: "test 400 error",
		},
		{
			name:         "Forbidden",
			url:          "/test",
			vars:         map[string]string{"should_fail": "true", "error_type": "forbidden"},
			expectedCode: http.StatusForbidden,
			expectedBody: "test 403 error",
		},
		{
			name:         "InternalServerError",
			url:          "/test",
			vars:         map[string]string{"should_fail": "true", "error_type": "internal_server_error"},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "test 500 error",
		},
		{
			name:         "Invalid Input",
			url:          "/test",
			vars:         map[string]string{"should_fail": "abc"},
			expectedCode: http.StatusBadRequest,
			expectedBody: "should_fail: given value 'abc' is not a boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, stringAsBody(tt.bodyStr))
			req.Header.Set("Content-Type", "application/json")
			if tt.vars != nil {
				req = mux.SetURLVars(req, tt.vars)
			}
			w := httptest.NewRecorder()

			handler := Handle(http.StatusOK, logicMock, false)
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			assert.Contains(t, string(body), tt.expectedBody)
		})
	}
}
func TestWriteResponse_EmptyResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "Nil response",
			statusCode: http.StatusNoContent,
		},
		{
			name:       "Nil response - other code",
			statusCode: http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			writeResponse(rec, req, tt.statusCode, nil)

			result := rec.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.statusCode, result.StatusCode, "Unexpected status code")

			bodyBytes := new(bytes.Buffer)
			_, err := bodyBytes.ReadFrom(result.Body)
			assert.NoError(t, err, "Error reading response body")

			assert.Equal(t, "", bodyBytes.String(), "Expected empty body")
		})
	}
}

func TestWriteResponse_JSONResponse(t *testing.T) {
	tests := []struct {
		name         string
		inputCode    int
		response     interface{}
		expectedBody string
		expectedCode int
	}{
		{
			name:      "Valid JSON response",
			inputCode: http.StatusOK,
			response: map[string]string{
				"message": "success",
			},
			expectedBody: `{"message":"success"}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Error encoding JSON",
			inputCode:    http.StatusOK,
			response:     make(chan int), // Invalid JSON type
			expectedBody: `{"error":"error encoding json data"}`,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			writeResponse(rec, req, tt.inputCode, tt.response)

			result := rec.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expectedCode, result.StatusCode, "Unexpected status code")

			bodyBytes := new(bytes.Buffer)
			_, err := bodyBytes.ReadFrom(result.Body)
			assert.NoError(t, err, "Error reading response body")

			assert.JSONEq(t, tt.expectedBody, bodyBytes.String(), "Unexpected response body")
		})
	}
}

func TestHandleErr(t *testing.T) {
	var testValidationErr = validationError{"field": "wrong"}
	tests := []struct {
		name         string
		err          error
		expectedCode int
		expectedBody string
	}{
		{"ValidationError", testValidationErr.AsError(), http.StatusBadRequest, `{"error":"bad request: field: wrong"}`},
		{"BadRequest", errBadRequest, http.StatusBadRequest, `{"error":"bad request"}`},
		{"NotFound", dataaccess.ErrNotFound, http.StatusNotFound, `{"error":"not found"}`},
		{"Unauthorized", errUnauthorized, http.StatusUnauthorized, `{"error":"unauthorized"}`},
		{"Forbidden", errForbidden, http.StatusForbidden, `{"error":"forbidden"}`},
		{"Conflict", errConflict, http.StatusConflict, `{"error":"conflict"}`},
		{"ServiceUnavailable", services.ErrWaiting, http.StatusServiceUnavailable, `{"error":"waiting for service to be initialized"}`},
		{"TooManyRequests", errTooManyRequests, http.StatusTooManyRequests, `{"error":"too many requests"}`},
		{"InternalServerError", errInternalServer, http.StatusInternalServerError, `{"error":"internal server error"}`},
		{"Gone", errGone, http.StatusGone, `{"error":"gone"}`},
		{"UnknownError", errors.New("test error"), http.StatusInternalServerError, `{"error":"test error"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()
			handleErr(rr, req, tt.err)

			// Validate HTTP status code
			assert.Equal(t, tt.expectedCode, rr.Code, "unexpected HTTP status code")

			// Validate response body
			assert.Equal(t, tt.expectedBody, rr.Body.String(), "unexpected response body")
		})
	}
}
func TestHandleErr_CanceledContext(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		// Simulate context cancellation
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		req = req.WithContext(ctx)

		handleErr(rr, req, context.Canceled)

		// Validate HTTP status code
		assert.Equal(t, http.StatusOK, rr.Code, "unexpected HTTP status code")

		// Validate response body
		assert.Equal(t, "", rr.Body.String(), "unexpected response body")
	})
	t.Run("InternalServerError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handleErr(rr, req, context.Canceled)

		// Validate HTTP status code
		assert.Equal(t, http.StatusInternalServerError, rr.Code, "unexpected HTTP status code")

		// Validate response body
		assert.Equal(t, `{"error":"context canceled"}`, rr.Body.String(), "unexpected response body")
	})
}
