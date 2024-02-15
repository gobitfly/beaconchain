package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HttpReqHttpError struct {
	StatusCode int
	Url        string
	Body       []byte
}

func (err *HttpReqHttpError) Error() string {
	return fmt.Sprintf("error response: url: %s, status: %d, body: %s", err.Url, err.StatusCode, err.Body)
}

func HttpReq(ctx context.Context, method, url string, params, result interface{}) error {
	var err error
	var req *http.Request
	if params != nil {
		paramsJSON, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("error marshaling params for request: %w, url: %v", err, url)
		}
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(paramsJSON))
		if err != nil {
			return fmt.Errorf("error creating request with params: %w, url: %v", err, url)
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return fmt.Errorf("error creating request: %w, url: %v", err, url)
		}
	}
	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{Timeout: time.Minute}
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return &HttpReqHttpError{
			StatusCode: res.StatusCode,
			Url:        url,
			Body:       body,
		}
	}
	if result != nil {
		err = json.NewDecoder(res.Body).Decode(result)
		if err != nil {
			return fmt.Errorf("error unmarshaling response: %w, url: %v", err, url)
		}
	}
	return nil
}
