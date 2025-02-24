package modules

import (
	"math"
	"math/big"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

type RocketpoolNetworkStats struct {
	RPLPrice               *big.Int
	ClaimIntervalTime      time.Duration
	ClaimIntervalTimeStart time.Time
	CurrentNodeFee         float64
	CurrentNodeDemand      *big.Int
	RETHSupply             *big.Int
	NodeOperatorRewards    *big.Int
	RETHPrice              float64
	TotalEthStaking        *big.Int
	TotalEthBalance        *big.Int
}

func (rp *RocketpoolExporter) UpdateNetworkStats() error {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.DebugWithFields(log.Fields{"duration": time.Since(timeStart)}, "updated rocketpool-network-stats")
	}(timeStart)

	price, err := network.GetRPLPrice(rp.API, nil)
	if err != nil {
		return err
	}
	claimIntervalTime, err := rewards.GetClaimIntervalTime(rp.API, nil)
	if err != nil {
		return err
	}
	claimIntervalTimeStart, err := rewards.GetClaimIntervalTimeStart(rp.API, nil)
	if err != nil {
		return err
	}

	currentNodeFee, err := network.GetNodeFee(rp.API, nil)
	if err != nil {
		return err
	}

	currentNodeDemand, err := network.GetNodeDemand(rp.API, nil)
	if err != nil {
		return err
	}

	exchangeRate, err := tokens.GetRETHExchangeRate(rp.API, nil)
	if err != nil {
		return err
	}

	rethSupply, err := network.GetTotalRETHSupply(rp.API, nil)
	if err != nil {
		return err
	}

	isMergeUpdateDeployed, err := IsMergeUpdateDeployed(rp.API)
	if err != nil {
		return err
	}

	nodeOperatorRewards, err := rp.getNodeOperatorRewards(isMergeUpdateDeployed, claimIntervalTime)
	if err != nil {
		return err
	}

	totalEthStaking, err := network.GetStakingETHBalance(rp.API, nil)
	if err != nil {
		return err
	}

	totalEthBalance, err := network.GetTotalETHBalance(rp.API, nil)
	if err != nil {
		return err
	}

	rp.NetworkStats = RocketpoolNetworkStats{
		RPLPrice:               price,
		ClaimIntervalTime:      claimIntervalTime,
		ClaimIntervalTimeStart: claimIntervalTimeStart,
		CurrentNodeFee:         currentNodeFee,
		CurrentNodeDemand:      currentNodeDemand,
		RETHSupply:             rethSupply,
		NodeOperatorRewards:    nodeOperatorRewards,
		RETHPrice:              exchangeRate,
		TotalEthStaking:        totalEthStaking,
		TotalEthBalance:        totalEthBalance,
	}
	return err
}

func (rp *RocketpoolExporter) getNodeOperatorRewards(isMergeUpdateDeployed bool, claimIntervalTime time.Duration) (*big.Int, error) {
	if !isMergeUpdateDeployed {
		return getBigIntFrom(rp.API, "rocketRewardsPool", "getClaimingContractAllowance", "rocketClaimNode")
	}

	inflationInterval, err := tokens.GetRPLInflationIntervalRate(rp.API, nil)
	if err != nil {
		return nil, err
	}

	totalRplSupply, err := tokens.GetRPLTotalSupply(rp.API, nil)
	if err != nil {
		return nil, err
	}

	nodeOperatorRewardsPercentRaw, err := rewards.GetNodeOperatorRewardsPercent(rp.API, nil)
	if err != nil {
		return nil, err
	}

	nodeOperatorRewardsPercent := eth.WeiToEth(nodeOperatorRewardsPercentRaw)
	rewardsIntervalDays := claimIntervalTime.Seconds() / (60 * 60 * 24)
	inflationPerDay := eth.WeiToEth(inflationInterval)
	totalRplAtNextCheckpoint := (math.Pow(inflationPerDay, rewardsIntervalDays) - 1) * eth.WeiToEth(totalRplSupply)
	if totalRplAtNextCheckpoint < 0 {
		totalRplAtNextCheckpoint = 0
	}

	return eth.EthToWei(totalRplAtNextCheckpoint * nodeOperatorRewardsPercent), nil
}

func (rp *RocketpoolExporter) SaveNetworkStats() error {
	return db.SaveRocketPoolNetworkStats(
		rp.NetworkStats.RPLPrice.String(),
		rp.NetworkStats.ClaimIntervalTime.String(),
		rp.NetworkStats.ClaimIntervalTimeStart,
		rp.NetworkStats.CurrentNodeFee,
		rp.NetworkStats.CurrentNodeDemand.String(),
		rp.NetworkStats.RETHSupply.String(),
		rp.NetworkStats.NodeOperatorRewards.String(),
		rp.NetworkStats.RETHPrice,
		len(rp.NodesByAddress),
		len(rp.MinipoolsByAddress),
		len(rp.DAOMembersByAddress),
		rp.NetworkStats.TotalEthStaking.String(),
		rp.NetworkStats.TotalEthBalance.String(),
	)
}

func getBigIntFrom(rp *rocketpool.RocketPool, contract string, method string, args ...interface{}) (*big.Int, error) {
	rocketRewardsPool, err := rp.GetContract(contract, nil)
	if err != nil {
		return nil, err
	}
	perc := new(*big.Int)
	if err = rocketRewardsPool.Call(nil, perc, method, args...); err != nil {
		return nil, err
	}
	return *perc, err
}
