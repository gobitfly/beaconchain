package cache

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

// LatestEpoch will return the latest epoch
func LatestEpoch() uint64 {
	cacheKey := fmt.Sprintf("%d:frontend:latestEpoch", utils.Config.Chain.ClConfig.DepositChainID)

	if wanted, err := TieredCache.GetUint64WithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return wanted
	} else {
		utils.LogError(err, "error retrieving latestEpoch from cache", 0, nil)
	}

	return 0
}

func LatestNodeEpoch() uint64 {
	cacheKey := fmt.Sprintf("%d:frontend:latestNodeEpoch", utils.Config.Chain.ClConfig.DepositChainID)

	if wanted, err := TieredCache.GetUint64WithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return wanted
	} else {
		utils.LogError(err, "error retrieving latestNodeEpoch from cache: %v", 0, nil)
	}

	return 0
}

func LatestNodeFinalizedEpoch() uint64 {
	cacheKey := fmt.Sprintf("%d:frontend:latestNodeFinalizedEpoch", utils.Config.Chain.ClConfig.DepositChainID)

	if wanted, err := TieredCache.GetUint64WithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return wanted
	} else {
		utils.LogError(err, "error retrieving latestNodeFinalizedEpoch from cache: %v", 0, nil)
	}

	return 0
}

// LatestFinalizedEpoch will return the most recent epoch that has been finalized.
func LatestFinalizedEpoch() uint64 {
	cacheKey := fmt.Sprintf("%d:frontend:latestFinalized", utils.Config.Chain.ClConfig.DepositChainID)

	if wanted, err := TieredCache.GetUint64WithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return wanted
	} else {
		utils.LogError(err, "error retrieving latestFinalized for key: %v from cache: %v", 0, map[string]interface{}{"cacheKey": cacheKey, "err": err})
	}
	return 0
}

// LatestSlot will return the latest slot
func LatestSlot() uint64 {
	cacheKey := fmt.Sprintf("%d:frontend:slot", utils.Config.Chain.ClConfig.DepositChainID)

	if wanted, err := TieredCache.GetUint64WithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return wanted
	} else {
		utils.LogError(err, "error retrieving latest slot from cache: %v", 0, nil)
	}
	return 0
}

// FinalizationDelay will return the current Finalization Delay
func FinalizationDelay() uint64 {
	return LatestNodeEpoch() - LatestNodeFinalizedEpoch()
}

// LatestProposedSlot will return the latest proposed slot
func LatestProposedSlot() uint64 {
	cacheKey := fmt.Sprintf("%d:frontend:latestProposedSlot", utils.Config.Chain.ClConfig.DepositChainID)

	if wanted, err := TieredCache.GetUint64WithLocalTimeout(cacheKey, time.Second*5); err == nil {
		return wanted
	} else {
		utils.LogError(err, "error retrieving latestProposedSlot from cache: %v", 0, nil)
	}
	return 0
}
