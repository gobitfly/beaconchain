// Package defines common service-layer logic for cursor-based pagination
package pagination

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
)

// Handle is a service-layer level wrapper function that handles cursor-based pagination.
// It parses the input cursor, calls the provided fetch function to retrieve data, and generates the next cursor if more data is available.
func Handle[Cursor any, Model any](
	cursorStr string,
	pageSize int,
	transform func(Model) Cursor,
	fetch func(cursor *Cursor, pageSize int) ([]Model, error)) ([]Model, model.Paging, error) {
	// parse cursor
	cursor, err := fromBase64JSONString[Cursor](cursorStr)
	if err != nil {
		return nil, model.Paging{}, fmt.Errorf("failed to parse cursor: %w", err)
	}

	// fetch data
	data, err := fetch(cursor, pageSize+1) // fetch one additional item to determine if there's a next page
	if err != nil {
		return nil, model.Paging{}, fmt.Errorf("failed to fetch data: %w", err)
	}
	if len(data) <= pageSize {
		// no more data, no next cursor
		return data, model.Paging{}, nil
	}

	data = data[:min(len(data), pageSize)] // trim to requested page size

	// generate next cursor
	nextCursor := transform(data[len(data)-1])
	nextCursorStr, err := toBase64JSONString(nextCursor)
	if err != nil {
		return nil, model.Paging{}, fmt.Errorf("failed to generate next_cursor: %w", err)
	}

	return data, model.Paging{NextCursor: nextCursorStr}, nil
}

func fromBase64JSONString[T any](str string) (*T, error) {
	if str == "" {
		return nil, nil
	}

	// decode base64 string
	bin, err := base64.RawURLEncoding.DecodeString(str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode string using base64: %w", err)
	}

	var t T
	d := json.NewDecoder(bytes.NewReader(bin))
	d.DisallowUnknownFields() // this optimistically prevents parsing a cursor into a wrong type
	err = d.Decode(&t)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal decoded base64 string: %w", err)
	}
	return &t, nil
}

func toBase64JSONString(src any) (string, error) {
	bytes, err := json.Marshal(src)
	if err != nil {
		return "", fmt.Errorf("failed to marshal src as json: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
