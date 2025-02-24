package modules

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/hashicorp/go-version"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rputil "github.com/rocket-pool/rocketpool-go/utils"
	smartnodeCfg "github.com/rocket-pool/smartnode/shared/services/config"
	smartnodeNetwork "github.com/rocket-pool/smartnode/shared/types/config"
	"golang.org/x/sync/errgroup"
)

var rpEth1RPRCClient *gethrpc.Client
var rpEth1Client *ethclient.Client

const GethEventLogInterval = 25000

var RP_CONFIG *smartnodeCfg.SmartnodeConfig
var firstBlockOfRedstone = map[string]uint64{
	"mainnet": 15451165,
	"prater":  7287326,
	"holesky": 0,
}
var leb16, _ = big.NewInt(0).SetString("16000000000000000000", 10)

func rocketpoolExporter() {
	RP_CONFIG = initRPConfig()
	endpoint := getWebSocketEndpoint(utils.Config.Eth1GethEndpoint)

	var err error
	rpEth1RPRCClient, err = gethrpc.Dial(endpoint)
	if err != nil {
		log.Fatal(err, "new rocketpool geth client error", 0)
	}

	rpEth1Client = ethclient.NewClient(rpEth1RPRCClient)

	rpExporter, err := createRocketPoolExporter(
		rpEth1Client,
		RP_CONFIG.GetStorageAddress(),
	)
	if err != nil {
		log.Fatal(err, "new rocketpool exporter error", 0)
	}

	err = rpExporter.Run()
	if err != nil {
		log.Error(err, "rocketpool exporter run error", 0)
	}
}

func getWebSocketEndpoint(endpoint string) string {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return "ws" + endpoint[4:]
	}
	return endpoint
}

func initRPConfig() *smartnodeCfg.SmartnodeConfig {
	config := smartnodeCfg.NewSmartnodeConfig(&smartnodeCfg.RocketPoolConfig{
		RocketPoolDirectory: "/tmp/rocketpool",
	})

	switch utils.Config.Chain.Name {
	case "mainnet":
		config.Network.Value = smartnodeNetwork.Network_Mainnet
	case "holesky":
		config.Network.Value = smartnodeNetwork.Network_Holesky
	default:
		log.Warnf("unknown network: %s", utils.Config.Chain.Name)
	}

	return config
}

type RocketpoolExporter struct {
	Eth1Client                         *ethclient.Client
	API                                *rocketpool.RocketPool
	UpdateInterval                     time.Duration
	MinipoolsByAddress                 map[string]*RocketpoolMinipool
	NodesByAddress                     map[string]*RocketpoolNode
	DAOProposalsByID                   map[uint64]*RocketpoolDAOProposal
	DAOMembersByAddress                map[string]*RocketpoolDAOMember
	NodeRPLCumulative                  map[string]*big.Int
	NetworkStats                       RocketpoolNetworkStats
	LastRewardTree                     uint64
	RocketpoolRewardTreesDownloadQueue []RocketpoolRewardTreeDownloadable
	RocketpoolRewardTreeData           map[uint64]RewardsFile
	OnchainConfig                      *RocketPoolOnchainConfig
}

type RocketPoolOnchainConfig struct {
	SmoothingPoolAddress common.Address
}

func createRocketPoolExporter(eth1Client *ethclient.Client, storageContractAddressHex string) (*RocketpoolExporter, error) {
	rp, err := rocketpool.NewRocketPool(eth1Client, common.HexToAddress(storageContractAddressHex))
	if err != nil {
		return nil, fmt.Errorf("failed to create RocketPool instance: %w", err)
	}

	return &RocketpoolExporter{
		Eth1Client:                         eth1Client,
		API:                                rp,
		UpdateInterval:                     time.Minute,
		MinipoolsByAddress:                 make(map[string]*RocketpoolMinipool),
		NodesByAddress:                     make(map[string]*RocketpoolNode),
		DAOProposalsByID:                   make(map[uint64]*RocketpoolDAOProposal),
		DAOMembersByAddress:                make(map[string]*RocketpoolDAOMember),
		NodeRPLCumulative:                  make(map[string]*big.Int),
		LastRewardTree:                     0,
		RocketpoolRewardTreesDownloadQueue: []RocketpoolRewardTreeDownloadable{},
		RocketpoolRewardTreeData:           make(map[uint64]RewardsFile),
		OnchainConfig: &RocketPoolOnchainConfig{
			SmoothingPoolAddress: common.HexToAddress(storageContractAddressHex),
		},
	}, nil
}

func (rp *RocketpoolExporter) Run() error {
	errorInterval := time.Minute
	t := time.NewTicker(rp.UpdateInterval)
	defer t.Stop()
	var count int64 = 0

	isMergeUpdateDeployed, err := IsMergeUpdateDeployed(rp.API)
	if err != nil {
		log.Error(err, "error retrieving rocketpool redstone deploy status", 0)
		return err
	}

	if isMergeUpdateDeployed {
		rp.RocketpoolRewardTreeData, err = rp.getRocketpoolRewardTrees()
		if err != nil {
			log.Error(err, "error retrieving known rocketpool reward tree data from db", 0)
			return err
		}

		for _, data := range rp.RocketpoolRewardTreeData {
			if data.Index > rp.LastRewardTree {
				rp.LastRewardTree = data.Index
			}
		}
	}

	log.Infof("rocketpool exporter initialized")

	for {
		timeStart := time.Now()
		// TODO: re-enable status reports after this thing is more stable
		//r := monitoringServices.NewStatusReport(constants.Event_ExporterLegacyRocketPool, time.Hour*4, rp.UpdateInterval) // currently takes 2h40m on mainnet...
		//r(constants.Running, nil)
		err := rp.updateAndSaveRocketpoolData(count)
		if err != nil {
			log.Error(err, "error updating or saving rocketpool-data", 0)
			time.Sleep(errorInterval)
			continue
		}

		services.ReportStatus("rocketpoolExporter", "Running", nil)

		metrics.TaskDuration.WithLabelValues("exporter_rocketpoolExporter").Observe(time.Since(timeStart).Seconds())
		//r(constants.Success, map[string]string{"took": time.Since(t0).String(), "took_raw": fmt.Sprintf("%v", time.Since(t0).Milliseconds())})

		log.InfoWithFields(log.Fields{"duration": time.Since(timeStart)}, "exported rocketpool-data")
		count++
		<-t.C
	}
}

func (rp *RocketpoolExporter) updateAndSaveRocketpoolData(count int64) error {
	if err := rp.Update(count); err != nil {
		//r(constants.Failure, map[string]string{"error": err.Error()})
		return fmt.Errorf("error updating rocketpool-data: %w", err)
	}

	if err := rp.Save(count); err != nil {
		//r(constants.Failure, map[string]string{"error": err.Error()})
		return fmt.Errorf("error saving rocketpool-data: %w", err)
	}

	return nil
}

func (rp *RocketpoolExporter) Update(count int64) error {
	var wg errgroup.Group
	wg.Go(func() error { return rp.DownloadMissingRewardTrees() })
	wg.Go(func() error { return rp.UpdateMinipools() })
	wg.Go(func() error { return rp.UpdateNodes(true) })
	wg.Go(func() error { return rp.UpdateDAOProposals() })
	wg.Go(func() error { return rp.UpdateDAOMembers() })
	wg.Go(func() error { return rp.UpdateNetworkStats() })
	wg.Go(func() error { return rp.UpdateConfigs() })

	return wg.Wait()
}

func (rp *RocketpoolExporter) Save(count int64) error {
	var err error
	err = rp.SaveMinipools()
	if err != nil {
		return err
	}
	err = rp.SaveNodes()
	if err != nil {
		return err
	}
	err = rp.SaveDAOProposals()
	if err != nil {
		return err
	}
	err = rp.SaveDAOProposalsMemberVotes()
	if err != nil {
		return err
	}
	err = rp.SaveDAOMembers()
	if err != nil {
		return err
	}
	err = rp.TagValidators()
	if err != nil {
		return err
	}
	if count%5 == 0 { // smart contracts aren't updated that often, so lets save it less often
		err = rp.SaveNetworkStats()
		if err != nil {
			return err
		}
	}
	err = rp.SaveRewardTrees()
	if err != nil {
		return err
	}

	err = rp.SaveConfigs()
	if err != nil {
		return err
	}

	return nil
}

func (rp *RocketpoolExporter) UpdateConfigs() error {
	t0 := time.Now()
	defer func(t0 time.Time) {
		log.DebugWithFields(log.Fields{"duration": time.Since(t0)}, "updated rocketpool-configs")
	}(t0)

	smoothingPoolAddress, err := rp.API.GetContract("rocketSmoothingPool", nil)
	if err != nil {
		return err
	}

	rp.OnchainConfig = &RocketPoolOnchainConfig{
		SmoothingPoolAddress: *smoothingPoolAddress.Address,
	}

	return nil
}

func (rp *RocketpoolExporter) SaveConfigs() error {
	timeStart := time.Now()
	defer func(timeStart time.Time) {
		log.DebugWithFields(log.Fields{"duration": time.Since(timeStart)}, "saved rocketpool-configs")
	}(timeStart)

	storageAddress, err := hex.DecodeString(strings.Trim(RP_CONFIG.GetStorageAddress(), "0x"))
	if err != nil {
		return errors.Wrap(err, "error decoding storage address")
	}

	err = db.SaveRocketPoolConfig(storageAddress, rp.OnchainConfig.SmoothingPoolAddress.Bytes())
	if err != nil {
		return errors.Wrap(err, "error inserting into rocketpool_onchain_configs")
	}

	return nil
}

// Redstone activation check
// Credit https://github.com/rocket-pool/smartnode/blob/4fd78852a331a7ec7a7e462fef2bcd49d1f0b0af/shared/utils/rp/update-checks.go
func IsMergeUpdateDeployed(rp *rocketpool.RocketPool) (bool, error) {
	currentVersion, err := rputil.GetCurrentVersion(rp, nil)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.1.0")
	return constraint.Check(currentVersion), nil
}

func IsAtlasDeployed(rp *rocketpool.RocketPool) (bool, error) {
	currentVersion, err := rputil.GetCurrentVersion(rp, nil)
	if err != nil {
		return false, err
	}

	constraint, _ := version.NewConstraint(">= 1.2.0")
	return constraint.Check(currentVersion), nil
}

func createValueStringsTemplate(nArgs int) string {
	valueStringsArr := make([]string, nArgs)
	for i := range valueStringsArr {
		valueStringsArr[i] = "$%d"
	}

	return "(" + strings.Join(valueStringsArr, ",") + ")"
}

type QuotedBigInt struct {
	big.Int
}
