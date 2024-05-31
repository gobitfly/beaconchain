package network

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gobitfly/beaconchain/pkg/consapi/utils"
)

// Helper for get and unmarshal
func Get[T any](r *http.Client, url string) (*T, error) {
	result, err := HTTPReq("GET", url, r)
	if err != nil || result == nil {
		var target T
		return &target, err
	}
	return utils.Unmarshal[T](result, err)
}

// Helper for post and unmarshal
func Post[T any](r *http.Client, url string) (*T, error) {
	result, err := HTTPReq("POST", url, r)
	if err != nil || result == nil {
		var target T
		return &target, err
	}
	return utils.Unmarshal[T](result, err)
}

func HTTPReq(method string, requestURL string, httpClient *http.Client) (io.ReadCloser, error) {
	data := []byte{}
	if method == "POST" {
		data = []byte("[]")
	}
	r, err := http.NewRequest(method, requestURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}

	r.Header.Add("Content-Type", "application/json")

	res, err := httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, &HttpReqHttpError{
			StatusCode: res.StatusCode,
			Url:        requestURL,
			Body:       body,
		}
	}

	return res.Body, nil
}

type RPCErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type HttpReqHttpError struct {
	StatusCode int
	Url        string
	Body       []byte
}

func (err *HttpReqHttpError) Error() string {
	return fmt.Sprintf("error response: url: %s, status: %d, body: %s", err.Url, err.StatusCode, err.Body)
}

type RPCError struct {
	Code    int
	Message string
	Url     string
}

func (err *RPCError) Error() string {
	return fmt.Sprintf("error rpc: url: %s, code: %d, message: %s", err.Url, err.Code, err.Message)
}

func SpecificError(err error) *HttpReqHttpError {
	var apiErr *HttpReqHttpError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}
