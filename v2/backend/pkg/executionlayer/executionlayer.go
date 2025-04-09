package executionlayer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/db2/database"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/rpc"
	"github.com/gobitfly/beaconchain/pkg/commons/services"
	"github.com/gobitfly/beaconchain/pkg/executionlayer/evm"
)

var logger = log.Logger.WithField("service", "el_indexer")

type ServiceConfig struct {
	TokenPriceFrequency time.Duration
	BlockFrequency      time.Duration
}

type IndexerService struct {
	tokenPricer  *TokenPricer
	indexer      *Indexer
	stateReader  StateReader
	reorgWatcher *ReorgWatcher

	config ServiceConfig
}

func NewIndexerService(config Config, store db2.StoreV1, client *rpc.ErigonClient, localCache database.RemoteCache, remoteCache database.RemoteCache) *IndexerService {
	lastBlockStore := db2.NewCachedLastBlocks(remoteCache, store)

	blockIndexer := NewBlockIndexer(store, lastBlockStore, config.BlockIndexer, client, AllTransformers...)
	batcher := evm.NewBatcher(client.GetChainID(), client.GetNativeClient(), config.Batcher)
	balanceUpdater := NewBalanceUpdater(store, store, batcher)
	ensImporter := NewENSImporter(store, db2.NewENSStore(db.WriterDb), NewEnsContracts(client.GetNativeClient()))

	indexer := NewIndexer(
		db2.CachedBalanceUpdates{RemoteCache: localCache},
		blockIndexer,
		&balanceUpdater,
		store,
		ensImporter,
		config.Indexer,
	)

	tokenPricer := NewTokenPricer(
		store,
		client.GetChainID().String(),
		NewLlamaClient(),
		batcher,
	)

	stateReader := NewStateReader(client.GetChainID().String(), client.GetNativeClient(), lastBlockStore)
	reorgWatcher := NewReorgWatcher(client.GetNativeClient(), store, config.Reorg, client.GetChainID().String(), lastBlockStore)

	return &IndexerService{
		tokenPricer:  tokenPricer,
		indexer:      indexer,
		stateReader:  stateReader,
		reorgWatcher: reorgWatcher,
		config:       config.Service,
	}
}

func NewIndexerServiceWithComponents(stateReader StateReader, indexer *Indexer, reorgWatcher *ReorgWatcher, tokenPricer *TokenPricer, config ServiceConfig) *IndexerService {
	return &IndexerService{
		tokenPricer:  tokenPricer,
		indexer:      indexer,
		stateReader:  stateReader,
		reorgWatcher: reorgWatcher,
		config:       config,
	}
}

func (service *IndexerService) SyncRange(start, end uint64, skipNode, skipData bool) error {
	state, err := service.stateReader.state()
	if err != nil {
		return fmt.Errorf("cannot not get state: %w", err)
	}
	logger.WithFields(state.Fields()).Info("last blocks")
	return service.indexer.Range(state, start, end, skipNode, skipData)
}

func (service *IndexerService) SyncLive() {
	for ; ; time.Sleep(service.config.BlockFrequency) {
		reorgDepth, err := service.reorgWatcher.LookForReorg()
		if err != nil {
			logger.WithField("error", err).Error("reorg lookup")
			break
		}
		if reorgDepth != 0 {
			indexingMetrics.ReorgBlockTotal(service.stateReader.chainID, reorgDepth)
			logger.WithField("depth", reorgDepth).Info("reorg detected")
		}

		for {
			state, err := service.stateReader.state()
			if err != nil {
				logger.WithField("error", err).Error("cannot get state")
				break
			}
			if state.node == state.LastProcessed() {
				// we caught up with blockchain head, we can stop indexing
				break
			}
			indexingMetrics.BlockDifference(state.chainID, state.node-state.LastProcessed())
			startBlock := state.NextBlock()
			bulk := min(service.indexer.config.Bulk, state.node-startBlock+1)
			endBlock := min(startBlock+bulk-1, state.node)
			start := time.Now()

			logger := logger.WithFields(state.Fields()).WithFields(logrus.Fields{
				"startBlock": startBlock,
				"endBlock":   endBlock,
			})

			if err := service.indexer.FromHead(state, startBlock, endBlock); err != nil {
				logger.WithField("error", err).Error("indexing from head")
				break
			}
			blockIndexingTime := time.Since(start)

			var balanceTime, ensTime time.Duration
			g := errgroup.Group{}
			g.Go(func() error {
				start := time.Now()
				service.indexer.Balances(state)
				balanceTime = time.Since(start)
				return nil
			})
			g.Go(func() error {
				start := time.Now()
				service.indexer.ENS(state)
				ensTime = time.Since(start)
				return nil
			})
			_ = g.Wait()
			duration := time.Since(start) // save duration to unify log and metric value
			logger.WithFields(logrus.Fields{
				"indexingTime":      duration,
				"blockIndexingTime": blockIndexingTime,
				"balanceTime":       balanceTime,
				"ensTime":           ensTime,
			}).Info("indexing done")
			indexingMetrics.IndexingTime(state.chainID, duration)
		}
		// TODO: remove that, it seems weird to write to a database that the service is running
		// we have logs, metrics and other indicator for that purpose
		services.ReportStatus("eth1indexer", "Running", nil)
	}
}

func (service *IndexerService) SyncTokenPrice(path string) {
	for {
		start := time.Now()
		tokenList, err := readTokenListFile(path)
		if err != nil {
			logger.WithError(err).Error("error reading token list file")
			continue
		}
		if err := service.tokenPricer.UpdateTokens(tokenList); err != nil {
			logger.WithError(err).Error("error updating token prices")
			continue
		}
		logger.WithFields(logrus.Fields{
			"duration": time.Since(start),
			"tokens":   tokenList.Names(),
		}).Info("token prices updated")
		time.Sleep(service.config.TokenPriceFrequency)
	}
}

type StateReader struct {
	chainID        string
	client         ethereum.BlockNumberReader
	lastBlockStore db2.LastBlocksStore
}

func NewStateReader(chainID string, client *ethclient.Client, lastBlockStore db2.LastBlocksStore) StateReader {
	return StateReader{
		chainID:        chainID,
		client:         client,
		lastBlockStore: lastBlockStore,
	}
}

func (r StateReader) state() (syncState, error) {
	var genesis bool
	lastBlock, err := r.client.BlockNumber(context.Background())
	if err != nil {
		return syncState{}, fmt.Errorf("get chain head: %w", err)
	}
	lastBlockFromBlocksTable, err := r.lastBlockStore.GetInBlocksTable(r.chainID)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return syncState{}, fmt.Errorf("get last block from blocks table: %w", err)
		}
		genesis = true
	}
	lastBlockFromDataTable, err := r.lastBlockStore.GetInDataTable(r.chainID)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return syncState{}, fmt.Errorf("get last block from data table: %w", err)
		}
		genesis = true
	}
	return syncState{
		chainID: r.chainID,
		node:    lastBlock,
		blocks:  lastBlockFromBlocksTable,
		data:    lastBlockFromDataTable,
		genesis: genesis,
	}, nil
}

type syncState struct {
	chainID string
	node    uint64
	blocks  uint64
	data    uint64

	genesis bool
}

// LastProcessed returns the real last block processed by taking the smallest block between state.data and state.blocks
func (state syncState) LastProcessed() uint64 {
	return min(state.data, state.blocks)
}

func (state syncState) NextBlock() uint64 {
	lastProcessed := state.LastProcessed()
	if state.genesis {
		return lastProcessed
	}
	return lastProcessed + 1
}

func (state syncState) Fields() map[string]interface{} {
	fields := map[string]interface{}{
		"chainID": state.chainID,
		"node":    state.node,
		"blocks":  state.blocks,
		"data":    state.data,
	}
	if state.genesis {
		fields["genesis"] = true
	}
	return fields
}

type IndexerConfig struct {
	BalanceUpdaterBatchSize int64
	EnableENS               bool
	ENSImportBatchSize      int64
	Bulk                    uint64
}

type Indexer struct {
	balanceCache   db2.CachedBalanceUpdates
	indexer        *BlockIndexer
	balanceUpdater *BalanceUpdater
	store          db2.StoreV1
	ensImporter    *ENSImporter

	config IndexerConfig

	muBalance sync.Mutex
	muENS     sync.Mutex
}

func NewIndexer(cache db2.CachedBalanceUpdates, blockIndexer *BlockIndexer, balanceUpdater *BalanceUpdater, store db2.StoreV1, ensImporter *ENSImporter, config IndexerConfig) *Indexer {
	if !config.EnableENS {
		ensImporter = nil
	}
	return &Indexer{
		balanceCache:   cache,
		indexer:        blockIndexer,
		balanceUpdater: balanceUpdater,
		store:          store,
		ensImporter:    ensImporter,
		config:         config,
	}
}

func (indexer *Indexer) FromHead(state syncState, startBlock, endBlock uint64) error {
	// clear balance cache
	defer indexer.balanceCache.Clear(state.chainID)
	return indexer.indexer.Index(state.chainID, startBlock, endBlock)
}

func (indexer *Indexer) Balances(state syncState) {
	indexer.muBalance.Lock()
	defer indexer.muBalance.Unlock()

	globalStart := time.Now()
	for {
		logger := logger.WithFields(state.Fields())
		start := time.Now()

		total, err := indexer.store.CountBalanceUpdates(state.chainID)
		if err != nil {
			logger.WithField("error", err).Error("error while counting balance updates")
			continue
		}
		logger = logger.WithField("pending", total)
		indexingMetrics.PendingBalanceUpdate(state.chainID, total)
		if total == 0 {
			logger.WithField("duration", time.Since(globalStart)).Info("balances updated")
			return
		}
		balances, err := indexer.balanceUpdater.UpdateBalances(state.chainID, indexer.config.BalanceUpdaterBatchSize)
		if err != nil {
			logger.WithField("error", err).Error("error while updating balances")
			return
		}
		logger.WithFields(logrus.Fields{
			"processed":   len(balances),
			"lastAddress": balances[len(balances)-1].Address.String(),
			"duration":    time.Since(start),
		}).Info("update balances")
	}
}

func (indexer *Indexer) ENS(state syncState) {
	if indexer.ensImporter == nil {
		return
	}
	indexer.muENS.Lock()
	defer indexer.muENS.Unlock()

	for {
		logger := logger.WithFields(state.Fields())
		start := time.Now()

		total, err := indexer.store.CountEnsUpdates(state.chainID)
		if err != nil {
			logger.WithField("error", err).Error("error while counting total ens updates")
			continue
		}
		logger = logger.WithField("pending", total)
		indexingMetrics.PendingENSUpdate(state.chainID, total)
		if total == 0 {
			logger.Info("ens imported")
			return
		}
		res, err := indexer.ensImporter.Import(state.chainID, indexer.config.ENSImportBatchSize)
		if err != nil {
			logger.WithField("error", err).Error("error while importing ens")
			continue
		}
		logger.WithFields(logrus.Fields{
			"processed": len(res.updated) + len(res.deleted),
			"updated":   res.updated,
			"deleted":   res.deleted,
			"duration":  time.Since(start),
		}).Info("import ens")
	}
}

func (indexer *Indexer) Range(state syncState, start, end uint64, skipNode, skipData bool) error {
	if !skipNode {
		if err := indexer.indexer.IndexNode(state.chainID, start, end); err != nil {
			return fmt.Errorf("error indexing blocks from node: %v", err)
		}
	}
	if !skipData {
		if err := indexer.indexer.IndexEvents(state.chainID, start, end); err != nil {
			return fmt.Errorf("error indexing events from node: %v", err)
		}
	}
	if indexer.balanceUpdater != nil {
		indexer.Balances(state)
	}
	return nil
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
