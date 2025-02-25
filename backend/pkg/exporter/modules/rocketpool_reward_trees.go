package modules

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/klauspost/compress/zstd"
	"github.com/pkg/errors"
	"github.com/rocket-pool/rocketpool-go/rewards"
	smartnodeCfg "github.com/rocket-pool/smartnode/shared/services/config"
	smartnodeRewards "github.com/rocket-pool/smartnode/shared/services/rewards"
)

type RocketpoolRewards struct {
	RplColl          *big.Int
	SmoothingPoolEth *big.Int
	OdaoRpl          *big.Int
}

type RocketpoolRewardTreeDownloadable struct {
	ID   uint64
	Data []byte
}

func (rp *RocketpoolExporter) DownloadMissingRewardTrees() error {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.InfoWithFields(log.Fields{"duration": time.Since(timeStart)}, "updated rocketpool-reward-trees")
	}(timeStart)

	isMergeUpdateDeployed, err := IsMergeUpdateDeployed(rp.API)
	if err != nil {
		return err
	}

	if !isMergeUpdateDeployed {
		return nil
	}

	missingIntervals, err := rp.findMissingRewardIntervals()
	if err != nil {
		return err
	}

	if len(missingIntervals) == 0 {
		return nil
	}

	log.Infof("downloading %v reward trees", len(missingIntervals))

	return rp.downloadAndValidateRewardTrees(missingIntervals)
}

func (rp *RocketpoolExporter) findMissingRewardIntervals() ([]rewards.RewardsEvent, error) {
	missingIntervals := []rewards.RewardsEvent{}
	for interval := rp.LastRewardTree; ; interval++ {
		event, err := smartnodeRewards.GetRewardSnapshotEvent(
			rp.API,
			&smartnodeCfg.RocketPoolConfig{
				Smartnode:    RP_CONFIG,
				IsNativeMode: true,
			},
			interval,
			nil,
		)
		if err != nil {
			if strings.Contains(err.Error(), "found") {
				break
			}
			return nil, fmt.Errorf("error retrieving reward tree for interval %v: %w", interval, err)
		}

		if _, exists := rp.RocketpoolRewardTreeData[event.Index.Uint64()]; !exists {
			missingIntervals = append(missingIntervals, event)
		} else {
			rp.LastRewardTree = interval + 1
		}
	}

	return missingIntervals, nil
}

func (rp *RocketpoolExporter) downloadAndValidateRewardTrees(missingIntervals []rewards.RewardsEvent) error {
	for _, missingInterval := range missingIntervals {
		if contains(rp.RocketpoolRewardTreesDownloadQueue, missingInterval.Index.Uint64()) {
			continue
		}

		bytes, err := DownloadRewardsFile(fmt.Sprintf("rp-rewards-%v-%v.json", utils.Config.Chain.Name, missingInterval.Index), missingInterval.Index.Uint64(), missingInterval.MerkleTreeCID, true, nil)
		if err != nil {
			return fmt.Errorf("can not download reward file %v: %w", missingInterval.Index, err)
		}

		proofWrapper, err := getRewardsData(bytes)
		if err != nil {
			return fmt.Errorf("can not parse reward file %v: %w", missingInterval.Index, err)
		}

		merkleRootFromFile := common.HexToHash(proofWrapper.MerkleRoot)
		if missingInterval.MerkleRoot != merkleRootFromFile {
			return fmt.Errorf("invalid merkle root value: %s != %s", missingInterval.MerkleRoot, merkleRootFromFile)
		}

		rp.RocketpoolRewardTreesDownloadQueue = append(rp.RocketpoolRewardTreesDownloadQueue, RocketpoolRewardTreeDownloadable{
			ID:   missingInterval.Index.Uint64(),
			Data: bytes,
		})

		log.Infof("Downloaded rocketpool reward tree %v", missingInterval.Index)

		if missingInterval.Index.Uint64() > rp.LastRewardTree {
			rp.LastRewardTree = missingInterval.Index.Uint64()
		}
	}

	return nil
}

func getRewardsData(jsonData []byte) (RewardsFile, error) {
	var proofWrapper RewardsFile
	err := json.Unmarshal(jsonData, &proofWrapper)
	if err != nil {
		return proofWrapper, fmt.Errorf("error deserializing reward data: %w", err)
	}

	return proofWrapper, err
}

func contains(rewardTree []RocketpoolRewardTreeDownloadable, index uint64) bool {
	for _, reward := range rewardTree {
		if reward.ID == index {
			return true
		}
	}
	return false
}

// Total cumulative rewards for an interval
type TotalRewards struct {
	ProtocolDaoRpl               *QuotedBigInt `json:"protocolDaoRpl"`
	TotalCollateralRpl           *QuotedBigInt `json:"totalCollateralRpl"`
	TotalOracleDaoRpl            *QuotedBigInt `json:"totalOracleDaoRpl"`
	TotalSmoothingPoolEth        *QuotedBigInt `json:"totalSmoothingPoolEth"`
	PoolStakerSmoothingPoolEth   *QuotedBigInt `json:"poolStakerSmoothingPoolEth"`
	NodeOperatorSmoothingPoolEth *QuotedBigInt `json:"nodeOperatorSmoothingPoolEth"`
}

// Rewards per network
type NetworkRewardsInfo struct {
	CollateralRpl    *QuotedBigInt `json:"collateralRpl"`
	OracleDaoRpl     *QuotedBigInt `json:"oracleDaoRpl"`
	SmoothingPoolEth *QuotedBigInt `json:"smoothingPoolEth"`
}

// JSON struct for a complete rewards file
type RewardsFile struct {
	// Serialized fields
	RewardsFileVersion         uint64                              `json:"rewardsFileVersion"`
	Index                      uint64                              `json:"index"`
	Network                    string                              `json:"network"`
	StartTime                  time.Time                           `json:"startTime,omitempty"`
	EndTime                    time.Time                           `json:"endTime"`
	ConsensusStartBlock        uint64                              `json:"consensusStartBlock,omitempty"`
	ConsensusEndBlock          uint64                              `json:"consensusEndBlock"`
	ExecutionStartBlock        uint64                              `json:"executionStartBlock,omitempty"`
	ExecutionEndBlock          uint64                              `json:"executionEndBlock"`
	IntervalsPassed            uint64                              `json:"intervalsPassed"`
	MerkleRoot                 string                              `json:"merkleRoot,omitempty"`
	MinipoolPerformanceFileCID string                              `json:"minipoolPerformanceFileCid,omitempty"`
	TotalRewards               *TotalRewards                       `json:"totalRewards"`
	NetworkRewards             map[uint64]*NetworkRewardsInfo      `json:"networkRewards"`
	NodeRewards                map[common.Address]*NodeRewardsInfo `json:"nodeRewards"`
	MinipoolPerformanceFile    MinipoolPerformanceFile             `json:"-"`
}

func DownloadRewardsFile(fileName string, interval uint64, cid string, isDaemon bool, client *http.Client) ([]byte, error) {
	ipfsFilename := fileName + ".zst"

	split := strings.Split(fileName, "-")
	var network string
	if len(split) > 3 {
		network = split[2]
	}

	if client == nil {
		client = &http.Client{
			Timeout: 40 * time.Second,
		}
	}

	// Create URL list
	urls := []string{
		fmt.Sprintf("https://%s.ipfs.dweb.link/%s", cid, ipfsFilename),
		fmt.Sprintf("https://ipfs.io/ipfs/%s/%s", cid, ipfsFilename),
		fmt.Sprintf("https://github.com/rocket-pool/rewards-trees/raw/main/%s/%s", network, fileName),
	}

	// Attempt downloads
	errBuilder := strings.Builder{}
	for _, url := range urls {
		resp, err := client.Get(url)
		if err != nil {
			errBuilder.WriteString(fmt.Sprintf("Downloading %s failed (%s)\n", url, err.Error()))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errBuilder.WriteString(fmt.Sprintf("Downloading %s failed with status %s\n", url, resp.Status))
			continue
		} else {
			// If we got here, we have a successful download
			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				errBuilder.WriteString(fmt.Sprintf("Error reading response bytes from %s: %s\n", url, err.Error()))
				continue
			}

			// Decompress it
			writeBytes := bytes
			if strings.HasSuffix(url, ".zst") {
				writeBytes, err = decompressFile(bytes)
				if err != nil {
					errBuilder.WriteString(fmt.Sprintf("Error decompressing %s: %s\n", url, err.Error()))
					continue
				}
			}

			return writeBytes, nil
		}
	}

	return nil, errors.New(errBuilder.String())
}

// Decompresses a rewards file
func decompressFile(compressedBytes []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating compression decoder: %w", err)
	}

	decompressedBytes, err := decoder.DecodeAll(compressedBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("error decompressing rewards file: %w", err)
	}

	return decompressedBytes, nil
}

func (rp *RocketpoolExporter) SaveRewardTrees() error {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.InfoWithFields(log.Fields{"duration": time.Since(timeStart)}, "saved rocketpool reward trees")
	}(timeStart)

	if len(rp.RocketpoolRewardTreesDownloadQueue) == 0 {
		return nil
	}

	log.Infof("saving %v rocketpool reward trees", len(rp.RocketpoolRewardTreesDownloadQueue))

	if err := rp.saveRewardTrees(); err != nil {
		return err
	}

	if err := rp.refreshMaterializedView(); err != nil {
		return err
	}

	var err error
	rp.RocketpoolRewardTreeData, err = rp.getRocketpoolRewardTrees()
	if err != nil {
		return err
	}

	// Delete download queue after refreshing the trees from db in case
	// refreshing throws an error so we try again in the next iteration
	// and always have an up to date tree
	rp.RocketpoolRewardTreesDownloadQueue = []RocketpoolRewardTreeDownloadable{}

	return nil
}

func (rp *RocketpoolExporter) getRocketpoolRewardTrees() (map[uint64]RewardsFile, error) {
	allRewards := make(map[uint64]RewardsFile)

	log.Infof("rocketpool refreshing all reward tree data...")

	jsonData, err := db.GetRocketPoolRewardTrees()
	if err != nil {
		return nil, fmt.Errorf("error while getting claimedInterval tree from database, is it exported? %v", err)
	}

	for _, data := range jsonData {
		allRewards[data.ID], err = getRewardsData(data.Data)
		if err != nil {
			return nil, fmt.Errorf("error while parsing reward tree data to struct for interval %v, error: %w", data.ID, err)
		}
	}

	return allRewards, nil
}

func (rp *RocketpoolExporter) saveRewardTrees() error {
	for _, rewardTree := range rp.RocketpoolRewardTreesDownloadQueue {
		err := db.SaveRocketPoolRewardTree(rewardTree.ID, rewardTree.Data)
		if err != nil {
			return fmt.Errorf("can not store reward file %v. Error %w", rewardTree.ID, err)
		}
	}
	return nil
}

func (rp *RocketpoolExporter) refreshMaterializedView() error {
	exists, err := db.CheckRocketPoolMVExists()
	if err != nil {
		return fmt.Errorf("failed to check if materialized view exists: %w", err)
	}

	// If the view exists, refresh it concurrently
	if exists {
		err = db.RefreshRocketPoolMV()
		if err != nil {
			return fmt.Errorf("cannot refresh materialized view rocketpool_rewards_summary. Error %w", err)
		}
	} else {
		log.Infof("Materialized view rocketpool_rewards_summary does not exist, skipping refresh.")
	}

	return nil
}
