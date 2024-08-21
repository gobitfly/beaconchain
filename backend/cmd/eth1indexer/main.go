package eth1indexer

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
	"github.com/gobitfly/beaconchain/pkg/commons/version"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"

	//nolint:gosec
	_ "net/http/pprof"
)

func Run() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	erigonEndpoint := fs.String("erigon", "", "Erigon archive node enpoint")
	block := fs.Int64("block", 0, "Index a specific block")

	reorgDepth := fs.Int("reorg.depth", 20, "Lookback to check and handle chain reorgs")

	concurrencyBlocks := fs.Int64("blocks.concurrency", 30, "Concurrency to use when indexing blocks from erigon")
	startBlocks := fs.Int64("blocks.start", 0, "Block to start indexing")
	endBlocks := fs.Int64("blocks.end", 0, "Block to finish indexing")
	bulkBlocks := fs.Int64("blocks.bulk", 8000, "Maximum number of blocks to be processed before saving")
	offsetBlocks := fs.Int64("blocks.offset", 100, "Blocks offset")
	checkBlocksGaps := fs.Bool("blocks.gaps", false, "Check for gaps in the blocks table")
	checkBlocksGapsLookback := fs.Int("blocks.gaps.lookback", 1000000, "Lookback for gaps check of the blocks table")
	traceMode := fs.String("blocks.tracemode", "parity/geth", "Trace mode to use, can bei either 'parity', 'geth' or 'parity/geth' for both")

	concurrencyData := fs.Int64("data.concurrency", 30, "Concurrency to use when indexing data from bigtable")
	startData := fs.Int64("data.start", 0, "Block to start indexing")
	endData := fs.Int64("data.end", 0, "Block to finish indexing")
	bulkData := fs.Int64("data.bulk", 8000, "Maximum number of blocks to be processed before saving")
	offsetData := fs.Int64("data.offset", 1000, "Data offset")
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

	fs.Parse(os.Args[2:])

	log.Info(*configPath)
	if *versionFlag {
		log.Infof(version.Version)
		log.Infof(version.GoVersion)
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
		go func(addr string) {
			log.Infof("serving metrics on %v", addr)
			if err := metrics.Serve(addr); err != nil {
				log.Fatal(err, "error serving metrics", 0)
			}
		}(utils.Config.Metrics.Address)
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

	db.WriterDb, db.ReaderDb = db.MustInitDB(&types.DatabaseConfig{
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
	defer db.WriterDb.Close()

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

	balanceUpdaterPrefix := chainId + ":B:"

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

	if *tokenPriceExport {
		go func() {
			for {
				err = UpdateTokenPrices(bt, client, *tokenPriceExportList)
				if err != nil {
					log.Error(err, "error while updating token prices", 0)
					time.Sleep(*tokenPriceExportFrequency)
				}
				time.Sleep(*tokenPriceExportFrequency)
			}
		}()
	}

	if *enableFullBalanceUpdater {
		ProcessMetadataUpdates(bt, client, balanceUpdaterPrefix, *balanceUpdaterBatchSize, -1)
		return
	}

	transforms := make([]func(blk *types.Eth1Block, cache *freecache.Cache) (*types.BulkMutations, *types.BulkMutations, error), 0)
	transforms = append(transforms,
		bt.TransformBlock,
		bt.TransformTx,
		bt.TransformItx,
		bt.TransformBlobTx,
		bt.TransformERC20,
		bt.TransformERC721,
		bt.TransformERC1155,
		bt.TransformUncle,
		bt.TransformWithdrawals,
		bt.TransformEnsNameRegistered,
		bt.TransformContract)

	cache := freecache.NewCache(100 * 1024 * 1024) // 100 MB limit

	if *block != 0 {
		err = IndexFromNode(bt, client, *block, *block, *concurrencyBlocks, *traceMode)
		if err != nil {
			log.Fatal(err, "error indexing from node", 0, map[string]interface{}{"block": *block, "concurrency": *concurrencyBlocks})
		}
		err = bt.IndexEventsWithTransformers(*block, *block, transforms, *concurrencyData, cache)
		if err != nil {
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
		err = IndexFromNode(bt, client, *startBlocks, *endBlocks, *concurrencyBlocks, *traceMode)
		if err != nil {
			log.Fatal(err, "error indexing from node", 0, map[string]interface{}{"start": *startBlocks, "end": *endBlocks, "concurrency": *concurrencyBlocks})
		}
		return
	}

	if *endData != 0 && *startData < *endData {
		err = bt.IndexEventsWithTransformers(*startData, *endData, transforms, *concurrencyData, cache)
		if err != nil {
			log.Fatal(err, "error indexing from bigtable", 0)
		}
		cache.Clear()
		return
	}

	lastSuccessulBlockIndexingTs := time.Now()
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

		if *enableEnsUpdater {
			err := bt.ImportEnsUpdates(client.GetNativeClient(), *ensBatchSize)
			if err != nil {
				log.Error(err, "error importing ens updates", 0, nil)
				continue
			}
		}

		log.Infof("index run completed")
		services.ReportStatus("eth1indexer", "Running", nil)
	}

	// utils.WaitForCtrlC()
}

func UpdateTokenPrices(bt *db.Bigtable, client *rpc.ErigonClient, tokenListPath string) error {
	tokenListContent, err := os.ReadFile(tokenListPath)
	if err != nil {
		return err
	}

	tokenList := &erc20.ERC20TokenList{}

	err = json.Unmarshal(tokenListContent, tokenList)
	if err != nil {
		return err
	}

	type defillamaPriceRequest struct {
		Coins []string `json:"coins"`
	}
	coinsList := make([]string, 0, len(tokenList.Tokens))
	for _, token := range tokenList.Tokens {
		coinsList = append(coinsList, "ethereum:"+token.Address)
	}

	req := &defillamaPriceRequest{
		Coins: coinsList,
	}

	reqEncoded, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpClient := &http.Client{Timeout: time.Second * 10}

	resp, err := httpClient.Post("https://coins.llama.fi/prices", "application/json", bytes.NewReader(reqEncoded))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error querying defillama api: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	type defillamaCoin struct {
		Decimals  int64            `json:"decimals"`
		Price     *decimal.Decimal `json:"price"`
		Symbol    string           `json:"symbol"`
		Timestamp int64            `json:"timestamp"`
	}

	type defillamaResponse struct {
		Coins map[string]defillamaCoin `json:"coins"`
	}

	respParsed := &defillamaResponse{}
	err = json.Unmarshal(body, respParsed)
	if err != nil {
		return err
	}

	tokenPrices := make([]*types.ERC20TokenPrice, 0, len(respParsed.Coins))
	for address, data := range respParsed.Coins {
		tokenPrices = append(tokenPrices, &types.ERC20TokenPrice{
			Token: common.FromHex(strings.TrimPrefix(address, "ethereum:0x")),
			Price: []byte(data.Price.String()),
		})
	}

	g := new(errgroup.Group)
	g.SetLimit(20)
	for i := range tokenPrices {
		i := i
		g.Go(func() error {
			metadata, err := client.GetERC20TokenMetadata(tokenPrices[i].Token)
			if err != nil {
				return err
			}
			tokenPrices[i].TotalSupply = metadata.TotalSupply
			// log.LogInfo("price for token %x is %s @ %v", tokenPrices[i].Token, tokenPrices[i].Price, new(big.Int).SetBytes(tokenPrices[i].TotalSupply))
			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		return err
	}

	return bt.SaveERC20TokenPrices(tokenPrices)
}

func HandleChainReorgs(bt *db.Bigtable, client *rpc.ErigonClient, depth int) error {
	ctx := context.Background()
	// get latest block from the node
	latestNodeBlock, err := client.GetNativeClient().BlockByNumber(ctx, nil)
	if err != nil {
		return err
	}
	latestNodeBlockNumber := latestNodeBlock.NumberU64()

	// for each block check if block node hash and block db hash match
	if depth > int(latestNodeBlockNumber) {
		depth = int(latestNodeBlockNumber)
	}
	for i := latestNodeBlockNumber - uint64(depth); i <= latestNodeBlockNumber; i++ {
		nodeBlock, err := client.GetNativeClient().HeaderByNumber(ctx, big.NewInt(int64(i)))
		if err != nil {
			return err
		}

		dbBlock, err := bt.GetBlockFromBlocksTable(i)
		if err != nil {
			if err == db.ErrBlockNotFound { // exit if we hit a block that is not yet in the db
				return nil
			}
			return err
		}

		if !bytes.Equal(nodeBlock.Hash().Bytes(), dbBlock.Hash) {
			log.Warnf("found incosistency at height %v, node block hash: %x, db block hash: %x", i, nodeBlock.Hash().Bytes(), dbBlock.Hash)

			// first we set the cached marker of the last block in the blocks/data table to the block prior to the forked one
			if i > 0 {
				previousBlock := i - 1
				err := bt.SetLastBlockInBlocksTable(int64(previousBlock))
				if err != nil {
					return fmt.Errorf("error setting last block [%v] in blocks table: %w", previousBlock, err)
				}
				err = bt.SetLastBlockInDataTable(int64(previousBlock))
				if err != nil {
					return fmt.Errorf("error setting last block [%v] in data table: %w", previousBlock, err)
				}
				// now we can proceed to delete all blocks including and after the forked block
			}
			// delete all blocks starting from the fork block up to the latest block in the db
			for j := i; j <= latestNodeBlockNumber; j++ {
				dbBlock, err := bt.GetBlockFromBlocksTable(j)
				if err != nil {
					if err == db.ErrBlockNotFound { // exit if we hit a block that is not yet in the db
						return nil
					}
					return err
				}
				log.Infof("deleting block at height %v with hash %x", dbBlock.Number, dbBlock.Hash)

				err = bt.DeleteBlock(dbBlock.Number, dbBlock.Hash)
				if err != nil {
					return err
				}
			}
		} else {
			log.Infof("height %v, node block hash: %x, db block hash: %x", i, nodeBlock.Hash().Bytes(), dbBlock.Hash)
		}
	}

	return nil
}

func ProcessMetadataUpdates(bt *db.Bigtable, client *rpc.ErigonClient, prefix string, batchSize int, iterations int) {
	lastKey := prefix

	its := 0
	for {
		start := time.Now()
		keys, pairs, err := bt.GetMetadataUpdates(prefix, lastKey, batchSize)
		if err != nil {
			log.Error(err, "error retrieving metadata updates from bigtable", 0)
			return
		}

		if len(keys) == 0 {
			return
		}

		balances := make([]*types.Eth1AddressBalance, 0, len(pairs))
		for b := 0; b < len(pairs); b += batchSize {
			start := b
			end := b + batchSize
			if len(pairs) < end {
				end = len(pairs)
			}

			log.Infof("processing batch %v with start %v and end %v", b, start, end)

			b, err := client.GetBalances(pairs[start:end], 2, 4)

			if err != nil {
				log.Error(err, "error retrieving balances from node", 0)
				return
			}
			balances = append(balances, b...)
		}

		err = bt.SaveBalances(balances, keys)
		if err != nil {
			log.Error(err, "error saving balances to bigtable", 0)
			return
		}

		lastKey = keys[len(keys)-1]
		log.Infof("retrieved %v balances in %v, currently at %v", len(balances), time.Since(start), lastKey)

		its++

		if iterations != -1 && its > iterations {
			return
		}
	}
}

func IndexFromNode(bt *db.Bigtable, client *rpc.ErigonClient, start, end, concurrency int64, traceMode string) error {
	ctx := context.Background()
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(int(concurrency))

	startTs := time.Now()
	lastTickTs := time.Now()

	processedBlocks := int64(0)

	for i := start; i <= end; i++ {
		i := i
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}

			blockStartTs := time.Now()
			bc, timings, err := client.GetBlock(i, traceMode)
			if err != nil {
				return fmt.Errorf("error getting block: %v from ethereum node err: %w", i, err)
			}

			dbStart := time.Now()
			err = bt.SaveBlock(bc)
			if err != nil {
				return fmt.Errorf("error saving block: %v to bigtable: %w", i, err)
			}
			current := atomic.AddInt64(&processedBlocks, 1)
			if current%100 == 0 {
				r := end - start
				if r == 0 {
					r = 1
				}
				perc := float64(i-start) * 100 / float64(r)

				log.Infof("retrieved & saved block %v (0x%x) in %v (header: %v, receipts: %v, traces: %v, db: %v)", bc.Number, bc.Hash, time.Since(blockStartTs), timings.Headers, timings.Receipts, timings.Traces, time.Since(dbStart))
				log.Infof("processed %v blocks in %v (%.1f blocks / sec); sync is %.1f%% complete", current, time.Since(startTs), float64((current))/time.Since(lastTickTs).Seconds(), perc)

				lastTickTs = time.Now()
				atomic.StoreInt64(&processedBlocks, 0)
			}
			return nil
		})
	}

	err := g.Wait()

	if err != nil {
		return err
	}

	lastBlockInCache, err := bt.GetLastBlockInBlocksTable()
	if err != nil {
		return err
	}

	if end > int64(lastBlockInCache) {
		err := bt.SetLastBlockInBlocksTable(end)

		if err != nil {
			return err
		}
	}
	return nil
}

func ImportMainnetERC20TokenMetadataFromTokenDirectory(bt *db.Bigtable) {
	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Get("<INSERT_TOKENLIST_URL>")

	if err != nil {
		log.Fatal(err, "getting client error", 0)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err, "reading body for ERC20 tokens error", 0)
	}

	type TokenDirectory struct {
		ChainID       int64    `json:"chainId"`
		Keywords      []string `json:"keywords"`
		LogoURI       string   `json:"logoURI"`
		Name          string   `json:"name"`
		Timestamp     string   `json:"timestamp"`
		TokenStandard string   `json:"tokenStandard"`
		Tokens        []struct {
			Address    string `json:"address"`
			ChainID    int64  `json:"chainId"`
			Decimals   int64  `json:"decimals"`
			Extensions struct {
				Description   string      `json:"description"`
				Link          string      `json:"link"`
				OgImage       interface{} `json:"ogImage"`
				OriginAddress string      `json:"originAddress"`
				OriginChainID int64       `json:"originChainId"`
			} `json:"extensions"`
			LogoURI string `json:"logoURI"`
			Name    string `json:"name"`
			Symbol  string `json:"symbol"`
		} `json:"tokens"`
	}

	td := &TokenDirectory{}

	err = json.Unmarshal(body, td)

	if err != nil {
		log.Fatal(err, "unmarshal json body error", 0)
	}

	for _, token := range td.Tokens {
		address, err := hex.DecodeString(strings.TrimPrefix(token.Address, "0x"))
		if err != nil {
			log.Fatal(err, "decoding string to hex error", 0)
		}
		log.Infof("processing token %v at address %x", token.Name, address)

		meta := &types.ERC20Metadata{}
		meta.Decimals = big.NewInt(token.Decimals).Bytes()
		meta.Description = token.Extensions.Description
		if len(token.LogoURI) > 0 {
			resp, err := client.Get(token.LogoURI)

			if err == nil && resp.StatusCode == 200 {
				body, err := io.ReadAll(resp.Body)

				if err != nil {
					log.Fatal(err, "reading body for ERC20 token logo URI error", 0)
				}

				resp.Body.Close()

				meta.Logo = body
				meta.LogoFormat = token.LogoURI
			}
		}
		meta.Name = token.Name
		meta.OfficialSite = token.Extensions.Link
		meta.Symbol = token.Symbol

		err = bt.SaveERC20Metadata(address, meta)
		if err != nil {
			log.Fatal(err, "error while saving ERC20 metadata", 0)
		}
		time.Sleep(time.Millisecond * 250)
	}
}
