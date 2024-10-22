package utils

import (
	"io"

	"github.com/pkg/errors"
	"github.com/segmentio/encoding/json"
)

func Unmarshal[T any](source io.ReadCloser, err error) (*T, error) {
	var target T
	if err != nil {
		return &target, err
	}

	if err := json.NewDecoder(source).Decode(&target); err != nil {
		return &target, errors.Wrap(err, "unmarshal json failed")
	}

	return &target, nil
}

func UnmarshalOld[T any](source []byte, err error) (*T, error) {
	var target T
	if err != nil {
		return &target, err
	}

	if err := json.Unmarshal(source, &target); err != nil {
		return &target, errors.Wrap(err, "unmarshal json failed")
	}

	return &target, nil
}
