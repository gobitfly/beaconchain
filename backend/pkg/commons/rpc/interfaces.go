package rpc

import (
	"math/big"

	"github.com/gobitfly/beaconchain/pkg/commons/types"

	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
)

// Client provides an interface for RPC clients
type Client interface {
	BlockClient
	EpochClient
	SyncCommittee
	ValidatorClient
}

type BlockClient interface {
	GetNewBlockChan() chan *types.Block
	GetBlockHeader(slot uint64) (*constypes.StandardBeaconHeaderResponse, error)
	GetBlockBySlot(slot uint64) (*types.Block, error)
}

type EpochClient interface {
	GetChainHead() (*types.ChainHead, error)
	GetEpochData(epoch uint64, skipHistoricBalances bool) (*types.EpochData, error)
	GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error)
	GetBalancesForEpoch(epoch int64) (map[uint64]uint64, error)
}

type SyncCommittee interface {
	GetSyncCommittee(stateID string, epoch uint64) (*constypes.StandardSyncCommittee, error)
}

type ValidatorClient interface {
	GetValidatorQueue() (*types.ValidatorQueue, error)
	GetValidatorState(epoch uint64) (*constypes.StandardValidatorsResponse, error)
	GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error)
}

type Eth1Client interface {
	GetBlock(number uint64) (*types.Eth1Block, *types.GetBlockTimings, error)
	GetLatestEth1BlockNumber() (uint64, error)
	GetChainID() *big.Int
	Close()
}
