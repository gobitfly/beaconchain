package executionlayer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"

	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

var logger = log.Logger.WithField("service", "el_indexer")

type Config struct {
	TokenPriceFrequency time.Duration
	BlockFrequency      time.Duration
}

var defaultConfig = Config{
	TokenPriceFrequency: time.Hour,
	BlockFrequency:      14 * time.Second,
}

// init set default value for unset fields
func (config *Config) init() {
	if config.TokenPriceFrequency == 0 {
		config.TokenPriceFrequency = defaultConfig.TokenPriceFrequency
	}
	if config.BlockFrequency == 0 {
		config.BlockFrequency = defaultConfig.BlockFrequency
	}
}

type IndexerService struct {
	tokenPricer  *TokenPricer
	indexer      *Indexer
	stateReader  StateReader
	reorgWatcher *ReorgWatcher

	config Config
}

func NewIndexerService(stateReader StateReader, indexer *Indexer, reorgWatcher *ReorgWatcher, tokenPricer *TokenPricer, config Config) *IndexerService {
	config.init()
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
		state, err := service.stateReader.state()
		if err != nil {
			logger.WithField("error", err).Error("cannot get state")
			continue
		}
		logger.WithFields(state.Fields()).Info("last blocks")

		if err := service.reorgWatcher.LookForReorg(); err != nil {
			logger.WithField("error", err).Error("reorg lookup")
			continue
		}

		if err := service.indexer.FromHead(state); err != nil {
			logger.WithField("error", err).Error("indexing from head")
			continue
		}

		go service.indexer.Balances(state)

		go service.indexer.ENS(state)
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
	lastBlock, err := r.client.BlockNumber(context.Background())
	if err != nil {
		return syncState{}, fmt.Errorf("get chain head: %w", err)
	}
	lastBlockFromBlocksTable, err := r.lastBlockStore.GetInBlocksTable(r.chainID)
	if err != nil {
		return syncState{}, fmt.Errorf("get last block from blocks table: %w", err)
	}
	lastBlockFromDataTable, err := r.lastBlockStore.GetInDataTable(r.chainID)
	if err != nil {
		return syncState{}, fmt.Errorf("get last block from data table: %w", err)
	}
	return syncState{
		chainID: r.chainID,
		node:    lastBlock,
		blocks:  lastBlockFromBlocksTable,
		data:    lastBlockFromDataTable,
	}, nil
}

type syncState struct {
	chainID string
	node    uint64
	blocks  uint64
	data    uint64
}

func (state syncState) Fields() map[string]interface{} {
	return map[string]interface{}{
		"chainID": state.chainID,
		"node":    state.node,
		"blocks":  state.blocks,
		"data":    state.data,
	}
}

type IndexerConfig struct {
	BalanceUpdaterBatchSize int64
	ENSImportBatchSize      int64
	Bulk                    uint64
}

var defaultIndexerConfig = IndexerConfig{
	BalanceUpdaterBatchSize: 1000,
	ENSImportBatchSize:      200,
	Bulk:                    8000,
}

// init set default value for unset fields
func (c *IndexerConfig) init() {
	if c.BalanceUpdaterBatchSize == 0 {
		c.BalanceUpdaterBatchSize = defaultIndexerConfig.BalanceUpdaterBatchSize
	}
	if c.ENSImportBatchSize == 0 {
		c.ENSImportBatchSize = defaultIndexerConfig.ENSImportBatchSize
	}
	if c.Bulk == 0 {
		c.Bulk = defaultIndexerConfig.Bulk
	}
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
	config.init()
	return &Indexer{
		balanceCache:   cache,
		indexer:        blockIndexer,
		balanceUpdater: balanceUpdater,
		store:          store,
		ensImporter:    ensImporter,
		config:         config,
	}
}

func (service *Indexer) FromHead(state syncState) error {
	start := time.Now()
	logger := logger.WithFields(logrus.Fields{
		"blocksRange": fmt.Sprintf("%d-%d", state.blocks+1, state.node),
		"dataRange":   fmt.Sprintf("%d-%d", state.data+1, state.node),
	})
	// clear balance cache
	defer service.balanceCache.Clear(state.chainID)

	// get the real last block processed by taking the smallest block between state.data and state.blocks
	startBlock := max(min(state.data, state.blocks)+1, 0)
	bulk := min(service.config.Bulk, state.node-startBlock+1)

	for ; startBlock <= state.node; startBlock += bulk {
		start := time.Now()
		endBlock := min(startBlock+bulk-1, state.node)
		logger = logger.WithFields(logrus.Fields{
			"start": startBlock,
			"end":   endBlock,
		})
		if err := service.indexer.Index(state.chainID, startBlock, endBlock); err != nil {
			logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("error indexing")
			return err
		}
		logger.WithFields(logrus.Fields{
			"elapsed": time.Since(start),
		}).Info("indexed blocks")
	}

	logger.WithFields(logrus.Fields{
		"duration": time.Since(start),
	}).Info("indexed head")
	return nil
}

func (service *Indexer) Balances(state syncState) {
	service.muBalance.Lock()
	defer service.muBalance.Unlock()

	for {
		logger := logger.WithFields(state.Fields())
		start := time.Now()

		total, err := service.store.CountBalanceUpdates(state.chainID)
		if err != nil {
			logger.WithField("error", err).Error("error while counting balance updates")
			continue
		}
		logger = logger.WithField("pending", total)
		if total == 0 {
			logger.Info("finished updating balances")
			return
		}
		balances, err := service.balanceUpdater.UpdateBalances(state.chainID, service.config.BalanceUpdaterBatchSize)
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

func (service *Indexer) ENS(state syncState) {
	if service.ensImporter == nil {
		return
	}
	service.muENS.Lock()
	defer service.muENS.Unlock()

	for {
		logger := logger.WithFields(state.Fields())
		start := time.Now()

		total, err := service.store.CountEnsUpdates(state.chainID)
		if err != nil {
			logger.WithField("error", err).Error("error while counting total ens updates")
			continue
		}
		logger = logger.WithField("pending", total)
		if total == 0 {
			logger.Info("finished importing ens")
			return
		}
		res, err := service.ensImporter.Import(state.chainID, service.config.ENSImportBatchSize)
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

func (service *Indexer) Range(state syncState, start, end uint64, skipNode, skipData bool) error {
	if !skipNode {
		if err := service.indexer.IndexNode(state.chainID, start, end); err != nil {
			return fmt.Errorf("error indexing blocks from node: %v", err)
		}
	}
	if !skipData {
		if err := service.indexer.IndexEvents(state.chainID, start, end); err != nil {
			return fmt.Errorf("error indexing events from node: %v", err)
		}
	}
	if service.balanceUpdater != nil {
		service.Balances(state)
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
