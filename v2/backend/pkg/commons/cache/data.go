package cache

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
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

var LatestExportedStatisticDay UInt64Cached = UInt64Cached{
	cacheKey: func() string {
		return fmt.Sprintf("%d:frontend:lastExportedStatisticDay", utils.Config.Chain.ClConfig.DepositChainID)
	},
}

var LatestStats Cached[types.Stats] = Cached[types.Stats]{
	cacheKey: func() string {
		return fmt.Sprintf("%d:frontend:latestStats", utils.Config.Chain.ClConfig.DepositChainID)
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
		log.Error(err, "error retrieving uint64 for key", 0, map[string]interface{}{"cacheKey": cfg.cacheKey(), "err": err})
	}
	return 0
}

func (cfg UInt64Cached) GetOrDefault(provideDefault func() (uint64, error)) (uint64, error) {
	if wanted, err := TieredCache.GetUint64WithLocalTimeout(cfg.cacheKey(), time.Second*5); err == nil {
		return wanted, nil
	}
	return provideDefault()
}

func (cfg UInt64Cached) Set(epoch uint64) error {
	return TieredCache.SetUint64(cfg.cacheKey(), epoch, utils.Day)
}

type Cached[T any] struct {
	cacheKey func() string
}

func (cfg Cached[T]) Get() T {
	var wanted T
	if wanted, err := TieredCache.GetWithLocalTimeout(cfg.cacheKey(), time.Second*5, &wanted); err == nil {
		value, ok := wanted.(T)
		if !ok {
			log.Error(err, "error during type assertion for key", 0, map[string]interface{}{"cacheKey": cfg.cacheKey(), "err": err})
		} else {
			return value
		}
	} else {
		log.Error(err, "error retrieving values for key", 0, map[string]interface{}{"cacheKey": cfg.cacheKey(), "err": err})
	}
	var defaultValue T
	return defaultValue
}

func (cfg Cached[T]) GetOrDefault(provideDefault func() (T, error)) (T, error) {
	var wanted T
	if wanted, err := TieredCache.GetWithLocalTimeout(cfg.cacheKey(), time.Second*5, &wanted); err == nil {
		value, ok := wanted.(T)
		if !ok {
			log.Error(err, "error during type assertion for key", 0, map[string]interface{}{"cacheKey": cfg.cacheKey(), "err": err})
		} else {
			return value, nil
		}
	}
	return provideDefault()
}

func (cfg Cached[T]) Set(value T) error {
	return TieredCache.Set(cfg.cacheKey(), value, utils.Day)
}
