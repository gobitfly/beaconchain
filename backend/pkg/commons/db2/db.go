package db2

import (
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type Store interface {
	AddIndexedBlock(block IndexedBlock) error
	RevertBlock(chainID string, number uint64, blockHash []byte) error
}

type MetadataStore interface {
	CountBalanceUpdates(chainID string) (int64, error)
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

type Monitoring interface {
	SaveNewStatusReport(status StatusReport) error
	GetLatestStatusReport() ([]Victims, error)
	GetEmitters() ([]string, error)
	GetLatestEpoch() (time.Time, error)
	GetEpochEnd(rolling string) (uint64, error)
}
