package rpc

import (
	"math/big"

	"github.com/gobitfly/beaconchain/pkg/commons/types"
	constypes "github.com/gobitfly/beaconchain/pkg/consapi/types"
	"github.com/sirupsen/logrus"
)

// Client provides an interface for RPC clients
type Client interface {
	GetChainHead() (*types.ChainHead, error)
	GetEpochData(epoch uint64, skipHistoricBalances bool) (*types.EpochData, error)
	GetValidatorQueue() (*types.ValidatorQueue, error)
	GetEpochAssignments(epoch uint64) (*types.EpochAssignments, error)
	GetBlockBySlot(slot uint64) (*types.Block, error)
	GetValidatorParticipation(epoch uint64) (*types.ValidatorParticipation, error)
	GetNewBlockChan() chan *types.Block
	GetSyncCommittee(stateID string, epoch uint64) (*constypes.StandardSyncCommittee, error)
	GetBalancesForEpoch(epoch int64) (map[uint64]uint64, error)
	GetValidatorState(epoch uint64) (*constypes.StandardValidatorsResponse, error)
	GetBlockHeader(slot uint64) (*constypes.StandardBeaconHeaderResponse, error)
}

type Eth1Client interface {
	GetBlock(number uint64) (*types.Eth1Block, *types.GetBlockTimings, error)
	GetLatestEth1BlockNumber() (uint64, error)
	GetChainID() *big.Int
	Close()
}

var logger = logrus.New().WithField("module", "rpc")
