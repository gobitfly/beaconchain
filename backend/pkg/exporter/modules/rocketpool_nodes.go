package modules

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	smartnodeRewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"golang.org/x/sync/errgroup"
)

type RocketpoolNode struct {
	Address                []byte   `db:"address"`
	TimezoneLocation       string   `db:"timezone_location"`
	RPLStake               *big.Int `db:"rpl_stake"`
	EffectiveRPLStake      *big.Int `db:"effective_rpl_stake"`
	MinRPLStake            *big.Int `db:"min_rpl_stake"`
	MaxRPLStake            *big.Int `db:"max_rpl_stake"`
	RPLCumulativeRewards   *big.Int `db:"rpl_cumulative_rewards"`
	SmoothingPoolOptedIn   bool     `db:"smoothing_pool_opted_in"`
	ClaimedSmoothingPool   *big.Int `db:"claimed_smoothing_pool"`
	UnclaimedSmoothingPool *big.Int `db:"unclaimed_smoothing_pool"`
	UnclaimedRPLRewards    *big.Int `db:"unclaimed_rpl_rewards"`
	DepositCredit          *big.Int `db:"deposit_credit"`
}

// Node operator rewards
type NodeRewardsInfo struct {
	RewardNetwork                uint64        `json:"rewardNetwork"`
	CollateralRpl                *QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl                 *QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth             *QuotedBigInt `json:"smoothingPoolEth"`
	SmoothingPoolEligibilityRate float64       `json:"smoothingPoolEligibilityRate"`
	MerkleData                   []byte        `json:"-"`
	MerkleProof                  []string      `json:"merkleProof"`
}

func NewRocketpoolNode(rp *rocketpool.RocketPool, addr []byte, rewardTrees map[uint64]RewardsFile, legacyClaims map[string]*big.Int, atlasDeployed bool) (*RocketpoolNode, error) {
	rpn := &RocketpoolNode{
		Address: addr,
	}

	err := rpn.Update(rp, rewardTrees, true, legacyClaims, atlasDeployed)
	if err != nil {
		return nil, err
	}

	return rpn, nil
}

func (rp *RocketpoolExporter) SaveNodes() error {
	if len(rp.NodesByAddress) == 0 {
		return nil
	}

	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.DebugWithFields(log.Fields{"duration": time.Since(timeStart)}, "saved rocketpool-nodes")
	}(timeStart)

	data := rp.prepareNodeData()
	return rp.saveNodeData(data)
}

func (rp *RocketpoolExporter) prepareNodeData() []*RocketpoolNode {
	data := make([]*RocketpoolNode, len(rp.NodesByAddress))
	i := 0
	for _, node := range rp.NodesByAddress {
		data[i] = node
		i++
	}

	return data
}

func (rp *RocketpoolExporter) saveNodeData(data []*RocketpoolNode) error {
	nArgs := 13
	valueStringsTpl := createValueStringsTemplate(nArgs)
	batchSize := 1000

	for b := 0; b < len(data); b += batchSize {
		start := b
		end := b + batchSize
		if len(data) < end {
			end = len(data)
		}

		valueStrings, valueArgs := rp.prepareNodeBatch(data[start:end], valueStringsTpl, nArgs)
		if err := db.SaveRocketPoolNodes(valueStrings, valueArgs); err != nil {
			return fmt.Errorf("error inserting into rocketpool_nodes table: %w", err)
		}
	}

	return nil
}

func (rp *RocketpoolExporter) prepareNodeBatch(data []*RocketpoolNode, valueStringsTpl string, nArgs int) ([]string, []interface{}) {
	valueStringsArgs := make([]interface{}, nArgs)
	valueStrings := make([]string, 0, len(data))
	valueArgs := make([]interface{}, 0, len(data)*nArgs)

	for i, d := range data {
		for j := 0; j < nArgs; j++ {
			valueStringsArgs[j] = i*nArgs + j + 1
		}
		valueStrings = append(valueStrings, fmt.Sprintf(valueStringsTpl, valueStringsArgs...))
		valueArgs = append(valueArgs, rp.API.RocketStorageContract.Address.Bytes())
		valueArgs = append(valueArgs, d.Address)
		valueArgs = append(valueArgs, d.TimezoneLocation)
		valueArgs = append(valueArgs, d.RPLStake.String())
		valueArgs = append(valueArgs, d.MinRPLStake.String())
		valueArgs = append(valueArgs, d.MaxRPLStake.String())
		valueArgs = append(valueArgs, d.RPLCumulativeRewards.String())
		valueArgs = append(valueArgs, d.SmoothingPoolOptedIn)
		valueArgs = append(valueArgs, d.ClaimedSmoothingPool.String())
		valueArgs = append(valueArgs, d.UnclaimedSmoothingPool.String())
		valueArgs = append(valueArgs, d.UnclaimedRPLRewards.String())
		valueArgs = append(valueArgs, d.EffectiveRPLStake.String())
		valueArgs = append(valueArgs, d.DepositCredit.String())
	}

	return valueStrings, valueArgs
}

func (rp *RocketpoolExporter) UpdateNodes(includeCumulativeRpl bool) error {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.DebugWithFields(log.Fields{"duration": time.Since(timeStart)}, "updated rocketpool-nodes")
	}(timeStart)

	nodeAddresses, err := node.GetNodeAddresses(rp.API, nil)
	if err != nil {
		return err
	}

	atlasDeployed, err := IsAtlasDeployed(rp.API)
	if err != nil {
		return err
	}

	if includeCumulativeRpl {
		if err := rp.calculateCumulativeRPL(); err != nil {
			return err
		}
	}

	return rp.updateNodes(nodeAddresses, includeCumulativeRpl, atlasDeployed)
}

func (rp *RocketpoolExporter) calculateCumulativeRPL() error {
	legacyRewardsPool := RP_CONFIG.GetV100RewardsPoolAddress()
	legacyClaimNode := RP_CONFIG.GetV100ClaimNodeAddress()
	var err error
	rp.NodeRPLCumulative, err = CalculateLifetimeNodeRewardsAllLegacy(
		rp.API,
		big.NewInt(GethEventLogInterval),
		&legacyRewardsPool,
		&legacyClaimNode,
	)
	return err
}

func (rp *RocketpoolExporter) updateNodes(nodeAddresses []common.Address, includeCumulativeRpl bool, atlasDeployed bool) error {
	for _, a := range nodeAddresses {
		addrHex := a.Hex()
		if node, exists := rp.NodesByAddress[addrHex]; exists {
			if err := node.Update(rp.API, rp.RocketpoolRewardTreeData, includeCumulativeRpl, rp.NodeRPLCumulative, atlasDeployed); err != nil {
				return err
			}
			continue
		}
		node, err := NewRocketpoolNode(rp.API, a.Bytes(), rp.RocketpoolRewardTreeData, rp.NodeRPLCumulative, atlasDeployed)
		if err != nil {
			return err
		}
		rp.NodesByAddress[addrHex] = node
	}

	return nil
}

type RocketpoolNodeDetails struct {
	TimezoneLocation  string
	RPLStake          *big.Int
	MinRPLStake       *big.Int
	MaxRPLStake       *big.Int
	EffectiveRPLStake *big.Int
	DepositCredit     *big.Int
}

func (r *RocketpoolNode) Update(rp *rocketpool.RocketPool, rewardTrees map[uint64]RewardsFile, includeCumulativeRpl bool, legacyClaims map[string]*big.Int, atlasDeployed bool) error {
	address := common.BytesToAddress(r.Address)

	nodeDetails, err := r.fetchNodeDetails(rp, address, atlasDeployed)
	if err != nil {
		return err
	}

	// handle reward trees if available
	if len(rewardTrees) > 0 {
		if err := r.updateRewards(rp, address, rewardTrees, includeCumulativeRpl, legacyClaims); err != nil {
			return err
		}
	}

	r.initializeRewardsFields()

	// update node fields
	r.TimezoneLocation = nodeDetails.TimezoneLocation
	r.RPLStake = nodeDetails.RPLStake
	r.MinRPLStake = nodeDetails.MinRPLStake
	r.MaxRPLStake = nodeDetails.MaxRPLStake
	r.EffectiveRPLStake = nodeDetails.EffectiveRPLStake
	r.DepositCredit = nodeDetails.DepositCredit

	return nil
}

func (r *RocketpoolNode) updateRewards(rp *rocketpool.RocketPool, address common.Address, rewardTrees map[uint64]RewardsFile, includeCumulativeRpl bool, legacyClaims map[string]*big.Int) error {
	var err error
	r.SmoothingPoolOptedIn, err = node.GetSmoothingPoolRegistrationState(rp, address, nil)
	if err != nil {
		return err
	}

	if includeCumulativeRpl {
		claimedSum, unclaimedSum, err := r.calculateRewards(rp, address, rewardTrees)
		if err != nil {
			return err
		}

		r.RPLCumulativeRewards = claimedSum.RplColl
		if legacyAmount, exists := legacyClaims[address.Hex()]; exists {
			r.RPLCumulativeRewards = r.RPLCumulativeRewards.Add(r.RPLCumulativeRewards, legacyAmount)
		}
		r.ClaimedSmoothingPool = claimedSum.SmoothingPoolEth
		r.UnclaimedSmoothingPool = unclaimedSum.SmoothingPoolEth
		r.UnclaimedRPLRewards = unclaimedSum.RplColl
	}

	return nil
}

func (r *RocketpoolNode) calculateRewards(rp *rocketpool.RocketPool, address common.Address, rewardTrees map[uint64]RewardsFile) (RocketpoolRewards, RocketpoolRewards, error) {
	claimedSum := RocketpoolRewards{
		SmoothingPoolEth: big.NewInt(0),
		OdaoRpl:          big.NewInt(0),
		RplColl:          big.NewInt(0),
	}
	unclaimedSum := RocketpoolRewards{
		SmoothingPoolEth: big.NewInt(0),
		OdaoRpl:          big.NewInt(0),
		RplColl:          big.NewInt(0),
	}

	unclaimed, claimed, err := smartnodeRewards.GetClaimStatus(rp, address)
	if err != nil {
		return claimedSum, unclaimedSum, err
	}

	// Get the info for each claimed interval
	for _, claimedInterval := range claimed {
		rewardData := rewardTrees[claimedInterval]
		if rewards, exists := rewardData.NodeRewards[address]; exists {
			claimedSum.RplColl = claimedSum.RplColl.Add(claimedSum.RplColl, &rewards.CollateralRpl.Int)
			claimedSum.SmoothingPoolEth = claimedSum.SmoothingPoolEth.Add(claimedSum.SmoothingPoolEth, &rewards.SmoothingPoolEth.Int)
			claimedSum.OdaoRpl = claimedSum.OdaoRpl.Add(claimedSum.OdaoRpl, &rewards.OracleDaoRpl.Int)
		}
	}

	// Get the unclaimed rewards
	for _, unclaimedInterval := range unclaimed {
		rewardData := rewardTrees[unclaimedInterval]
		if rewards, exists := rewardData.NodeRewards[address]; exists {
			unclaimedSum.RplColl = unclaimedSum.RplColl.Add(unclaimedSum.RplColl, &rewards.CollateralRpl.Int)
			unclaimedSum.SmoothingPoolEth = unclaimedSum.SmoothingPoolEth.Add(unclaimedSum.SmoothingPoolEth, &rewards.SmoothingPoolEth.Int)
			unclaimedSum.OdaoRpl = unclaimedSum.OdaoRpl.Add(unclaimedSum.OdaoRpl, &rewards.OracleDaoRpl.Int)
		}
	}

	return claimedSum, unclaimedSum, nil
}

func (r *RocketpoolNode) initializeRewardsFields() {
	if r.RPLCumulativeRewards == nil {
		r.RPLCumulativeRewards = big.NewInt(0)
	}
	if r.UnclaimedRPLRewards == nil {
		r.UnclaimedRPLRewards = big.NewInt(0)
	}
	if r.UnclaimedSmoothingPool == nil {
		r.UnclaimedSmoothingPool = big.NewInt(0)
	}
	if r.ClaimedSmoothingPool == nil {
		r.ClaimedSmoothingPool = big.NewInt(0)
	}
}

func (r *RocketpoolNode) fetchNodeDetails(rp *rocketpool.RocketPool, address common.Address, atlasDeployed bool) (*RocketpoolNodeDetails, error) {
	var wg errgroup.Group
	var details RocketpoolNodeDetails

	wg.Go(func() error {
		var err error
		details.TimezoneLocation, err = node.GetNodeTimezoneLocation(rp, address, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		details.RPLStake, err = node.GetNodeRPLStake(rp, address, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		details.MinRPLStake, err = node.GetNodeMinimumRPLStake(rp, address, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		details.MaxRPLStake, err = node.GetNodeMaximumRPLStake(rp, address, nil)
		return err
	})

	wg.Go(func() error {
		var err error
		details.EffectiveRPLStake, err = node.GetNodeEffectiveRPLStake(rp, address, nil)
		return err
	})

	if atlasDeployed {
		wg.Go(func() error {
			var err error
			details.DepositCredit, err = node.GetNodeDepositCredit(rp, address, nil)
			return err
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return &details, nil
}

func CalculateLifetimeNodeRewardsAllLegacy(rp *rocketpool.RocketPool, intervalSize *big.Int, legacyRocketRewardsPoolAddress *common.Address, legacyRocketClaimNodeAddress *common.Address) (map[string]*big.Int, error) {
	// Get contracts
	rocketRewardsPool, err := getRocketRewardsPoolLegacy(rp, legacyRocketRewardsPoolAddress)
	if err != nil {
		return nil, err
	}
	rocketClaimNode, err := getRocketClaimNodeLegacy(rp, legacyRocketClaimNodeAddress)
	if err != nil {
		return nil, err
	}

	maxBlockNumber := getMaxBlockNumberForLegacyRewards()

	if maxBlockNumber == nil {
		return make(map[string]*big.Int), nil
	}

	logs, err := getRPLTokensClaimedLogs(rp, rocketRewardsPool, rocketClaimNode, intervalSize, maxBlockNumber)
	if err != nil {
		return nil, fmt.Errorf("cannot load lifetime rewards: %w", err)
	}

	sumRewards := sumRewardsFromLogs(logs, rocketRewardsPool)

	return sumRewards, nil
}

func getMaxBlockNumberForLegacyRewards() *big.Int {
	prerecordedIntervals, exists := firstBlockOfRedstone[utils.Config.Chain.Name]
	if prerecordedIntervals == 0 || !exists {
		return nil
	}
	// only look for legacy lifetime rewards before the new rewards system went live
	return big.NewInt(0).SetUint64(prerecordedIntervals)
}

func getRPLTokensClaimedLogs(rp *rocketpool.RocketPool, rocketRewardsPool, rocketClaimNode *rocketpool.Contract, intervalSize, maxBlockNumber *big.Int) ([]types.Log, error) {
	// Construct a filter query for relevant logs
	addressFilter := []common.Address{*rocketRewardsPool.Address}
	// RPLTokensClaimed(address clamingContract, address claimingAddress, uint256 amount, uint256 time)
	topicFilter := [][]common.Hash{
		{rocketRewardsPool.ABI.Events["RPLTokensClaimed"].ID},
		{common.BytesToHash(rocketClaimNode.Address[:])},
	}

	logs, err := eth.GetLogs(rp, addressFilter, topicFilter, intervalSize, nil, maxBlockNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("can not load lifetime rewards: %w", err)
	}

	return logs, nil
}

func sumRewardsFromLogs(logs []types.Log, rocketRewardsPool *rocketpool.Contract) map[string]*big.Int {
	sumRewards := make(map[string]*big.Int)

	// Iterate over the logs and sum the amounts
	for _, log := range logs {
		values := make(map[string]interface{})
		// Decode the event
		if err := rocketRewardsPool.ABI.Events["RPLTokensClaimed"].Inputs.UnpackIntoMap(values, log.Data); err != nil {
			continue // Skip logs that cannot be decoded
		}

		// Add the amount argument to our sum
		amount := values["amount"].(*big.Int)
		claimAddress := common.BytesToAddress(log.Topics[2].Bytes())
		sum, ok := sumRewards[claimAddress.Hex()]
		if !ok {
			sum = big.NewInt(0)
		}
		sumRewards[claimAddress.Hex()] = sum.Add(sum, amount)
	}

	return sumRewards
}

// Get contracts
var rocketRewardsPoolLock sync.Mutex

func getRocketRewardsPoolLegacy(rp *rocketpool.RocketPool, address *common.Address) (*rocketpool.Contract, error) {
	rocketRewardsPoolLock.Lock()
	defer rocketRewardsPoolLock.Unlock()
	if address == nil {
		return rp.VersionManager.V1_0_0.GetContract("rocketRewardsPool", nil)
	} else {
		return rp.VersionManager.V1_0_0.GetContractWithAddress("rocketRewardsPool", *address)
	}
}

// Get contracts
var rocketClaimNodeLock sync.Mutex

func getRocketClaimNodeLegacy(rp *rocketpool.RocketPool, address *common.Address) (*rocketpool.Contract, error) {
	rocketClaimNodeLock.Lock()
	defer rocketClaimNodeLock.Unlock()
	if address == nil {
		return rp.VersionManager.V1_0_0.GetContract("rocketClaimNode", nil)
	} else {
		return rp.VersionManager.V1_0_0.GetContractWithAddress("rocketClaimNode", *address)
	}
}
