package evm_node_indexer

// imports
import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/hexutil"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gtuk/discordwebhook"
	"github.com/lib/pq"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"

	gcp_bigtable "cloud.google.com/go/bigtable"
)

// defines
const MAX_EL_BLOCK_NUMBER = int64(1_000_000_000_000 - 1)

const BT_COLUMNFAMILY_BLOCK = "b"
const BT_COLUMN_BLOCK = "b"
const BT_COLUMNFAMILY_RECEIPTS = "r"
const BT_COLUMN_RECEIPTS = "r"
const BT_COLUMNFAMILY_TRACES = "t"
const BT_COLUMN_TRACES = "t"
const BT_COLUMNFAMILY_UNCLES = "u"
const BT_COLUMN_UNCLES = "u"

const MAINNET_CHAINID = 1
const GOERLI_CHAINID = 5
const OPTIMISM_CHAINID = 10
const GNOSIS_CHAINID = 100
const HOLESKY_CHAINID = 17000
const ARBITRUM_CHAINID = 42161
const ARBITRUM_NITRO_BLOCKNUMBER = 22207815
const SEPOLIA_CHAINID = 11155111

const HTTP_TIMEOUT_IN_SECONDS = 120
const MAX_REORG_DEPTH = 256             // maxmimum value for reorg (that number of blocks we are looking 'back'), includes latest block
const MAX_NODE_REQUESTS_AT_ONCE = 1024  // maximum node requests allowed
const OUTPUT_CYCLE_IN_SECONDS = 8       // duration between 2 outputs / updates, just a visual thing
const TRY_TO_RECOVER_ON_ERROR_COUNT = 8 // total retries, so with a value of 4, it is 1 try + 4 retries

// structs
type jsonRpcReturnId struct {
	Id int64 `json:"id"`
}
type fullBlockRawData struct {
	blockNumber      int64
	blockHash        hexutil.Bytes
	blockUnclesCount int
	blockTxs         []string

	blockCompressed    hexutil.Bytes
	receiptsCompressed hexutil.Bytes
	tracesCompressed   hexutil.Bytes
	unclesCompressed   hexutil.Bytes
}
type intRange struct {
	start int64
	end   int64
}

// local globals
var currentNodeBlockNumber atomic.Int64
var elClient *ethclient.Client
var reorgDepth *int64
var httpClient *http.Client
var errorIdentifier *regexp.Regexp
var eth1RpcEndpoint string

// init
func init() {
	httpClient = &http.Client{Timeout: time.Second * HTTP_TIMEOUT_IN_SECONDS}

	var err error
	errorIdentifier, err = regexp.Compile(`\"error":\{\"code\":\-[0-9]+\,\"message\":\"([^\"]*)`)
	if err != nil {
		log.Fatal(err, "fatal, compiling regex", 0)
	}
}

// main
func Run() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)

	// read / set parameter
	configPath := fs.String("config", "config/default.config.yml", "Path to the config file")
	versionFlag := fs.Bool("version", false, "print version and exit")
	startBlockNumber := fs.Int64("start-block-number", -1, "trigger a REEXPORT, only working in combination with end-block-number, defined block is included, will be the first action done and will quite afterwards, ignore every other action")
	endBlockNumber := fs.Int64("end-block-number", -1, "trigger a REEXPORT, only working in combination with start-block-number, defined block is included, will be the first action done and will quite afterwards, ignore every other action")
	reorgDepth = fs.Int64("reorg.depth", 32, fmt.Sprintf("lookback to check and handle chain reorgs (MAX %s), you should NEVER reduce this after the first start, otherwise there will be unchecked areas", _formatInt64(MAX_REORG_DEPTH)))
	concurrency := fs.Int64("concurrency", 8, "maximum threads used (running on maximum whenever possible)")
	nodeRequestsAtOnce := fs.Int64("node-requests-at-once", 16, fmt.Sprintf("bulk size per node = bt = db request (MAX %s)", _formatInt64(MAX_NODE_REQUESTS_AT_ONCE)))
	skipHoleCheck := fs.Bool("skip-hole-check", false, "skips the initial check for holes, doesn't go very well with only-hole-check")
	onlyHoleCheck := fs.Bool("only-hole-check", false, "just check for holes and quit, can be used for a reexport running simulation to a normal setup, just remove entries in postgres and start with this flag, doesn't go very well with skip-hole-check")
	noNewBlocks := fs.Bool("ignore-new-blocks", false, "there are no new blocks, at all")
	noNewBlocksThresholdSeconds := fs.Int("fatal-if-no-new-block-for-x-seconds", 600, "will fatal if there is no new block for x seconds (MIN 30), will start throwing errors at 2/3 of the time, will start throwing warnings at 1/3 of the time, doesn't go very well with ignore-new-blocks")
	discordWebhookBlockThreshold := fs.Int64("discord-block-threshold", 100000, "every x blocks an update is send to Discord")
	discordWebhookReportUrl := fs.String("discord-url", "", "report progress to discord url")
	discordWebhookUser := fs.String("discord-user", "", "report progress to discord user")
	discordWebhookAddTextFatal := fs.String("discord-fatal-text", "", "this text will be added to the discord message in the case of an fatal")
	err := fs.Parse(os.Args[2:])
	if err != nil {
		log.Fatal(err, "error parsing flags", 0)
	}
	if *versionFlag {
		log.Info(version.Version)
		return
	}

	// tell the user about all parameter
	{
		log.Infof("config set to '%s'", *configPath)
		if *startBlockNumber >= 0 {
			log.Infof("start-block-number set to '%s'", _formatInt64(*startBlockNumber))
		}
		if *endBlockNumber >= 0 {
			log.Infof("end-block-number set to '%s'", _formatInt64(*endBlockNumber))
		}
		log.Infof("reorg.depth set to '%s'", _formatInt64(*reorgDepth))
		log.Infof("concurrency set to '%s'", _formatInt64(*concurrency))
		log.Infof("node-requests-at-once set to '%s'", _formatInt64(*nodeRequestsAtOnce))
		if *skipHoleCheck {
			log.Infof("skip-hole-check set true")
		}
		if *onlyHoleCheck {
			log.Infof("only-hole-check set true")
		}
		if *noNewBlocks {
			log.Infof("ignore-new-blocks set true")
		}
		log.Infof("fatal-if-no-new-block-for-x-seconds set to '%d' seconds", *noNewBlocksThresholdSeconds)
	}

	// check config
	{

		cfg := &types.Config{}
		err := utils.ReadConfig(cfg, *configPath)
		if err != nil {
			log.Fatal(err, "error reading config file", 0) // fatal, as there is no point without a config
		} else {
			log.Info("reading config completed")
		}
		utils.Config = cfg
		log.InfoWithFields(log.Fields{"config": *configPath, "version": version.Version, "commit": version.GitCommit, "chainName": utils.Config.Chain.ClConfig.ConfigName}, "starting")

		if len(utils.Config.Eth1ErigonEndpoint) > 0 {
			eth1RpcEndpoint = utils.Config.Eth1ErigonEndpoint
		} else {
			eth1RpcEndpoint = utils.Config.Eth1GethEndpoint
		}

		if utils.Config.Metrics.Enabled {
			go func() {
				log.Infof("serving metrics on %v", utils.Config.Metrics.Address)
				if err := metrics.Serve(utils.Config.Metrics.Address, utils.Config.Metrics.Pprof, utils.Config.Metrics.PprofExtra); err != nil {
					log.Fatal(err, "error serving metrics", 0)
				}
			}()
		}
	}

	// check parameters
	if *nodeRequestsAtOnce < 1 {
		log.Warnf("node-requests-at-once set to %s, corrected to 1", _formatInt64(*nodeRequestsAtOnce))
		*nodeRequestsAtOnce = 1
	}
	if *nodeRequestsAtOnce > MAX_NODE_REQUESTS_AT_ONCE {
		log.Warnf("node-requests-at-once set to %s, corrected to %s", _formatInt64(*nodeRequestsAtOnce), _formatInt64(MAX_NODE_REQUESTS_AT_ONCE))
		*nodeRequestsAtOnce = MAX_NODE_REQUESTS_AT_ONCE
	}
	if *reorgDepth < 0 || *reorgDepth > MAX_REORG_DEPTH {
		log.Warnf("reorg.depth parameter set to %s, corrected to %s", _formatInt64(*reorgDepth), _formatInt64(MAX_REORG_DEPTH))
		*reorgDepth = MAX_REORG_DEPTH
	}
	if *concurrency < 1 {
		log.Warnf("concurrency parameter set to %s, corrected to 1", _formatInt64(*concurrency))
		*concurrency = 1
	}
	if *noNewBlocksThresholdSeconds < 30 {
		log.Warnf("fatal-if-no-new-block-for-x-seconds set to %d, corrected to 30", *noNewBlocksThresholdSeconds)
		*noNewBlocksThresholdSeconds = 30
	}

	// init postgres
	{
		db.WriterDb, db.ReaderDb = db.MustInitDB(&types.DatabaseConfig{
			Username:     utils.Config.WriterDatabase.Username,
			Password:     utils.Config.WriterDatabase.Password,
			Name:         utils.Config.WriterDatabase.Name,
			Host:         utils.Config.WriterDatabase.Host,
			Port:         utils.Config.WriterDatabase.Port,
			MaxOpenConns: utils.Config.WriterDatabase.MaxOpenConns,
			MaxIdleConns: utils.Config.WriterDatabase.MaxIdleConns,
			SSL:          utils.Config.WriterDatabase.SSL,
		}, &types.DatabaseConfig{
			Username:     utils.Config.ReaderDatabase.Username,
			Password:     utils.Config.ReaderDatabase.Password,
			Name:         utils.Config.ReaderDatabase.Name,
			Host:         utils.Config.ReaderDatabase.Host,
			Port:         utils.Config.ReaderDatabase.Port,
			MaxOpenConns: utils.Config.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: utils.Config.ReaderDatabase.MaxIdleConns,
			SSL:          utils.Config.ReaderDatabase.SSL,
		}, "pgx", "postgres")
		defer db.ReaderDb.Close()
		defer db.WriterDb.Close()
		log.Info("starting postgres completed")
	}

	// init bigtable
	log.Info("init BT...")
	btClient, err := gcp_bigtable.NewClient(context.Background(), utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, option.WithGRPCConnectionPool(1))
	if err != nil {
		log.Fatal(err, "creating new client for Bigtable", 0) // fatal, no point to continue without BT
	}
	tableBlocksRaw := btClient.Open("blocks-raw")
	if tableBlocksRaw == nil {
		log.Fatal(err, "open blocks-raw table", 0) // fatal, no point to continue without BT
	}
	defer btClient.Close()
	log.Info("...init BT done.")

	// init el client
	log.Info("init el client endpoint...")
	// #RECY IMPROVE split http / ws endpoint, http is mandatory, ws optional - So add an http/ws config entry, where ws is optional (to use subscribe)
	elClient, err = ethclient.Dial(eth1RpcEndpoint)
	if err != nil {
		log.Fatal(err, "error dialing eth url", 0) // fatal, no point to continue without node connection
	}
	log.Info("...init el client endpoint done.")

	// check chain id
	{
		log.Info("check chain id...")
		chainID, err := rpciGetChainId()
		if chainID == ARBITRUM_CHAINID { // #RECY REMOVE currently necessary as there is no default config / setting in utils for Arbitrum
			utils.Config.Chain.Id = ARBITRUM_CHAINID
		}
		if chainID == OPTIMISM_CHAINID { // #RECY REMOVE currently necessary as there is no default config / setting in utils for Optimism
			utils.Config.Chain.Id = OPTIMISM_CHAINID
		}
		if err != nil {
			log.Fatal(err, "error get chain id", 0) // fatal, no point to continue without chain id
		}
		if chainID != utils.Config.Chain.Id { // if the chain id is removed from the config, just remove this if, there is no point, except checking consistency
			log.Fatal(err, "node chain different from config chain", 0) // fatal, config doesn't match node
		}
		log.InfoWithFields(log.Fields{"chainId": utils.Config.Chain.Id, "eth1RpcEndpoint": eth1RpcEndpoint, "bt.project": utils.Config.Bigtable.Project, "bt.instance": utils.Config.Bigtable.Instance}, "...check chain id done.")
	}

	// get latest block (as it's global, so we have a initial value)
	log.Info("get latest block from node...")
	updateBlockNumber(true, *noNewBlocks, time.Duration(*noNewBlocksThresholdSeconds)*time.Second, discordWebhookReportUrl, discordWebhookUser, discordWebhookAddTextFatal)
	log.Infof("...get latest block (%s) from node done.", _formatInt64(currentNodeBlockNumber.Load()))

	// //////////////////////////////////////////
	// Config done, now actually "doing" stuff //
	// //////////////////////////////////////////

	// check if reexport requested
	if *startBlockNumber >= 0 && *endBlockNumber >= 0 && *startBlockNumber <= *endBlockNumber {
		log.Infof("Found REEXPORT for block %s to %s...", _formatInt64(*startBlockNumber), _formatInt64(*endBlockNumber))
		err := bulkExportBlocksRange(tableBlocksRaw, []intRange{{start: *startBlockNumber, end: *endBlockNumber}}, *concurrency, *nodeRequestsAtOnce, discordWebhookBlockThreshold, discordWebhookReportUrl, discordWebhookUser)
		if err != nil {
			sendMessage(fmt.Sprintf("%s NODE EXPORT: Fatal, reexport not completed, check logs %s", getChainNamePretty(), *discordWebhookAddTextFatal), discordWebhookReportUrl, discordWebhookUser)
			log.Fatal(err, "error while reexport blocks for bigtable (reexport range)", 0) // fatal, as there is nothing more todo anyway
		}
		log.Info("Job done, have a nice day :)")
		return
	}

	// find holes in our previous runs / sanity check
	if *skipHoleCheck {
		log.Warn("Skipping hole check!")
	} else {
		log.Info("Checking for holes...")
		startTime := time.Now()
		missingBlocks, err := psqlFindGaps() // find the holes
		findHolesTook := time.Since(startTime)
		if err != nil {
			log.Fatal(err, "error checking for holes", 0) // fatal, as we highly depend on postgres, if this is not working, we can quit
		}
		l := len(missingBlocks)
		if l > 0 { // some holes found
			log.Warnf("Found %s missing block ranges in %v, fixing them now...", _formatInt(l), findHolesTook)
			if l <= 10 {
				log.Warnf("%v", missingBlocks)
			} else {
				log.Warnf("%v<...>", missingBlocks[:10])
			}
			startTime = time.Now()
			err := bulkExportBlocksRange(tableBlocksRaw, missingBlocks, *concurrency, *nodeRequestsAtOnce, discordWebhookBlockThreshold, discordWebhookReportUrl, discordWebhookUser) // reexport the holes
			if err != nil {
				log.Fatal(err, "error while reexport blocks for bigtable (fixing holes)", 0) // fatal, as if we wanna start with holes, we should set the skip-hole-check parameter
			}
			log.Warnf("...fixed them in %v", time.Since(startTime))
		} else {
			log.Infof("...no missing block found in %v", findHolesTook)
		}
	}
	if *onlyHoleCheck {
		log.Info("only-hole-check set, job done, have a nice day :)")
		return
	}

	// waiting for new blocks and export them, while checking reorg before every new block
	latestPGBlock, err := psqlGetLatestBlock(false)
	if err != nil {
		log.Fatal(err, "error while using psqlGetLatestBlock (start / read)", 0) // fatal, as if there is no initial value, we have nothing to start from
	}
	var consecutiveErrorCount int
	consecutiveErrorCountThreshold := 0 // after threshold + 1 errors it will be fatal instead #TODO not working correct wenn syncing big amount of data, setting meanwhile to 0, as an error will result in a fully retry (which is wrong)
	for {
		currentNodeBN := currentNodeBlockNumber.Load()
		if currentNodeBN < latestPGBlock {
			// fatal, as this is an impossible error
			log.Fatal(err, "impossible error currentNodeBN < lastestPGBlock", 0, map[string]interface{}{"currentNodeBN": currentNodeBN, "latestPGBlock": latestPGBlock})
		} else if currentNodeBN == latestPGBlock {
			time.Sleep(time.Second)
			continue // still the same block
		} else {
			// checking for reorg
			if *reorgDepth > 0 && latestPGBlock >= 0 {
				// define length to check
				l := *reorgDepth
				if l > latestPGBlock+1 {
					l = latestPGBlock + 1
				}

				// fill array with block numbers to check
				blockRawData := make([]fullBlockRawData, l)
				for i := int64(0); i < l; i++ {
					blockRawData[i].blockNumber = latestPGBlock + i - l + 1
				}

				// get all hashes from node
				err = rpciGetBulkBlockRawHash(blockRawData, *nodeRequestsAtOnce)
				if err != nil {
					consecutiveErrorCount++
					if consecutiveErrorCount <= consecutiveErrorCountThreshold {
						log.Error(err, "error when bulk getting raw block hashes", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "latestPGBlock": latestPGBlock, "reorgDepth": *reorgDepth})
					} else {
						log.Fatal(err, "error when bulk getting raw block hashes", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "latestPGBlock": latestPGBlock, "reorgDepth": *reorgDepth})
					}
					continue
				}

				// get a list of all block_ids where the hashes are fine
				var matchingHashesBlockIdList []int64
				matchingHashesBlockIdList, err = psqlGetHashHitsIdList(blockRawData)
				if err != nil {
					consecutiveErrorCount++
					if consecutiveErrorCount <= consecutiveErrorCountThreshold {
						log.Error(err, "error when getting hash hits id list", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "latestPGBlock": latestPGBlock, "reorgDepth": *reorgDepth})
					} else {
						log.Fatal(err, "error when getting hash hits id list", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "latestPGBlock": latestPGBlock, "reorgDepth": *reorgDepth})
					}
					continue
				}

				matchingLength := len(matchingHashesBlockIdList)
				if len(blockRawData) != matchingLength { // nothing todo if all elements are fine, but if not...
					if len(blockRawData) < matchingLength {
						// fatal, as this is an impossible error
						log.Fatal(err, "impossible error len(blockRawData) < matchingLength", 0, map[string]interface{}{"latestPGBlock": latestPGBlock, "matchingLength": matchingLength})
					}

					// reverse the "fine" list, so we have a "not fine" list
					wrongHashRanges := []intRange{{start: -1}}
					wrongHashRangesIndex := 0
					var i int
					var failCounter int
					for _, v := range blockRawData {
						for i < matchingLength && v.blockNumber > matchingHashesBlockIdList[i] {
							i++
						}
						if i >= matchingLength || v.blockNumber != matchingHashesBlockIdList[i] {
							failCounter++
							if wrongHashRanges[wrongHashRangesIndex].start < 0 {
								wrongHashRanges[wrongHashRangesIndex].start = v.blockNumber
								wrongHashRanges[wrongHashRangesIndex].end = v.blockNumber
							} else if wrongHashRanges[wrongHashRangesIndex].end+1 == v.blockNumber {
								wrongHashRanges[wrongHashRangesIndex].end = v.blockNumber
							} else {
								wrongHashRangesIndex++
								wrongHashRanges[wrongHashRangesIndex].start = v.blockNumber
								wrongHashRanges[wrongHashRangesIndex].end = v.blockNumber
							}
						}
					}
					if failCounter != len(blockRawData)-matchingLength {
						// fatal, as this is an impossible error
						log.Fatal(err, "impossible error failureLength != len(blockRawData)-matchingLength", 0, map[string]interface{}{"failCounter": failCounter, "len(blockRawData)-matchingLength": len(blockRawData) - matchingLength})
					}
					log.Infof("found %s wrong hashes when checking for reorgs, reexporting them now...", _formatInt(failCounter))
					log.Infof("%v", wrongHashRanges)

					// export the hits again
					err = bulkExportBlocksRange(tableBlocksRaw, wrongHashRanges, *concurrency, *nodeRequestsAtOnce, discordWebhookBlockThreshold, discordWebhookReportUrl, discordWebhookUser)
					// we will retry again, but it's important to skip the export of new blocks in the case of an error
					if err != nil {
						consecutiveErrorCount++
						if consecutiveErrorCount <= consecutiveErrorCountThreshold {
							log.Error(err, "error exporting hits on reorg", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "len(blockRawData)": len(blockRawData), "reorgDepth": *reorgDepth, "matchingHashesBlockIdList": matchingHashesBlockIdList, "wrongHashRanges": wrongHashRanges})
						} else {
							log.Fatal(err, "error exporting hits on reorg", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "len(blockRawData)": len(blockRawData), "reorgDepth": *reorgDepth, "matchingHashesBlockIdList": matchingHashesBlockIdList, "wrongHashRanges": wrongHashRanges})
						}
						continue
					} else {
						log.Info("...done. Everything fine with reorgs again.")
					}
				}
			}

			// export all new blocks
			newerNodeBN := currentNodeBlockNumber.Load() // just in case it took a while doing the reorg stuff, no problem if range > reorg limit, as the exported blocks will be newest also
			if newerNodeBN < currentNodeBN {
				// fatal, as this is an impossible error
				log.Fatal(err, "impossible error newerNodeBN < currentNodeBN", 0, map[string]interface{}{"newerNodeBN": newerNodeBN, "currentNodeBN": currentNodeBN})
			}
			err = bulkExportBlocksRange(tableBlocksRaw, []intRange{{start: latestPGBlock + 1, end: newerNodeBN}}, *concurrency, *nodeRequestsAtOnce, discordWebhookBlockThreshold, discordWebhookReportUrl, discordWebhookUser)
			// we can try again, as throw a fatal will result in try again anyway
			if err != nil {
				consecutiveErrorCount++
				if consecutiveErrorCount <= consecutiveErrorCountThreshold {
					log.Error(err, "error while reexport blocks for bigtable (newest blocks)", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "latestPGBlock+1": latestPGBlock + 1, "newerNodeBN": newerNodeBN})
				} else {
					log.Fatal(err, "error while reexport blocks for bigtable (newest blocks)", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount, "latestPGBlock+1": latestPGBlock + 1, "newerNodeBN": newerNodeBN})
				}
				continue
			} else {
				latestPGBlock, err = psqlGetLatestBlock(true)
				if err != nil {
					consecutiveErrorCount++
					if consecutiveErrorCount <= consecutiveErrorCountThreshold {
						log.Error(err, "error while using psqlGetLatestBlock (ongoing / write)", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount})
					} else {
						log.Fatal(err, "error while using psqlGetLatestBlock (ongoing / write)", 0, map[string]interface{}{"reorgErrorCount": consecutiveErrorCount})
					}
					continue
				} else if latestPGBlock != newerNodeBN {
					// fatal, as this is a nearly impossible error
					log.Fatal(err, "impossible error latestPGBlock != newerNodeBN", 0, map[string]interface{}{"latestPGBlock": latestPGBlock, "newerNodeBN": newerNodeBN})
				}
			}

			// reset consecutive error count if no change during this run
			if consecutiveErrorCount > 0 {
				log.Infof("reset consecutive error count to 0, as no error in this run (was %d)", consecutiveErrorCount)
				consecutiveErrorCount = 0
			}
		}
	}
}

// improve the behaviour in case of an error
func _bulkExportBlocksHandler(tableBlocksRaw *gcp_bigtable.Table, blockRawData []fullBlockRawData, nodeRequestsAtOnce int64, deep int) error {
	err := _bulkExportBlocksImpl(tableBlocksRaw, blockRawData, nodeRequestsAtOnce)
	if err != nil {
		if deep < TRY_TO_RECOVER_ON_ERROR_COUNT {
			elementCount := len(blockRawData)

			// output the error
			{
				s := errorIdentifier.FindStringSubmatch(err.Error())
				if len(s) >= 2 { // if we have a valid json error available, should be the case if it's a node issue
					log.WarnWithFields(log.Fields{"deep": deep, "cause": s[1], "0block": blockRawData[0].blockNumber, "elements": elementCount}, "got an error and will try to fix it (sub)")
				} else { // if we have a no json error available, should be the case if it's a BT or Postgres issue
					log.WarnWithFields(log.Fields{"deep": deep, "cause": err, "0block": blockRawData[0].blockNumber, "elements": elementCount}, "got an error and will try to fix it (err)")
				}
			}

			if deep > TRY_TO_RECOVER_ON_ERROR_COUNT/2 {
				duration := deep - TRY_TO_RECOVER_ON_ERROR_COUNT/2
				if duration > 8 {
					duration = 8
				}
				time.Sleep(time.Second * time.Duration(duration))
			}

			// try to recover
			if elementCount == 1 { // if there is only 1 element, no split possible
				err = _bulkExportBlocksHandler(tableBlocksRaw, blockRawData, nodeRequestsAtOnce, deep+1)
			} else if elementCount > 1 { // split the elements in half and try again to put less strain on the node
				err = _bulkExportBlocksHandler(tableBlocksRaw, blockRawData[:elementCount/2], nodeRequestsAtOnce, deep+1)
				if err == nil {
					err = _bulkExportBlocksHandler(tableBlocksRaw, blockRawData[elementCount/2:], nodeRequestsAtOnce, deep+1)
				}
			}
		}
	}
	if err != nil {
		return fmt.Errorf("_bulkExportBlocksHandler with deep (%d): %w", deep, err)
	}
	return nil
}

// export all blocks, heavy use of bulk & concurrency, providing a block raw data array (used by the other bulkExportBlocks+ functions)
func _bulkExportBlocksImpl(tableBlocksRaw *gcp_bigtable.Table, blockRawData []fullBlockRawData, nodeRequestsAtOnce int64) error {
	// check values
	{
		if tableBlocksRaw == nil {
			return fmt.Errorf("tableBlocksRaw == nil")
		}

		l := int64(len(blockRawData))
		if l < 1 || l > nodeRequestsAtOnce {
			return fmt.Errorf("blockRawData length (%d) is 0 or greater 'node requests at once' (%d)", l, nodeRequestsAtOnce)
		}
	}

	// get block_hash, block_unclesCount, block_compressed & block_txs
	err := rpciGetBulkBlockRawData(blockRawData, nodeRequestsAtOnce)
	if err != nil {
		return fmt.Errorf("rpciGetBulkBlockRawData: %w", err)
	}
	err = rpciGetBulkRawUncles(blockRawData, nodeRequestsAtOnce)
	if err != nil {
		return fmt.Errorf("rpciGetBulkRawUncles: %w", err)
	}
	err = rpciGetBulkRawReceipts(blockRawData, nodeRequestsAtOnce)
	if err != nil {
		return fmt.Errorf("rpciGetBulkRawReceipts: %w", err)
	}
	err = rpciGetBulkRawTraces(blockRawData, nodeRequestsAtOnce)
	if err != nil {
		return fmt.Errorf("rpciGetBulkRawTraces: %w", err)
	}

	// write to bigtable
	{
		// prepare array
		muts := []*gcp_bigtable.Mutation{}
		keys := []string{}
		for _, v := range blockRawData {
			if len(v.blockCompressed) == 0 || len(v.tracesCompressed) == 0 {
				log.Fatal(nil, "tried writing empty data to BT", 0, map[string]interface{}{"len(v.blockCompressed)": len(v.blockCompressed), "len(v.receiptsCompressed)": len(v.receiptsCompressed), "len(v.tracesCompressed)": len(v.tracesCompressed)}) // fatal, as if this is not working in the first place, it will never work
			}
			mut := gcp_bigtable.NewMutation()
			mut.Set(BT_COLUMNFAMILY_BLOCK, BT_COLUMN_BLOCK, gcp_bigtable.Timestamp(0), v.blockCompressed)
			if len(v.receiptsCompressed) < 1 {
				log.Warnf("empty receipts at block %d lRec %d lTxs %d", v.blockNumber, len(v.receiptsCompressed), len(v.blockTxs))
			}
			mut.Set(BT_COLUMNFAMILY_RECEIPTS, BT_COLUMN_RECEIPTS, gcp_bigtable.Timestamp(0), v.receiptsCompressed)
			mut.Set(BT_COLUMNFAMILY_TRACES, BT_COLUMN_TRACES, gcp_bigtable.Timestamp(0), v.tracesCompressed)
			if v.blockUnclesCount > 0 {
				mut.Set(BT_COLUMNFAMILY_UNCLES, BT_COLUMN_UNCLES, gcp_bigtable.Timestamp(0), v.unclesCompressed)
			}
			muts = append(muts, mut)
			keys = append(keys, fmt.Sprintf("%d:%12d", utils.Config.Chain.Id, MAX_EL_BLOCK_NUMBER-v.blockNumber))
		}

		// write
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		var errs []error
		errs, err = tableBlocksRaw.ApplyBulk(ctx, keys, muts)
		if err != nil {
			return fmt.Errorf("tableBlocksRaw.ApplyBulk err: %w", err)
		}
		for i, e := range errs {
			return fmt.Errorf("tableBlocksRaw.ApplyBulk errs(%d): %w", i, e)
		}
	}

	// write to SQL
	err = psqlAddElements(blockRawData)
	if err != nil {
		return fmt.Errorf("psqlAddElements: %w", err)
	}

	return nil
}

// export all blocks, heavy use of bulk & concurrency, providing a range array
func bulkExportBlocksRange(tableBlocksRaw *gcp_bigtable.Table, blockRanges []intRange, concurrency int64, nodeRequestsAtOnce int64, discordWebhookBlockThreshold *int64, discordWebhookReportUrl *string, discordWebhookUser *string) error {
	{
		var blocksTotalCount int64
		l := len(blockRanges)
		if l <= 0 {
			return fmt.Errorf("got empty blockRanges array")
		}
		for i, v := range blockRanges {
			if v.start <= v.end {
				blocksTotalCount += v.end - v.start + 1
			} else {
				return fmt.Errorf("blockRanges at index %d has wrong start (%s) > end (%s) combination", i, _formatInt64(v.start), _formatInt64(v.end))
			}
		}

		if l == 1 {
			log.Infof("Only 1 range found, started export of blocks %s to %s, total block amount %s, using an updater every %d seconds for more details.", _formatInt64(blockRanges[0].start), _formatInt64(blockRanges[0].end), _formatInt64(blocksTotalCount), OUTPUT_CYCLE_IN_SECONDS)
		} else {
			log.Infof("%d ranges found, total block amount %d, using an updater every %d seconds for more details.", l, blocksTotalCount, OUTPUT_CYCLE_IN_SECONDS)
		}
	}

	gOuterMustStop := atomic.Bool{}
	gOuter := &errgroup.Group{}
	gOuter.SetLimit(int(concurrency))

	totalStart := time.Now()
	exportStart := totalStart
	var lastDiscordReportAtBlocksProcessedTotal int64
	blocksProcessedTotal := atomic.Int64{}
	blocksProcessedIntv := atomic.Int64{}

	go func() {
		for {
			time.Sleep(time.Second * OUTPUT_CYCLE_IN_SECONDS)
			if gOuterMustStop.Load() {
				break
			}

			bpi := blocksProcessedIntv.Swap(0)
			newStart := time.Now()
			blocksProcessedTotal.Add(bpi)
			bpt := blocksProcessedTotal.Load()

			var totalBlocks int64
			latestNodeBlock := currentNodeBlockNumber.Load()
			for _, v := range blockRanges {
				if v.end > latestNodeBlock {
					totalBlocks += latestNodeBlock - v.start + 1
				} else {
					totalBlocks += v.end - v.start + 1
				}
			}
			blocksPerSecond := float64(bpi) / time.Since(exportStart).Seconds()
			blocksPerSecondTotal := float64(bpt) / time.Since(totalStart).Seconds()
			durationRemainingTotal := time.Second * time.Duration(float64(totalBlocks-bpt)/float64(blocksPerSecondTotal))

			log.Infof("current speed: %0.1f b/s %0.1f t/s %s remain %s total %0.2fh (=%0.2fd to go)", blocksPerSecond, blocksPerSecondTotal, _formatInt64(totalBlocks-bpt), _formatInt64(totalBlocks), durationRemainingTotal.Hours(), durationRemainingTotal.Hours()/24)
			exportStart = newStart
			if lastDiscordReportAtBlocksProcessedTotal+(*discordWebhookBlockThreshold) <= bpt {
				lastDiscordReportAtBlocksProcessedTotal += (*discordWebhookBlockThreshold)
				sendMessage(fmt.Sprintf("%s NODE EXPORT: %0.1f block/s %s remaining (%0.1f day/s to go)", getChainNamePretty(), blocksPerSecondTotal, _formatInt64(totalBlocks-bpt), durationRemainingTotal.Hours()/24), discordWebhookReportUrl, discordWebhookUser)
			}
		}
	}()
	defer gOuterMustStop.Store(true) // kill the updater

	blockRawData := make([]fullBlockRawData, 0, nodeRequestsAtOnce)
	blockRawDataLen := int64(0)
Loop:
	for _, blockRange := range blockRanges {
		current := blockRange.start
		for blockRange.end-current+1 > 0 {
			if gOuterMustStop.Load() {
				break Loop
			}

			currentNodeBlockNumberLocalCopy := currentNodeBlockNumber.Load()
			for blockRawDataLen < nodeRequestsAtOnce && current <= blockRange.end {
				if currentNodeBlockNumberLocalCopy >= current {
					blockRawData = append(blockRawData, fullBlockRawData{blockNumber: current})
					blockRawDataLen++
					current++
				} else {
					log.Warnf("tried to export block %d, but latest block on node is %d, so stopping all further export till %d", current, currentNodeBlockNumberLocalCopy, blockRange.end)
					current = blockRange.end + 1
				}
			}
			if blockRawDataLen == nodeRequestsAtOnce {
				brd := blockRawData
				gOuter.Go(func() error {
					err := _bulkExportBlocksHandler(tableBlocksRaw, brd, nodeRequestsAtOnce, 0)
					if err != nil {
						gOuterMustStop.Store(true)
						return err
					}
					blocksProcessedIntv.Add(int64(len(brd)))
					return nil
				})
				blockRawData = make([]fullBlockRawData, 0, nodeRequestsAtOnce)
				blockRawDataLen = 0
			}
		}
	}

	// write the rest
	if !gOuterMustStop.Load() && blockRawDataLen > 0 {
		brd := blockRawData
		gOuter.Go(func() error {
			err := _bulkExportBlocksHandler(tableBlocksRaw, brd, nodeRequestsAtOnce, 0)
			if err != nil {
				gOuterMustStop.Store(true)
				return err
			}
			blocksProcessedIntv.Add(int64(len(brd)))
			return nil
		})
	}

	return gOuter.Wait()
}

// //////////
// HELPERs //
// //////////
// Send message to discord
func sendMessage(content string, webhookUrl *string, username *string) {
	if len(*webhookUrl) > 0 {
		err := discordwebhook.SendMessage(*webhookUrl, discordwebhook.Message{Username: username, Content: &content})
		if err != nil {
			log.Error(err, "error sending message to discord", 0, map[string]interface{}{"content": content, "webhookUrl": *webhookUrl, "username": *username})
		}
	}
}

// Get pretty name for chain
func getChainNamePretty() string {
	switch utils.Config.Chain.Id {
	case MAINNET_CHAINID:
		return "<:eth:1184470363967598623> ETHEREUM mainnet"
	case GOERLI_CHAINID:
		return "GOERLI testnet"
	case OPTIMISM_CHAINID:
		return "<:op:1184470125458489354> OPTIMISM mainnet"
	case GNOSIS_CHAINID:
		return "<:gnosis:1184470353947398155> GNOSIS mainnet"
	case HOLESKY_CHAINID:
		return "HOLESKY testnet"
	case ARBITRUM_CHAINID:
		return "<:arbitrum:1184470344506036334> ARBITRUM mainnet"
	case SEPOLIA_CHAINID:
		return "SEPOLIA testnet"
	}
	return fmt.Sprintf("%d", utils.Config.Chain.Id)
}

// format int for pretty output
func _formatInt(value int) string {
	return _formatInt64(int64(value))
}

// format int64 for pretty output
func _formatInt64(value int64) string {
	result := ""
	for value >= 1000 {
		lastPart := value % 1000
		value /= 1000
		if len(result) > 0 {
			result = fmt.Sprintf("%03d,%s", lastPart, result)
		} else {
			result = fmt.Sprintf("%03d", lastPart)
		}
	}
	if len(result) > 0 {
		return fmt.Sprintf("%d,%s", value, result)
	}
	return fmt.Sprintf("%d", value)
}

// compress given byte slice
func compress(src []byte) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(src); err != nil {
		log.Fatal(err, "error writing to gzip writer", 0) // fatal, as if this is not working in the first place, it will never work
	}
	if err := zw.Close(); err != nil {
		log.Fatal(err, "error closing gzip writer", 0) // fatal, as if this is not working in the first place, it will never work
	}
	return buf.Bytes()
}

// decompress given byte slice
/* func decompress(src []byte) []byte {
	zr, err := gzip.NewReader(bytes.NewReader(src))
	if err != nil {
		log.Fatal(err, "error creating gzip reader", 0) // fatal, as if this is not working in the first place, it will never work
	}
	data, err := io.ReadAll(zr)
	if err != nil {
		log.Fatal(err, "error reading from gzip reader", 0) // fatal, as if this is not working in the first place, it will never work
	}
	return data
} */

// used by splitAndVerifyJsonArray to add an element to the list depending on its Id
func _splitAndVerifyJsonArrayAddElement(r *[][]byte, element []byte, lastId int64) (int64, error) {
	// adding empty elements will cause issues, so we don't allow it
	if len(element) <= 0 {
		return -1, fmt.Errorf("error, tried to add empty element, lastId (%d)", lastId)
	}

	// unmarshal
	data := &jsonRpcReturnId{}
	err := json.Unmarshal(element, data)
	if err != nil {
		return -1, fmt.Errorf("error decoding '%s': %w", element, err)
	}

	// negativ ids signals an issue
	if data.Id < 0 {
		return -1, fmt.Errorf("error, provided Id (%d) < 0", data.Id)
	}
	// id must ascending or equal
	if data.Id < lastId {
		return -1, fmt.Errorf("error, provided Id (%d) < lastId (%d)", data.Id, lastId)
	}

	// new element
	if data.Id != lastId {
		*r = append(*r, element)
	} else { // append element (same id)
		i := len(*r) - 1
		if (*r)[i][0] == byte('[') {
			(*r)[i] = (*r)[i][1 : len((*r)[i])-1]
		}
		(*r)[i] = append(append(append(append([]byte("["), (*r)[i]...), byte(',')), element...), byte(']'))
	}

	return data.Id, nil
}

// split a bulk json request in single requests
func _splitAndVerifyJsonArray(jArray []byte, providedElementCount int64) ([][]byte, error) {
	endDigit := byte('}')
	searchValue := []byte(`{"jsonrpc":"`)
	searchLen := len(searchValue)
	foundElementCount := int64(0)

	// remove everything before the first hit
	i := bytes.Index(jArray, searchValue)
	if i < 0 {
		return nil, fmt.Errorf("no element found")
	}
	jArray = jArray[i:]

	// find all elements
	var err error
	lastId := int64(-1)
	r := make([][]byte, 0)
	for {
		if len(jArray) < searchLen { // weird corner case, shouldn't happen at all
			i = -1
		} else { // get next hit / ignore current (at index 0)
			i = bytes.Index(jArray[searchLen:], searchValue)
		}
		// handle last element
		if i < 0 {
			for l := len(jArray) - 1; l >= 0 && jArray[l] != endDigit; l-- {
				jArray = jArray[:l]
			}
			foundElementCount++
			_, err = _splitAndVerifyJsonArrayAddElement(&r, jArray, lastId)
			if err != nil {
				return nil, fmt.Errorf("error calling split and verify json array add element - last element: %w", err)
			}
			break
		}
		// handle normal element
		foundElementCount++
		lastId, err = _splitAndVerifyJsonArrayAddElement(&r, jArray[:i+searchLen-1], lastId)
		if err != nil {
			return nil, fmt.Errorf("error calling split and verify json array add element: %w", err)
		}
		// set cursor to new start
		jArray = jArray[i+searchLen:]
	}
	if foundElementCount != providedElementCount {
		return r, fmt.Errorf("provided element count %d doesn't match found %d", providedElementCount, foundElementCount)
	}
	return r, nil
}

// get newest block number from node, should be called always with TRUE
func updateBlockNumber(firstCall bool, noNewBlocks bool, noNewBlocksThresholdDuration time.Duration, discordWebhookReportUrl *string, discordWebhookUser *string, discordWebhookAddTextFatal *string) {
	if firstCall {
		blockNumber, err := rpciGetLatestBlock()
		if err != nil {
			sendMessage(fmt.Sprintf("%s NODE EXPORT: Fatal, failed to get newest block from node, on first try %s", getChainNamePretty(), *discordWebhookAddTextFatal), discordWebhookReportUrl, discordWebhookUser)
			log.Fatal(err, "fatal, failed to get newest block from node, on first try", 0)
		}
		currentNodeBlockNumber.Store(blockNumber)
		if !noNewBlocks {
			go updateBlockNumber(false, false, noNewBlocksThresholdDuration, discordWebhookReportUrl, discordWebhookUser, discordWebhookAddTextFatal)
		}
		return
	}

	var errorText string
	gotNewBlockAt := time.Now()
	timePerBlock := time.Second * time.Duration(utils.Config.Chain.ClConfig.SecondsPerSlot)
	if strings.HasPrefix(eth1RpcEndpoint, "ws") {
		log.Infof("ws node endpoint found, will use subscribe")
		var timer *time.Timer
		previousBlock := int64(-1)
		newestBlock := int64(-1)
		for {
			headers := make(chan *gethtypes.Header)
			sub, err := rpciSubscribeNewHead(headers)
			if err != nil {
				errorText = "error, init subscribe for new head"
			} else {
			Loop:
				for {
					if timer != nil && !timer.Stop() {
						<-timer.C
					}
					timer = time.NewTimer(noNewBlocksThresholdDuration / 3)

					select {
					case err = <-sub.Err():
						errorText = "error, subscribe new head was canceled"
						break Loop
					case <-timer.C:
						errorText = "error, timer triggered for subscribe of new head"
						break Loop
					case header := <-headers:
						previousBlock = currentNodeBlockNumber.Load()
						newestBlock = header.Number.Int64()
						if newestBlock <= previousBlock {
							log.Fatal(nil, "impossible error, newest block <= previous block", 0, map[string]interface{}{"previousBlock": previousBlock, "newestBlock": newestBlock})
						}
						currentNodeBlockNumber.Store(newestBlock)
						gotNewBlockAt = time.Now()
					}
				}
			}

			durationSinceLastBlockReceived := time.Since(gotNewBlockAt)
			if durationSinceLastBlockReceived < noNewBlocksThresholdDuration/3*2 {
				log.WarnWithFields(log.Fields{"durationSinceLastBlockReceived": durationSinceLastBlockReceived, "error": err}, errorText)
			} else if durationSinceLastBlockReceived < noNewBlocksThresholdDuration {
				log.Error(err, errorText, 0, map[string]interface{}{"durationSinceLastBlockReceived": durationSinceLastBlockReceived, "previousBlock": previousBlock, "newestBlock": newestBlock})
			} else {
				sendMessage(fmt.Sprintf("%s NODE EXPORT: Fatal, %s, %v, %v %s", getChainNamePretty(), errorText, err, durationSinceLastBlockReceived, *discordWebhookAddTextFatal), discordWebhookReportUrl, discordWebhookUser)
				log.Fatal(err, errorText, 0, map[string]interface{}{"durationSinceLastBlockReceived": durationSinceLastBlockReceived, "previousBlock": previousBlock, "newestBlock": newestBlock})
			}

			close(headers)
			sub.Unsubscribe()
			time.Sleep(timePerBlock) // Sleep for 1 block in case of an error
		}
	} else { // no ws node endpoint available
		log.Infof("no ws node endpoint found, can't use subscribe")
		errorText := "error, no new block for a longer time"
		for {
			time.Sleep(timePerBlock / 2) // wait half a block
			previousBlock := currentNodeBlockNumber.Load()
			newestBlock, err := rpciGetLatestBlock()
			if err == nil {
				if previousBlock > newestBlock {
					log.Fatal(nil, "impossible error, newest block <= previous block", 0, map[string]interface{}{"previousBlock": previousBlock, "newestBlock": newestBlock})
				} else if previousBlock < newestBlock {
					currentNodeBlockNumber.Store(newestBlock)
					gotNewBlockAt = time.Now()
					continue
				}
			}

			durationSinceLastBlockReceived := time.Since(gotNewBlockAt)
			if durationSinceLastBlockReceived >= noNewBlocksThresholdDuration {
				sendMessage(fmt.Sprintf("%s NODE EXPORT: Fatal, %s, %d, %d, %v, %v %s", getChainNamePretty(), errorText, previousBlock, newestBlock, err, durationSinceLastBlockReceived, *discordWebhookAddTextFatal), discordWebhookReportUrl, discordWebhookUser)
				log.Fatal(err, errorText, 0, map[string]interface{}{"durationSinceLastBlockReceived": durationSinceLastBlockReceived, "previousBlock": previousBlock, "newestBlock": newestBlock})
			} else if durationSinceLastBlockReceived >= noNewBlocksThresholdDuration/3*2 {
				log.Error(err, errorText, 0, map[string]interface{}{"durationSinceLastBlockReceived": durationSinceLastBlockReceived, "previousBlock": previousBlock, "newestBlock": newestBlock})
			} else if durationSinceLastBlockReceived >= noNewBlocksThresholdDuration/3 {
				log.WarnWithFields(log.Fields{"durationSinceLastBlockReceived": durationSinceLastBlockReceived, "error": err, "previousBlock": previousBlock, "newestBlock": newestBlock}, errorText)
			}
		}
	}
}

// /////////////////////
// Postgres interface //
// /////////////////////
// find gaps (missing ids) in raw_block_status
func psqlFindGaps() ([]intRange, error) {
	gaps := []intRange{}

	// check for a gap at the beginning
	{
		var firstBlock int64
		err := db.ReaderDb.Get(&firstBlock, `SELECT block_id FROM raw_block_status WHERE chain_id = $1 ORDER BY block_id LIMIT 1;`, utils.Config.Chain.Id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) { // no entries = no gaps
				return []intRange{}, nil
			}
			return []intRange{}, fmt.Errorf("error reading first block from postgres: %w", err)
		}
		if firstBlock != 0 {
			gaps = append(gaps, intRange{start: 0, end: firstBlock - 1})
		}
	}

	// check for gaps everywhere else
	rows, err := db.ReaderDb.Query(`
		SELECT 
			block_id + 1 as gapStart, 
			nextNumber - 1 as gapEnd
		FROM 
			(
			SELECT 
				block_id, LEAD(block_id) OVER (ORDER BY block_id) as nextNumber
			FROM
				raw_block_status
			WHERE
				chain_id = $1
			) number
		WHERE 
			block_id + 1 <> nextNumber
		ORDER BY
			gapStart;`, utils.Config.Chain.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return gaps, nil
		}
		return []intRange{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var gap intRange
		err := rows.Scan(&gap.start, &gap.end)
		if err != nil {
			return []intRange{}, err
		}
		gaps = append(gaps, gap)
	}

	return gaps, nil
}

// get latest block in postgres db
func psqlGetLatestBlock(useWriterDb bool) (int64, error) {
	var err error
	var latestBlock int64
	query := `SELECT block_id FROM raw_block_status WHERE chain_id = $1 ORDER BY block_id DESC LIMIT 1;`
	if useWriterDb {
		err = db.WriterDb.Get(&latestBlock, query, utils.Config.Chain.Id)
	} else {
		err = db.ReaderDb.Get(&latestBlock, query, utils.Config.Chain.Id)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, nil
		}
		return -1, fmt.Errorf("error reading latest block in postgres: %w", err)
	}
	return latestBlock, nil
}

// will add elements to sql, based on blockRawData
// on conflict, it will only overwrite / change current entry if hash is different
func psqlAddElements(blockRawData []fullBlockRawData) error {
	l := len(blockRawData)
	if l <= 0 {
		return fmt.Errorf("error, got empty blockRawData array (%d)", l)
	}

	block_number := make([]int64, l)
	block_hash := make(pq.ByteaArray, l)
	for i, v := range blockRawData {
		block_number[i] = v.blockNumber
		block_hash[i] = v.blockHash
	}

	_, err := db.WriterDb.Exec(`
		INSERT INTO raw_block_status
			(chain_id, block_id, block_hash)
		SELECT
			$1,
			UNNEST($2::int[]),
			UNNEST($3::bytea[][])
		ON CONFLICT (chain_id, block_id) DO
			UPDATE SET
				block_hash = excluded.block_hash,
				indexed_bt = FALSE
			WHERE
				raw_block_status.block_hash != excluded.block_hash;`,
		utils.Config.Chain.Id, pq.Array(block_number), block_hash)
	return err
}

// will return a list of all provided block_ids where the hash in the database matches the provided list
func psqlGetHashHitsIdList(blockRawData []fullBlockRawData) ([]int64, error) {
	l := len(blockRawData)
	if l <= 0 {
		return nil, fmt.Errorf("error, got empty blockRawData array (%d)", l)
	}

	block_number := make([]int64, l)
	block_hash := make(pq.ByteaArray, l)
	for i, v := range blockRawData {
		block_number[i] = v.blockNumber
		block_hash[i] = v.blockHash
	}

	// as there are corner cases, to be on the safe side, we will use WriterDb here
	rows, err := db.WriterDb.Query(`
		SELECT 
			raw_block_status.block_id 
		FROM 
			raw_block_status, 
			(SELECT UNNEST($1::int[]) as block_id, UNNEST($2::bytea[][]) as block_hash) as node_block_status 
		WHERE
			chain_id = $3
			AND
			raw_block_status.block_id = node_block_status.block_id 
			AND 
			raw_block_status.block_hash = node_block_status.block_hash 
		ORDER 
			by raw_block_status.block_id;`,
		pq.Array(block_number), block_hash, utils.Config.Chain.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []int64{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	result := []int64{}
	for rows.Next() {
		var block_id int64
		err := rows.Scan(&block_id)
		if err != nil {
			return nil, err
		}
		result = append(result, block_id)
	}

	return result, nil
}

// ////////////////
// RPC interface //
// ////////////////
// get chain id from node
func rpciGetChainId() (uint64, error) {
	chainId, err := elClient.ChainID(context.Background())
	if err != nil {
		return 0, fmt.Errorf("error retrieving chain id from node: %w", err)
	}
	return chainId.Uint64(), nil
}

// get latest block number from node
func rpciGetLatestBlock() (int64, error) {
	latestBlockNumber, err := elClient.BlockNumber(context.Background())
	if err != nil {
		return 0, fmt.Errorf("error retrieving latest block number: %w", err)
	}
	return int64(latestBlockNumber), nil
}

// subscribe for latest block
func rpciSubscribeNewHead(ch chan<- *gethtypes.Header) (ethereum.Subscription, error) {
	return elClient.SubscribeNewHead(context.Background(), ch)
}

// do all the http stuff
func _rpciGetHttpResult(body []byte, nodeRequestsAtOnce int64, count int64) ([][]byte, error) {
	/*
		funny thing: EL call can crash with 'batch response exceeded limit of 10000000 bytes' even if there is only 1 element.
		so to avoid a dead lock, we will remove the "batch" and make a single request, which doesn't have this error at all.
		imho this is a EL bug and should be fixed there, but meanwhile this "workaround" will be fine as well.
	*/
	_ = nodeRequestsAtOnce
	if count == 1 && len(body) > 2 && body[0] == byte('[') && body[len(body)-1] == byte(']') {
		body = body[1 : len(body)-1]
	}

	r, err := http.NewRequest(http.MethodPost, eth1RpcEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating post request: %w", err)
	}

	r.Header.Add("Content-Type", "application/json")
	res, err := httpClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("error executing post request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error unexpected status code: %d", res.StatusCode)
	}

	defer res.Body.Close()
	resByte, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading request body: %w", err)
	}

	errorToCheck := []byte(`"error":{"code"`)
	if bytes.Contains(resByte, errorToCheck) {
		const keepDigitsTotal = 1000
		const keepDigitsFront = 100
		if len(resByte) > keepDigitsTotal {
			i := bytes.Index(resByte, errorToCheck)
			if i >= keepDigitsFront {
				resByte = append([]byte(`<...>`), resByte[i-keepDigitsFront:]...)
			}
			if len(resByte) > keepDigitsTotal {
				resByte = append(resByte[:keepDigitsTotal-5], []byte(`<...>`)...)
			}
		}
		return nil, fmt.Errorf("rpc error: %s", resByte)
	}

	return _splitAndVerifyJsonArray(resByte, count)
}

// will fill only receipts_compressed based on block, used by rpciGetBulkRawReceipts function
func _rpciGetBulkRawBlockReceipts(blockRawData []fullBlockRawData, nodeRequestsAtOnce int64) error {
	// check
	{
		l := int64(len(blockRawData))
		if l < 1 {
			return fmt.Errorf("empty blockRawData array received")
		}
		if l > nodeRequestsAtOnce {
			return fmt.Errorf("blockRawData array received with more elements (%d) than allowed (%d)", l, nodeRequestsAtOnce)
		}
	}

	// get array
	var rawData [][]byte
	{
		bodyStr := "["
		for i, v := range blockRawData {
			if i != 0 {
				bodyStr += ","
			}
			bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockReceipts","params":["0x%x"],"id":%d}`, v.blockNumber, i)
		}
		bodyStr += "]"
		var err error
		rawData, err = _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, int64(len(blockRawData)))
		if err != nil {
			return fmt.Errorf("error (_rpciGetBulkRawBlockReceipts) split and verify json array: %w", err)
		}
	}

	// get data
	for i, v := range rawData {
		blockRawData[i].receiptsCompressed = compress(v)
	}

	return nil
}

// will fill only receipts_compressed based on transaction, used by rpciGetBulkRawReceipts function
func _rpciGetBulkRawTransactionReceipts(blockRawData []fullBlockRawData, nodeRequestsAtOnce int64) error {
	// check
	{
		l := int64(len(blockRawData))
		if l < 1 {
			return fmt.Errorf("empty blockRawData array received")
		}
		if l > nodeRequestsAtOnce {
			return fmt.Errorf("blockRawData array received with more elements (%d) than allowed (%d)", l, nodeRequestsAtOnce)
		}
	}

	// iterate through array and get data when threshold reached
	var blockRawDataWriteIndex int
	var currentElementCount int64
	var dataAvailable bool
	bodyStr := "["
	for i, v := range blockRawData {
		l := int64(len(v.blockTxs))
		if l < 1 {
			continue // skip empty
		}

		// threshold reached, getting data...
		if dataAvailable {
			if currentElementCount+l > nodeRequestsAtOnce {
				bodyStr += "]"
				rawData, err := _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, currentElementCount)
				if err != nil {
					return fmt.Errorf("error (_rpciGetBulkRawTransactionReceipts) split and verify json array: %w", err)
				}

				for _, vv := range rawData {
					for len(blockRawData[blockRawDataWriteIndex].blockTxs) < 1 {
						blockRawDataWriteIndex++
					}
					blockRawData[blockRawDataWriteIndex].receiptsCompressed = compress(vv)
					blockRawDataWriteIndex++
				}

				currentElementCount = 0
				bodyStr = "["
			} else {
				bodyStr += ","
			}
		}

		// adding txs of current block
		dataAvailable = true
		currentElementCount += l
		for txIndex, txValue := range v.blockTxs {
			if txIndex != 0 {
				bodyStr += ","
			}
			bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["%s"],"id":%d}`, txValue, i)
		}
	}

	// getting data for the rest...
	{
		bodyStr += "]"
		if len(bodyStr) > 2 {
			rawData, err := _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, currentElementCount)
			if err != nil {
				return fmt.Errorf("error (_rpciGetBulkRawTransactionReceipts) split and verify json array: %w", err)
			}

			for _, vv := range rawData {
				for len(blockRawData[blockRawDataWriteIndex].blockTxs) < 1 {
					blockRawDataWriteIndex++
				}
				blockRawData[blockRawDataWriteIndex].receiptsCompressed = compress(vv)
				blockRawDataWriteIndex++
			}
		}
	}

	return nil
}

// will fill only block_hash, block_unclesCount, block_compressed & block_txs
func rpciGetBulkBlockRawData(blockRawData []fullBlockRawData, nodeRequestsAtOnce int64) error {
	// check
	{
		l := int64(len(blockRawData))
		if l < 1 {
			return fmt.Errorf("empty blockRawData array received")
		}
		if l > nodeRequestsAtOnce {
			return fmt.Errorf("blockRawData array received with more elements (%d) than allowed (%d)", l, nodeRequestsAtOnce)
		}
	}

	// get array
	var rawData [][]byte
	{
		bodyStr := "["
		for i, v := range blockRawData {
			if i != 0 {
				bodyStr += ","
			}
			bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", true],"id":%d}`, v.blockNumber, i)
		}
		bodyStr += "]"
		var err error
		rawData, err = _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, int64(len(blockRawData)))
		if err != nil {
			return fmt.Errorf("error (rpciGetBulkBlockRawData) split and verify json array: %w", err)
		}
	}

	// get data
	blockParsed := &types.Eth1RpcGetBlockResponse{}
	for i, v := range rawData {
		// block
		{
			blockRawData[i].blockCompressed = compress(v)
			err := json.Unmarshal(v, blockParsed)
			if err != nil {
				return fmt.Errorf("error decoding block '%d' response: %w", blockRawData[i].blockNumber, err)
			}
		}

		// id
		if i != blockParsed.Id {
			return fmt.Errorf("impossible error, i '%d' doesn't match blockParsed.Id '%d'", i, blockParsed.Id)
		}

		// number
		{
			blockParsedResultNumber := int64(binary.BigEndian.Uint64(append(make([]byte, 8-len(blockParsed.Result.Number)), blockParsed.Result.Number...)))
			if blockRawData[i].blockNumber != blockParsedResultNumber {
				log.Error(nil, "Doesn't match", 0, map[string]interface{}{"blockRawData[i].blockNumber": blockRawData[i].blockNumber, "blockParsedResultNumber": blockParsedResultNumber})
			}
		}

		// hash
		if blockParsed.Result.Hash == nil {
			return fmt.Errorf("blockParsed.Result.Hash is nil at block '%d'", blockRawData[i].blockNumber)
		}
		blockRawData[i].blockHash = blockParsed.Result.Hash

		// transaction
		if blockParsed.Result.Transactions == nil {
			return fmt.Errorf("blockParsed.Result.Transactions is nil at block '%d'", blockRawData[i].blockNumber)
		}
		blockRawData[i].blockTxs = make([]string, len(blockParsed.Result.Transactions))
		for ii, tx := range blockParsed.Result.Transactions {
			blockRawData[i].blockTxs[ii] = tx.Hash.String()
		}

		// uncle count
		if blockParsed.Result.Uncles != nil {
			blockRawData[i].blockUnclesCount = len(blockParsed.Result.Uncles)
		}
	}

	return nil
}

// will fill only block_hash
func rpciGetBulkBlockRawHash(blockRawData []fullBlockRawData, nodeRequestsAtOnce int64) error {
	// check
	{
		l := int64(len(blockRawData))
		if l < 1 {
			return fmt.Errorf("empty blockRawData array received")
		}
		if l > 1 && l > nodeRequestsAtOnce {
			err := rpciGetBulkBlockRawHash(blockRawData[:l/2], nodeRequestsAtOnce)
			if err == nil {
				err = rpciGetBulkBlockRawHash(blockRawData[l/2:], nodeRequestsAtOnce)
			}
			return err
		}
	}

	// get array
	var rawData [][]byte
	{
		bodyStr := "["
		for i, v := range blockRawData {
			if i != 0 {
				bodyStr += ","
			}
			bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", true],"id":%d}`, v.blockNumber, i)
		}
		bodyStr += "]"
		var err error
		rawData, err = _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, int64(len(blockRawData)))
		if err != nil {
			return fmt.Errorf("error (rpciGetBulkBlockRawHash) split and verify json array: %w", err)
		}
	}

	// get data
	blockParsed := &types.Eth1RpcGetBlockResponse{}
	for i, v := range rawData {
		err := json.Unmarshal(v, blockParsed)
		if err != nil {
			return fmt.Errorf("error decoding block '%d' response: %w", blockRawData[i].blockNumber, err)
		}
		if i != blockParsed.Id {
			return fmt.Errorf("impossible error, i '%d' doesn't match blockParsed.Id '%d'", i, blockParsed.Id)
		}
		{
			blockParsedResultNumber := int64(binary.BigEndian.Uint64(append(make([]byte, 8-len(blockParsed.Result.Number)), blockParsed.Result.Number...)))
			if blockRawData[i].blockNumber != blockParsedResultNumber {
				log.Error(nil, "Doesn't match", 0, map[string]interface{}{"blockRawData[i].blockNumber": blockRawData[i].blockNumber, "blockParsedResultNumber": blockParsedResultNumber})
			}
		}
		if blockParsed.Result.Hash == nil {
			return fmt.Errorf("blockParsed.Result.Hash is nil at block '%d'", blockRawData[i].blockNumber)
		}
		blockRawData[i].blockHash = blockParsed.Result.Hash
	}

	return nil
}

// will fill only uncles (if available)
func rpciGetBulkRawUncles(blockRawData []fullBlockRawData, nodeRequestsAtOnce int64) error {
	// check
	{
		l := int64(len(blockRawData))
		if l < 1 {
			return fmt.Errorf("empty blockRawData array received")
		}
		if l > nodeRequestsAtOnce {
			// I know, in the case of uncles, it's very unlikly that we need all slots, but handling this separate, would be way to much, so whatever
			return fmt.Errorf("blockRawData array received with more elements (%d) than allowed (%d)", l, nodeRequestsAtOnce)
		}
	}

	// get array
	var rawData [][]byte
	{
		requestedCount := int64(0)
		firstElement := true
		bodyStr := "["
		for _, v := range blockRawData {
			if v.blockUnclesCount > 2 || v.blockUnclesCount < 0 {
				// fatal, as this is an impossible error
				log.Fatal(nil, "impossible error, found impossible uncle count, expected 0, 1 or 2", 0, map[string]interface{}{"block_unclesCount": v.blockUnclesCount, "block_number": v.blockNumber})
			} else if v.blockUnclesCount == 0 {
				continue
			} else {
				if firstElement {
					firstElement = false
				} else {
					bodyStr += ","
				}
				if v.blockUnclesCount == 1 {
					requestedCount++
					bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getUncleByBlockNumberAndIndex","params":["0x%x", "0x0"],"id":%d}`, v.blockNumber, v.blockNumber)
				} else /* if v.block_unclesCount == 2 */ {
					requestedCount++
					bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getUncleByBlockNumberAndIndex","params":["0x%x", "0x0"],"id":%d},`, v.blockNumber, v.blockNumber)
					requestedCount++
					bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getUncleByBlockNumberAndIndex","params":["0x%x", "0x1"],"id":%d}`, v.blockNumber, v.blockNumber)
				}
			}
		}
		bodyStr += "]"
		if requestedCount == 0 { // nothing todo, no uncles in set
			return nil
		}

		var err error
		rawData, err = _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, requestedCount)
		if err != nil {
			return fmt.Errorf("error (rpciGetBulkRawUncles) split and verify json array: %w", err)
		}
	}

	// get data
	rdIndex := 0
	for i, v := range blockRawData {
		if v.blockUnclesCount > 0 { // Not the prettiest way, but the unmarshal would take much longer with the same result
			blockRawData[i].unclesCompressed = compress(rawData[rdIndex])
			rdIndex++
		}
	}

	return nil
}

// will fill only receipts_compressed
func rpciGetBulkRawReceipts(blockRawData []fullBlockRawData, nodeRequestsAtOnce int64) error {
	if utils.Config.Chain.Id == ARBITRUM_CHAINID {
		return _rpciGetBulkRawTransactionReceipts(blockRawData, nodeRequestsAtOnce)
	}
	return _rpciGetBulkRawBlockReceipts(blockRawData, nodeRequestsAtOnce)
}

// will fill only traces_compressed
func rpciGetBulkRawTraces(blockRawData []fullBlockRawData, nodeRequestsAtOnce int64) error {
	// check
	{
		l := int64(len(blockRawData))
		if l < 1 {
			return fmt.Errorf("empty blockRawData array received")
		}
		if l > nodeRequestsAtOnce {
			return fmt.Errorf("blockRawData array received with more elements (%d) than allowed (%d)", l, nodeRequestsAtOnce)
		}
	}

	// get array
	var rawData [][]byte
	{
		bodyStr := "["
		for i, v := range blockRawData {
			if i != 0 {
				bodyStr += ","
			}
			if utils.Config.Chain.Id == ARBITRUM_CHAINID && v.blockNumber < ARBITRUM_NITRO_BLOCKNUMBER {
				bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"arbtrace_block","params":["0x%x"],"id":%d}`, v.blockNumber, i)
			} else {
				bodyStr += fmt.Sprintf(`{"jsonrpc":"2.0","method":"debug_traceBlockByNumber","params":["0x%x", {"tracer": "callTracer"}],"id":%d}`, v.blockNumber, i)
			}
		}
		bodyStr += "]"
		var err error
		rawData, err = _rpciGetHttpResult([]byte(bodyStr), nodeRequestsAtOnce, int64(len(blockRawData)))
		if err != nil {
			return fmt.Errorf("error (rpciGetBulkRawTraces) split and verify json array: %w", err)
		}
	}

	// get data
	for i, v := range rawData {
		blockRawData[i].tracesCompressed = compress(v)
	}

	return nil
}
