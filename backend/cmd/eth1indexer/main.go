package eth1indexer

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
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

	bulk := fs.Uint64("bulk", 8000, "Maximum number of blocks to be processed before saving")
	concurrency := fs.Uint64("concurrency", 30, "Concurrency to use when indexing blocks from erigon")
	traceMode := fs.String("blocks.tracemode", "parity/geth", "Trace mode to use, can bei either 'parity', 'geth' or 'parity/geth' for both")

	startBlocks := fs.Uint64("resync.start", 0, "Block to start indexing")
	endBlocks := fs.Uint64("resync.end", 0, "Block to finish indexing")
	skipBlocks := fs.Bool("resync.skipBlocks", false, "Skip resync for blocks table")
	skipData := fs.Bool("resync.skipData", false, "Skip resync for data table")

	checkBlocksGaps := fs.Bool("blocks.gaps", false, "Check for gaps in the blocks table")
	checkBlocksGapsLookback := fs.Int("blocks.gaps.lookback", 1000000, "Lookback for gaps check of the blocks table")

	checkDataGaps := fs.Bool("data.gaps", false, "Check for gaps in the data table")
	checkDataGapsLookback := fs.Int("data.gaps.lookback", 1000000, "Lookback for gaps check of the blocks table")

	balanceUpdaterBatchSize := fs.Int64("balances.batch", 1000, "Batch size for balance updates")

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

	chainID := strconv.FormatUint(utils.Config.Chain.ClConfig.DepositChainID, 10)

	nodeChainID, err := client.GetNativeClient().ChainID(context.Background())
	if err != nil {
		log.Fatal(err, "node chain id error", 0)
	}

	if nodeChainID.String() != chainID {
		log.Fatal(fmt.Errorf("node chain id mismatch, wanted %v got %v", chainID, nodeChainID.String()), "", 0)
	}

	bt, err := db.InitBigtable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, chainID, utils.Config.RedisCacheEndpoint)
	if err != nil {
		log.Fatal(err, "error connecting to bigtable", 0)
	}
	defer bt.Close()

	bigtable, err := database.NewBigTable(utils.Config.Bigtable.Project, utils.Config.Bigtable.Instance, nil)
	if err != nil {
		log.Fatal(err, "error connecting to bigtable", 0)
	}

	cache := freecache.NewCache(100 * 1024 * 1024) // 100 MB limit
	store := db2.NewStoreV1FromBigtable(bigtable, db2.CachedBalanceUpdates{RemoteCache: database.FreeCache{Cache: cache}})

	batcherConfig := evm.BatcherConfig{
		Limit: utils.Config.Indexer.BatchLimit,
	}
	if utils.Config.Indexer.MulticallAddresses != "" {
		parsed := common.HexToAddress(utils.Config.Indexer.MulticallAddresses)
		batcherConfig.MulticallAddress = &parsed
	}
	batcher := evm.NewBatcher(nodeChainID, client.GetNativeClient(), batcherConfig)

	lastBlockStore := db2.NewCachedLastBlocks(database.Redis{Client: redisClient}, store)
	indexer := executionlayer.NewIndexer(store, lastBlockStore, executionlayer.IndexerConfig{
		Concurrency: *concurrency,
		TraceMode:   *traceMode,
	}, client, executionlayer.AllTransformers...)
	balanceUpdater := executionlayer.NewBalanceUpdater(chainID, store, store, batcher)
	reorgWatcher := executionlayer.NewReorgWatcher(client.GetNativeClient(), store, *reorgDepth, chainID, lastBlockStore)

	pricer := executionlayer.NewTokenPricer(
		store,
		chainID,
		executionlayer.NewLlamaClient(),
		batcher,
	)

	var importer *executionlayer.ENSImporter
	if *enableEnsUpdater {
		importer = executionlayer.NewENSImporter(store, db2.NewENSStore(db.WriterDb), executionlayer.NewEnsContracts(client.GetNativeClient()))
	}

	start, end := uint64(0), uint64(0)
	if *block != 0 {
		start, end = *block, *block
	}
	if startBlocks != nil {
		start = *startBlocks
	}
	if endBlocks != nil {
		end = *endBlocks
	}

	service := executionlayer.NewIndexerService(
		nodeChainID.String(),
		client.GetNativeClient(),
		&balanceUpdater,
		reorgWatcher,
		lastBlockStore,
		pricer,
		indexer,
		db2.CachedBalanceUpdates{RemoteCache: database.FreeCache{Cache: cache}},
		store,
		importer,
		executionlayer.Config{
			TokenPriceExportFrequency: *tokenPriceExportFrequency,
			BalanceUpdaterBatchSize:   *balanceUpdaterBatchSize,
			ENSImportBatchSize:        *ensBatchSize,
			Bulk:                      *bulk,
		},
	)
	if *tokenPriceExport {
		go service.SyncTokenPrice(*tokenPriceExportList)
	}

	if end != 0 {
		if err := service.SyncRange(start, end, *skipBlocks, *skipData); err != nil {
			log.Fatal(err, "error indexing from node", 0, map[string]interface{}{"start": *startBlocks, "end": *endBlocks, "concurrency": *concurrency})
		}
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

	service.SyncLive()
}
