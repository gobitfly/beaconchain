package network

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/consapi/utils"
)

// Helper for get and unmarshal
func Get[T any](r *http.Client, url string) (*T, error) {
	return retry[T]("GET", r, url)
}

// Helper for post and unmarshal
func Post[T any](r *http.Client, url string) (*T, error) {
	return retry[T]("POST", r, url)
}

func retry[T any](method string, r *http.Client, url string) (*T, error) {
	const maxRetries = 16             // Maximum number of retries
	var backoffTime = 1 * time.Second // Initial backoff time

	var resp *http.Response
	var err error
	var e *T
	tmr := time.AfterFunc(60*time.Second, func() {
		log.WarnWithFields(log.Fields{"url": url}, fmt.Sprintf("%s request taking more than 60 seconds", method))
	})
	defer tmr.Stop()

	for attempt := 0; attempt < maxRetries; attempt++ {
		start := time.Now()
		resp, err = HTTPReq(method, url, r)

		if resp != nil {
			e, err = utils.Unmarshal[T](resp.Body, err)
		}
		if time.Since(start) > 30*time.Second {
			log.Debugf("%s %s took %s", method, url, time.Since(start))
		}
		if err == nil {
			defer resp.Body.Close()
			break
		}
		// retry as long as it isn't a 404 http error or there has been no error
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			break
		}
		httpErr := SpecificError(err)
		if httpErr != nil && httpErr.StatusCode == http.StatusNotFound {
			break
		}

		log.Warnf("Attempt %d for %s %s failed: %v. Retrying in %v...", attempt+1, method, url, err, backoffTime)
		time.Sleep(backoffTime)
	}
	if err != nil {
		err = fmt.Errorf("after %d attempts, last error: %w", maxRetries, err)
	}

	return e, err
}

func HTTPReq(method string, requestURL string, httpClient *http.Client) (*http.Response, error) {
	data := []byte{}
	if method == "POST" {
		data = []byte("[]")
	}
	r, err := http.NewRequest(method, requestURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	if httpClient == nil {
		return nil, errors.New("httpClient is nil")
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

	return res, nil
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
