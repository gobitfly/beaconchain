package indexer

import (
	"fmt"
)

type NoopCache struct{}

func (n NoopCache) Set(key, value []byte, expireSeconds int) error {
	return nil
}

func (n NoopCache) Get(key []byte) ([]byte, error) {
	return nil, fmt.Errorf("noopCache")
}
