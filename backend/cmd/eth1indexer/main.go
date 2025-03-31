package eth1indexer

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gobitfly/beaconchain/pkg/executionlayer"
	"github.com/gobitfly/beaconchain/pkg/executionlayer/evm"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	_ "github.com/jackc/pgx/v5/stdlib"

	//nolint:gosec
	_ "net/http/pprof"
)

func Run() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	erigonEndpoint := fs.String("erigon", "", "Erigon archive node enpoint")
	block := fs.Uint64("block", 0, "Index a specific block")

	reorgDepth := fs.Uint64("reorg.depth", 20, "Lookback to check and handle chain reorgs")

	concurrencyBlocks := fs.Uint64("blocks.concurrency", 30, "Concurrency to use when indexing blocks from erigon")
	startBlocks := fs.Uint64("blocks.start", 0, "Block to start indexing")
	endBlocks := fs.Uint64("blocks.end", 0, "Block to finish indexing")
	bulkBlocks := fs.Uint64("blocks.bulk", 8000, "Maximum number of blocks to be processed before saving")
	offsetBlocks := fs.Uint64("blocks.offset", 100, "Blocks offset")
	checkBlocksGaps := fs.Bool("blocks.gaps", false, "Check for gaps in the blocks table")
	checkBlocksGapsLookback := fs.Int("blocks.gaps.lookback", 1000000, "Lookback for gaps check of the blocks table")
	traceMode := fs.String("blocks.tracemode", "parity/geth", "Trace mode to use, can bei either 'parity', 'geth' or 'parity/geth' for both")

	concurrencyData := fs.Uint64("data.concurrency", 30, "Concurrency to use when indexing data from bigtable")
	startData := fs.Uint64("data.start", 0, "Block to start indexing")
	endData := fs.Uint64("data.end", 0, "Block to finish indexing")
	bulkData := fs.Uint64("data.bulk", 8000, "Maximum number of blocks to be processed before saving")
	offsetData := fs.Uint64("data.offset", 1000, "Data offset")
	checkDataGaps := fs.Bool("data.gaps", false, "Check for gaps in the data table")
	checkDataGapsLookback := fs.Int("data.gaps.lookback", 1000000, "Lookback for gaps check of the blocks table")

	enableBalanceUpdater := fs.Bool("balances.enabled", false, "Enable balance update process")
	enableFullBalanceUpdater := fs.Bool("balances.full.enabled", false, "Enable full balance update process")
	balanceUpdaterBatchSize := fs.Int("balances.batch", 1000, "Batch size for balance updates")

	tokenPriceExport := fs.Bool("token.price.enabled", false, "Enable token export process")
	tokenPriceExportList := fs.String("token.price.list", "", "Tokenlist path to use for the token price export")
	tokenPriceExportFrequency := fs.Duration("token.price.frequency", time.Hour, "Token price export interval")

	versionFlag := fs.Bool("version", false, "Print version and exit")

	configPath := fs.String("config", "", "Path to the config file, if empty string defaults will be used")

	enableEnsUpdater := fs.Bool("ens.enabled", false, "Enable ens update process")
	ensBatchSize := fs.Int64("ens.batch", 200, "Batch size for ens updates")

	_ = fs.Parse(os.Args[2:])

	log.Info(*configPath)
	if *versionFlag {
		log.Info(version.Version)
		log.Info(version.GoVersion)
		return
	}

	cfg := &types.Config{}
	err := utils.ReadConfig(cfg, *configPath)
	if err != nil {
		log.Fatal(err, "error reading config file", 0)
	}
	utils.Config = cfg

	log.InfoWithFields(log.Fields{"config": *configPath, "version": version.Version, "chainName": utils.Config.Chain.ClConfig.ConfigName}, "starting")

	if utils.Config.Metrics.Enabled {
		go func() {
			log.Infof("serving metrics on %v", utils.Config.Metrics.Address)
			if err := metrics.Serve(utils.Config.Metrics.Address, utils.Config.Metrics.Pprof, utils.Config.Metrics.PprofExtra); err != nil {
				log.Fatal(err, "error serving metrics", 0)
			}
		}()
	}

	// enable pprof endpoint if requested
	if utils.Config.Pprof.Enabled {
		go func() {
			log.Infof("starting pprof http server on port %s", utils.Config.Pprof.Port)
			server := &http.Server{
				Addr:         fmt.Sprintf("localhost:%s", utils.Config.Pprof.Port),
				Handler:      nil,
				ReadTimeout:  60 * time.Second,
				WriteTimeout: 60 * time.Second,
			}
			err := server.ListenAndServe()

			if err != nil {
				log.Error(err, "error during ListenAndServe for pprof http server", 0)
			}
		}()
	}

	db.WriterDb, db.ReaderDb = db.MustInitDB(&cfg.WriterDatabase, &cfg.ReaderDatabase, "pgx", "postgres")
	defer db.ReaderDb.Close()
	defer db.WriterDb.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:        utils.Config.RedisCacheEndpoint,
		ReadTimeout: time.Second * 20,
	})

	if erigonEndpoint == nil || *erigonEndpoint == "" {
		if utils.Config.Eth1ErigonEndpoint == "" {
			log.Fatal(nil, "no erigon node url provided", 0)
		} else {
			log.Infof("applying erigon endpoint from config")
			*erigonEndpoint = utils.Config.Eth1ErigonEndpoint
		}
	}

	log.Infof("using erigon node at %v", *erigonEndpoint)
	client, err := rpc.NewErigonClient(*erigonEndpoint)
	if err != nil {
		log.Fatal(err, "erigon client creation error", 0)
	}

	chainId := strconv.FormatUint(utils.Config.Chain.ClConfig.DepositChainID, 10)

	nodeChainId, err := client.GetNativeClient().ChainID(context.Background())
	if err != nil {
		log.Fatal(err, "node chain id error", 0)
	}

	if nodeChainId.String() != chainId {
		log.Fatal(fmt.Errorf("node chain id mismatch, wanted %v got %v", chainId, nodeChainId.String()), "", 0)
	}

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, chainId, utils.Config.RedisCacheEndpoint)
	if err != nil {
		log.Fatal(err, "error connecting to bigtable", 0)
	}
	defer bt.Close()

	bigtable, err := database.NewBigTable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, nil)
	if err != nil {
		log.Fatal(err, "error connecting to bigtable", 0)
	}

	cache := freecache.NewCache(100 * 1024 * 1024) // 100 MB limit
	store := db2.NewStoreV1FromBigtable(bigtable, database.FreeCache{Cache: cache})

	batcherConfig := evm.BatcherConfig{
		Limit: utils.Config.Indexer.BatchLimit,
	}
	if utils.Config.Indexer.MulticallAddresses != "" {
		parsed := common.HexToAddress(utils.Config.Indexer.MulticallAddresses)
		batcherConfig.MulticallAddress = &parsed
	}
	batcher := evm.NewBatcher(nodeChainId, client.GetNativeClient(), batcherConfig)

	if *tokenPriceExport {
		go func() {
			for {
				tokenList, err := readTokenListFile(*tokenPriceExportList)
				if err != nil {
					log.Error(err, "error reading token list file", 0)
				}
				pricer := executionlayer.NewTokenPricer(
					store,
					chainId,
					executionlayer.NewLlamaClient(),
					tokenList,
					batcher,
				)
				if err := pricer.UpdateTokens(); err != nil {
					log.Error(err, "error updating tokens", 0)
				}
				time.Sleep(*tokenPriceExportFrequency)
			}
		}()
	}

	if *enableEnsUpdater {
		importer := executionlayer.NewENSImporter(store, db2.NewENSStore(db.WriterDb), executionlayer.NewEnsContracts(client.GetNativeClient()))
		go ImportEnsUpdatesLoop(chainId, importer, *ensBatchSize)
	}

	lastBlockStore := db2.NewCachedLastBlocks(database.Redis{Client: redisClient}, store)
	indexer := executionlayer.NewIndexer(store, lastBlockStore, executionlayer.AllTransformers...)
	balanceUpdater := executionlayer.NewBalanceUpdater(chainId, store, store, batcher)
	reorgWatcher := executionlayer.NewReorgWatcher(client.GetNativeClient(), store, *reorgDepth, chainId, lastBlockStore)

	if *enableFullBalanceUpdater {
		ProcessBalanceUpdates(balanceUpdater, *balanceUpdaterBatchSize, -1)
		return
	}

	if *block != 0 {
		if err := indexer.IndexNode(chainId, client, *block, *block, *concurrencyBlocks, *traceMode); err != nil {
			log.Fatal(err, "error indexing from node", 0, map[string]interface{}{"block": *block, "concurrency": *concurrencyBlocks})
		}
		if err := indexer.IndexEvents(chainId, *block, *block, *concurrencyData); err != nil {
			log.Fatal(err, "error indexing from bigtable", 0)
		}
		cache.Clear()

		log.Infof("indexing of block %v completed", *block)
		return
	}

	if *checkBlocksGaps {
		_, _, _, err := bt.CheckForGapsInBlocksTable(*checkBlocksGapsLookback)

		if err != nil {
			log.Fatal(err, "error checking for gaps in blocks table", 0)
		}
		return
	}

	if *checkDataGaps {
		err := bt.CheckForGapsInDataTable(*checkDataGapsLookback)
		if err != nil {
			log.Fatal(err, "error checking for gapis in data table", 0)
		}
		return
	}

	if *endBlocks != 0 && *startBlocks < *endBlocks {
		if err = indexer.IndexNode(chainId, client, *startBlocks, *endBlocks, *concurrencyBlocks, *traceMode); err != nil {
			log.Fatal(err, "error indexing from node", 0, map[string]interface{}{"start": *startBlocks, "end": *endBlocks, "concurrency": *concurrencyBlocks})
		}
		return
	}

	if *endData != 0 && *startData < *endData {
		if err := indexer.IndexEvents(chainId, *startData, *endData, *concurrencyData); err != nil {
			log.Fatal(err, "error indexing from bigtable", 0)
		}
		cache.Clear()
		return
	}

	lastSuccessfulBlockIndexingTs := time.Now()
	for ; ; time.Sleep(time.Second * 14) {
		if err := reorgWatcher.LookForReorg(); err != nil {
			log.Error(err, "error handling chain reorg", 0)
			continue
		}

		lastBlockFromNode, err := client.GetLatestEth1BlockNumber()
		if err != nil {
			log.Error(err, "error retrieving latest eth block number", 0)
			continue
		}

		lastBlockFromBlocksTable, err := lastBlockStore.GetInBlocksTable(chainId)
		if err != nil {
			log.Error(err, "error retrieving last blocks from blocks table", 0)
			continue
		}

		lastBlockFromDataTable, err := lastBlockStore.GetInDataTable(chainId)
		if err != nil {
			log.Error(err, "error retrieving last blocks from data table", 0)
			continue
		}

		log.InfoWithFields(log.Fields{
			"node":   lastBlockFromNode,
			"blocks": lastBlockFromBlocksTable,
			"data":   lastBlockFromDataTable,
		}, "last blocks")

		continueAfterError := false
		if lastBlockFromNode > 0 {
			if lastBlockFromBlocksTable < lastBlockFromNode {
				log.Infof("missing blocks %v to %v in blocks table, indexing ...", lastBlockFromBlocksTable+1, lastBlockFromNode)

				startBlock := uint64(0)
				if int64(lastBlockFromDataTable+1-*offsetData) > 0 {
					startBlock = lastBlockFromBlocksTable + 1 - *offsetBlocks
				}

				if *bulkBlocks <= 0 || *bulkBlocks > lastBlockFromNode-startBlock+1 {
					*bulkBlocks = lastBlockFromNode - startBlock + 1
				}

				for startBlock <= lastBlockFromNode && !continueAfterError {
					endBlock := startBlock + *bulkBlocks - 1
					if endBlock > lastBlockFromNode {
						endBlock = lastBlockFromNode
					}

					err = indexer.IndexNode(chainId, client, startBlock, endBlock, *concurrencyBlocks, *traceMode)
					if err != nil {
						errMsg := "error indexing from node"
						errFields := map[string]interface{}{
							"start":       startBlock,
							"end":         endBlock,
							"concurrency": *concurrencyBlocks}
						if time.Since(lastSuccessfulBlockIndexingTs) > time.Minute*30 {
							log.Fatal(err, errMsg, 0, errFields)
						} else {
							log.Error(err, errMsg, 0, errFields)
						}
						continueAfterError = true
						continue
					} else {
						lastSuccessfulBlockIndexingTs = time.Now()
					}

					startBlock = endBlock + 1
				}
				if continueAfterError {
					continue
				}
			}

			if lastBlockFromDataTable < lastBlockFromNode {
				log.Infof("missing blocks %v to %v in data table, indexing ...", lastBlockFromDataTable+1, lastBlockFromNode)

				startBlock := uint64(0)
				if int64(lastBlockFromDataTable+1-*offsetData) > 0 {
					startBlock = lastBlockFromDataTable + 1 - *offsetData
				}

				if *bulkData <= 0 || *bulkData > lastBlockFromNode-startBlock+1 {
					*bulkData = lastBlockFromNode - startBlock + 1
				}

				for startBlock <= lastBlockFromNode && !continueAfterError {
					endBlock := startBlock + *bulkData - 1
					if endBlock > lastBlockFromNode {
						endBlock = lastBlockFromNode
					}

					if err := indexer.IndexEvents(chainId, startBlock, endBlock, *concurrencyBlocks); err != nil {
						log.Error(err, "error indexing from bigtable", 0, map[string]interface{}{"start": startBlock, "end": endBlock, "concurrency": *concurrencyData})
						cache.Clear()
						continueAfterError = true
						continue
					}
					cache.Clear()

					startBlock = endBlock + 1
				}
				if continueAfterError {
					continue
				}
			}
		}

		if *enableBalanceUpdater {
			ProcessBalanceUpdates(balanceUpdater, *balanceUpdaterBatchSize, 10)
		}

		log.Infof("index run completed")
		services.ReportStatus("eth1indexer", "Running", nil)
	}

	// utils.WaitForCtrlC()
}

func ImportEnsUpdatesLoop(chainID string, importer executionlayer.ENSImporter, batchSize int64) {
	for {
		time.Sleep(time.Second * 5)
		if err := importer.Import(chainID, batchSize); err != nil {
			log.Error(err, "error importing ens updates", 0, nil)
			continue
		}
		services.ReportStatus("ensIndexer", "Running", nil)
	}
}

func readTokenListFile(path string) (erc20.ERC20TokenList, error) {
	tokenListContent, err := os.ReadFile(path)
	if err != nil {
		return erc20.ERC20TokenList{}, err
	}
	var tokenList erc20.ERC20TokenList
	if err := json.Unmarshal(tokenListContent, &tokenList); err != nil {
		return erc20.ERC20TokenList{}, err
	}
	return tokenList, nil
}

// ProcessBalanceUpdates will use the balanceUpdater to fetch and update the balances
// if iterations == -1 it will run forever
func ProcessBalanceUpdates(balanceUpdater executionlayer.BalanceUpdater, batchSize int, iterations int) {
	for its := 0; iterations == -1 || its < iterations; its++ {
		start := time.Now()
		balances, err := balanceUpdater.UpdateBalances(int64(batchSize))
		if err != nil {
			log.Error(err, "error updating balances", 0)
			return
		}
		log.Infof("retrieved %v balances in %v, currently at %s", len(balances), time.Since(start), balances[len(balances)-1].Address)
	}
}
