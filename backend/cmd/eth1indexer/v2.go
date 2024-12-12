package eth1indexer

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/coocood/freecache"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/gobitfly/beaconchain/pkg/commons/db2/data"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/raw"
	"github.com/gobitfly/beaconchain/pkg/commons/indexer"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"
)

func Run2() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	erigonEndpoint := fs.String("erigon", "", "Erigon archive node enpoint")
	block := fs.Int64("block", 0, "Index a specific block")

	// reorgDepth := fs.Int("reorg.depth", 20, "Lookback to check and handle chain reorgs")

	// concurrencyBlocks := fs.Int64("blocks.concurrency", 30, "Concurrency to use when indexing blocks from erigon")
	startBlocks := fs.Int64("blocks.start", 0, "Block to start indexing")
	endBlocks := fs.Int64("blocks.end", 0, "Block to finish indexing")
	// bulkBlocks := fs.Int64("blocks.bulk", 8000, "Maximum number of blocks to be processed before saving")
	// offsetBlocks := fs.Int64("blocks.offset", 100, "Blocks offset")
	// checkBlocksGaps := fs.Bool("blocks.gaps", false, "Check for gaps in the blocks table")
	// checkBlocksGapsLookback := fs.Int("blocks.gaps.lookback", 1000000, "Lookback for gaps check of the blocks table")
	// traceMode := fs.String("blocks.tracemode", "parity/geth", "Trace mode to use, can bei either 'parity', 'geth' or 'parity/geth' for both")

	// concurrencyData := fs.Int64("data.concurrency", 30, "Concurrency to use when indexing data from bigtable")
	startData := fs.Int64("data.start", 0, "Block to start indexing")
	endData := fs.Int64("data.end", 0, "Block to finish indexing")
	// bulkData := fs.Int64("data.bulk", 8000, "Maximum number of blocks to be processed before saving")
	// offsetData := fs.Int64("data.offset", 1000, "Data offset")
	// checkDataGaps := fs.Bool("data.gaps", false, "Check for gaps in the data table")
	// checkDataGapsLookback := fs.Int("data.gaps.lookback", 1000000, "Lookback for gaps check of the blocks table")

	// enableBalanceUpdater := fs.Bool("balances.enabled", false, "Enable balance update process")
	// enableFullBalanceUpdater := fs.Bool("balances.full.enabled", false, "Enable full balance update process")
	// balanceUpdaterBatchSize := fs.Int("balances.batch", 1000, "Batch size for balance updates")

	// tokenPriceExport := fs.Bool("token.price.enabled", false, "Enable token export process")
	// tokenPriceExportList := fs.String("token.price.list", "", "Tokenlist path to use for the token price export")
	// tokenPriceExportFrequency := fs.Duration("token.price.frequency", time.Hour, "Token price export interval")

	versionFlag := fs.Bool("version", false, "Print version and exit")

	configPath := fs.String("config", "", "Path to the config file, if empty string defaults will be used")

	// enableEnsUpdater := fs.Bool("ens.enabled", false, "Enable ens update process")
	// ensBatchSize := fs.Int64("ens.batch", 200, "Batch size for ens updates")

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

	/*	db.WriterDb, db.ReaderDb = db.MustInitDB(&types.DatabaseConfig{
			Username:     cfg.WriterDatabase.Username,
			Password:     cfg.WriterDatabase.Password,
			Name:         cfg.WriterDatabase.Name,
			Host:         cfg.WriterDatabase.Host,
			Port:         cfg.WriterDatabase.Port,
			MaxOpenConns: cfg.WriterDatabase.MaxOpenConns,
			MaxIdleConns: cfg.WriterDatabase.MaxIdleConns,
			SSL:          cfg.WriterDatabase.SSL,
		}, &types.DatabaseConfig{
			Username:     cfg.ReaderDatabase.Username,
			Password:     cfg.ReaderDatabase.Password,
			Name:         cfg.ReaderDatabase.Name,
			Host:         cfg.ReaderDatabase.Host,
			Port:         cfg.ReaderDatabase.Port,
			MaxOpenConns: cfg.ReaderDatabase.MaxOpenConns,
			MaxIdleConns: cfg.ReaderDatabase.MaxIdleConns,
			SSL:          cfg.ReaderDatabase.SSL,
		}, "pgx", "postgres")
		defer db.ReaderDb.Close()
		defer db.WriterDb.Close()*/

	/*	if erigonEndpoint == nil || *erigonEndpoint == "" {
			if utils.Config.Eth1ErigonEndpoint == "" {
				log.Fatal(nil, "no erigon node url provided", 0)
			} else {
				log.Infof("applying erigon endpoint from config")
				*erigonEndpoint = utils.Config.Eth1ErigonEndpoint
			}
		}
	*/
	log.Infof("using erigon node at %v", *erigonEndpoint)
	client, err := rpc.NewErigonClient(*erigonEndpoint)
	if err != nil {
		log.Fatal(err, "erigon client creation error", 0)
	}

	chainId := strconv.FormatUint(utils.Config.Chain.ClConfig.DepositChainID, 10)

	rawStore := raw.NewStore(database.NewRemoteClient(cfg.RawBigtable.Remote))

	multiClient, err := rpc.NewMultiClient(*erigonEndpoint, rawStore)
	if err != nil {
		log.Fatal(err, "erigon client creation error", 0)
	}

	// balanceUpdaterPrefix := chainId + ":B:"

	nodeChainId, err := client.GetNativeClient().ChainID(context.Background())
	if err != nil {
		log.Fatal(err, "node chain id error", 0)
	}

	if nodeChainId.String() != chainId {
		log.Fatal(fmt.Errorf("node chain id mismatch, wanted %v got %v", chainId, nodeChainId.String()), "", 0)
	}

	/*srv, err := bttest.NewServer("localhost:0")
	if err != nil {
		log.Fatal(err, "", 0)
	}
	defer srv.Close()
	ctx := context.Background()
	conn, err := grpc.NewClient(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	project, instance := "proj", "instance"*/
	// ctx := context.Background()
	/*btAdminClient, _ := bigtable.NewAdminClient(ctx, "project", "instance")
	btClient, _ := bigtable.NewClient(ctx, "project", "instance")*/

	/*	bt, err := database.NewBigTable("test", "test", data.Schema)
		if err != nil {
			panic(err)
		}*/
	dataStore := data.NewStore(database.NewRemoteClient(cfg.Bigtable.Remote))
	// dataStore := data.NewStore(database.NewRemoteClient(cfg.Bigtable.Remote))

	cache := freecache.NewCache(100 * 1024 * 1024) // 100 MB limit
	transformer := indexer.NewTransformer(cache)
	indxr := indexer.New(dataStore, transformer.Tx, transformer.ERC20)

	var start, end int64
	if *block != 0 {
		start, end = *block, *block
	}
	if *endBlocks != 0 && *startBlocks < *endBlocks {
		start, end = *startBlocks, *endBlocks
	}
	if *endData != 0 && *startData < *endData {
		start, end = *startData, *endData
	}

	concurrency := 1
	batchSize := int64(25)
	g := errgroup.Group{}
	g.SetLimit(concurrency)

	for i := start; i <= end; i = i + batchSize {
		height := i
		heightEnd := height + batchSize - 1
		g.Go(func() error {
			if heightEnd > end {
				heightEnd = end
			}
			blocks, err := multiClient.GetBlocks(height, heightEnd, "geth")
			if err != nil {
				log.Fatal(err, "", 0)
			}
			if err := indxr.IndexBlocks(chainId, blocks); err != nil {
				log.Fatal(err, "", 0)
			}
			logrus.WithFields(map[string]interface{}{
				"start": height,
				"end":   heightEnd,
			}).Info("read blocks range")
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Fatal(err, "", 0)
	}

	/*lastSuccessulBlockIndexingTs := time.Now()
	for ; ; time.Sleep(time.Second * 14) {
		err := HandleChainReorgs(bt, client, *reorgDepth)
		if err != nil {
			log.Error(err, "error handling chain reorg", 0)
			continue
		}

		lastBlockFromNode, err := client.GetLatestEth1BlockNumber()
		if err != nil {
			log.Error(err, "error retrieving latest eth block number", 0)
			continue
		}

		lastBlockFromBlocksTable, err := bt.GetLastBlockInBlocksTable()
		if err != nil {
			log.Error(err, "error retrieving last blocks from blocks table", 0)
			continue
		}

		lastBlockFromDataTable, err := bt.GetLastBlockInDataTable()
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
			if lastBlockFromBlocksTable < int(lastBlockFromNode) {
				log.Infof("missing blocks %v to %v in blocks table, indexing ...", lastBlockFromBlocksTable+1, lastBlockFromNode)

				startBlock := int64(lastBlockFromBlocksTable+1) - *offsetBlocks
				if startBlock < 0 {
					startBlock = 0
				}

				if *bulkBlocks <= 0 || *bulkBlocks > int64(lastBlockFromNode)-startBlock+1 {
					*bulkBlocks = int64(lastBlockFromNode) - startBlock + 1
				}

				for startBlock <= int64(lastBlockFromNode) && !continueAfterError {
					endBlock := startBlock + *bulkBlocks - 1
					if endBlock > int64(lastBlockFromNode) {
						endBlock = int64(lastBlockFromNode)
					}

					err = IndexFromNode(bt, client, startBlock, endBlock, *concurrencyBlocks, *traceMode)
					if err != nil {
						errMsg := "error indexing from node"
						errFields := map[string]interface{}{
							"start":       startBlock,
							"end":         endBlock,
							"concurrency": *concurrencyBlocks}
						if time.Since(lastSuccessulBlockIndexingTs) > time.Minute*30 {
							log.Fatal(err, errMsg, 0, errFields)
						} else {
							log.Error(err, errMsg, 0, errFields)
						}
						continueAfterError = true
						continue
					} else {
						lastSuccessulBlockIndexingTs = time.Now()
					}

					startBlock = endBlock + 1
				}
				if continueAfterError {
					continue
				}
			}

			if lastBlockFromDataTable < int(lastBlockFromNode) {
				log.Infof("missing blocks %v to %v in data table, indexing ...", lastBlockFromDataTable+1, lastBlockFromNode)

				startBlock := int64(lastBlockFromDataTable+1) - *offsetData
				if startBlock < 0 {
					startBlock = 0
				}

				if *bulkData <= 0 || *bulkData > int64(lastBlockFromNode)-startBlock+1 {
					*bulkData = int64(lastBlockFromNode) - startBlock + 1
				}

				for startBlock <= int64(lastBlockFromNode) && !continueAfterError {
					endBlock := startBlock + *bulkData - 1
					if endBlock > int64(lastBlockFromNode) {
						endBlock = int64(lastBlockFromNode)
					}

					err = bt.IndexEventsWithTransformers(startBlock, endBlock, transforms, *concurrencyData, cache)
					if err != nil {
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
			ProcessMetadataUpdates(bt, client, balanceUpdaterPrefix, *balanceUpdaterBatchSize, 10)
		}

		log.Infof("index run completed")
		services.ReportStatus("eth1indexer", "Running", nil)
	}*/

	// utils.WaitForCtrlC()
}
