package db2

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type Store interface {
	AddIndexedBlock(block IndexedBlock) error
	RevertBlock(chainID string, number uint64, blockHash []byte) error
}

type MetadataStore interface {
	UpdateBalance(chainID string, balances []Balance) error
	UpdateToken(chainID string, tokens []*types.ERC20TokenPrice) error
	TokenPrice(chainID string, token common.Address) (*types.ERC20TokenPrice, error)
}

type BlocksStore interface {
	GetBlock(chainID string, number uint64) (*types.Eth1Block, error)
}

type Cache interface {
	Set(key, value []byte, expireSeconds int) (err error)
	Get(key []byte) (value []byte, err error)
}

type LastBlocksStore interface {
	SetInBlocksTable(chainID string, number uint64) error
	SetInDataTable(chainID string, number uint64) error
	GetInDataTable(chainID string) (uint64, error)
	GetInBlocksTable(chainID string) (uint64, error)
}

type LastBlocksStoreWriter interface {
	SetInBlocksTable(chainID string, number uint64) error
	SetInDataTable(chainID string, number uint64) error
}
