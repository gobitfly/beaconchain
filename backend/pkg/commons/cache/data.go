package cache

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

// LatestEpoch will return the latest epoch
var LatestEpoch UInt64Cached = UInt64Cached{
	cacheKey: func() string {
		return fmt.Sprintf("%d:frontend:latestEpoch", utils.Config.Chain.ClConfig.DepositChainID)
	},
}

var LatestNodeEpoch UInt64Cached = UInt64Cached{
	cacheKey: func() string {
		return fmt.Sprintf("%d:frontend:latestNodeEpoch", utils.Config.Chain.ClConfig.DepositChainID)
	},
}

var LatestNodeFinalizedEpoch UInt64Cached = UInt64Cached{
	cacheKey: func() string {
		return fmt.Sprintf("%d:frontend:latestNodeFinalizedEpoch", utils.Config.Chain.ClConfig.DepositChainID)
	},
}

// LatestFinalizedEpoch will return the most recent epoch that has been finalized.
var LatestFinalizedEpoch UInt64Cached = UInt64Cached{
	cacheKey: func() string {
		return fmt.Sprintf("%d:frontend:latestFinalized", utils.Config.Chain.ClConfig.DepositChainID)
	},
}

// LatestSlot will return the latest slot
var LatestSlot UInt64Cached = UInt64Cached{
	cacheKey: func() string {
		return fmt.Sprintf("%d:frontend:slot", utils.Config.Chain.ClConfig.DepositChainID)
	},
}

// LatestProposedSlot will return the latest proposed slot
var LatestProposedSlot UInt64Cached = UInt64Cached{
	cacheKey: func() string {
		return fmt.Sprintf("%d:frontend:latestProposedSlot", utils.Config.Chain.ClConfig.DepositChainID)
	},
}

// FinalizationDelay will return the current Finalization Delay
func FinalizationDelay() uint64 {
	return LatestNodeEpoch.Get() - LatestNodeFinalizedEpoch.Get()
}

type UInt64Cached struct {
	cacheKey func() string
}

func (cfg UInt64Cached) Get() uint64 {
	if wanted, err := TieredCache.GetUint64WithLocalTimeout(cfg.cacheKey(), time.Second*5); err == nil {
		return wanted
	} else {
		log.LogError(err, "error retrieving uint64 for key", 0, map[string]interface{}{"cacheKey": cfg.cacheKey(), "err": err})
	}
	return 0
}

func (cfg UInt64Cached) Set(epoch uint64) error {
	return TieredCache.SetUint64(cfg.cacheKey(), epoch, utils.Day)
}
