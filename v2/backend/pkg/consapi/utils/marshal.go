package utils

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

func Unmarshal[T any](source io.ReadCloser, err error) (*T, error) {
	defer source.Close()
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
