package db2

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type Store interface {
	AddIndexedBlock(block IndexedBlock) error
}

type MetadataStore interface {
	UpdateContract(chainID string, blockNumber uint64, updates []ContractUpdateWithAddress) error
	UpdateBalance(chainID string, balances []Balance) error
	UpdateToken(chainID string, tokens []*types.ERC20TokenPrice) error
	TokenPrice(chainID string, token common.Address) (*types.ERC20TokenPrice, error)
}

type Cache interface {
	Set(key, value []byte, expireSeconds int) (err error)
	Get(key []byte) (value []byte, err error)
}
