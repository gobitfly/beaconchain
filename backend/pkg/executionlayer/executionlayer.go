package executionlayer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"

	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/erc20"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
)

var logger = log.Logger.WithField("service", "el_indexer")

type Config struct {
	TokenPriceExportFrequency time.Duration
	BalanceUpdaterBatchSize   int64
	ENSImportBatchSize        int64
	Bulk                      uint64
}

var defaultConfig = Config{
	TokenPriceExportFrequency: time.Hour,
	BalanceUpdaterBatchSize:   1000,
	ENSImportBatchSize:        200,
	Bulk:                      8000,
}

func (config *Config) validate() {
	if config.TokenPriceExportFrequency == 0 {
		config.TokenPriceExportFrequency = defaultConfig.TokenPriceExportFrequency
	}
	if config.BalanceUpdaterBatchSize == 0 {
		config.BalanceUpdaterBatchSize = defaultConfig.BalanceUpdaterBatchSize
	}
	if config.ENSImportBatchSize == 0 {
		config.ENSImportBatchSize = defaultConfig.ENSImportBatchSize
	}
	if config.Bulk == 0 {
		config.Bulk = defaultConfig.Bulk
	}
}

type IndexerService struct {
	chainID string

	client         *ethclient.Client
	lastBlockStore db2.LastBlocksStore

	balanceUpdater *BalanceUpdater
	tokenPricer    *TokenPricer
	indexer        *Indexer
	reorgWatcher   *ReorgWatcher
	cache          db2.CachedBalanceUpdates
	store          db2.StoreV1
	ensImporter    *ENSImporter

	muBalance sync.Mutex
	muENS     sync.Mutex

	config Config
}

func NewIndexerService(chainID string, client *ethclient.Client, balanceUpdater *BalanceUpdater, reorgWatcher *ReorgWatcher, lastBlockStore db2.LastBlocksStore, tokenPricer *TokenPricer, indexer *Indexer, cache db2.CachedBalanceUpdates, store db2.StoreV1, ensImporter *ENSImporter, config Config) *IndexerService {
	config.validate()
	return &IndexerService{
		chainID:        chainID,
		client:         client,
		lastBlockStore: lastBlockStore,
		balanceUpdater: balanceUpdater,
		tokenPricer:    tokenPricer,
		indexer:        indexer,
		reorgWatcher:   reorgWatcher,
		cache:          cache,
		store:          store,
		ensImporter:    ensImporter,
		muBalance:      sync.Mutex{},
		muENS:          sync.Mutex{},
		config:         config,
	}
}

func (service *IndexerService) SyncRange(start, end uint64, skipNode, skipData bool) error {
	if !skipNode {
		if err := service.indexer.IndexNode(service.chainID, start, end); err != nil {
			return fmt.Errorf("error indexing blocks from node: %v", err)
		}
	}
	if !skipData {
		if err := service.indexer.IndexEvents(service.chainID, start, end); err != nil {
			return fmt.Errorf("error indexing events from node: %v", err)
		}
	}
	if service.balanceUpdater != nil {
		state, err := service.state()
		if err != nil {
			return fmt.Errorf("cannot get state: %v", err)
		}

		service.indexBalances(state)
	}
	return nil
}

func (service *IndexerService) SyncLive() {
	for ; ; time.Sleep(time.Second * 14) {
		state, err := service.state()
		if err != nil {
			logger.WithField("error", err).Error("cannot get state")
			continue
		}
		logger.WithFields(state.Fields()).Info("last blocks")

		if err := service.reorgWatcher.LookForReorg(); err != nil {
			logger.WithField("error", err).Error("reorg lookup")
			continue
		}

		if err := service.indexFromHead(state); err != nil {
			logger.WithField("error", err).Error("indexing from head")
			continue
		}

		go service.indexBalances(state)

		go service.indexENS(state)
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
		time.Sleep(service.config.TokenPriceExportFrequency)
	}
}

func (service *IndexerService) indexFromHead(state syncState) error {
	start := time.Now()
	logger := logger.WithFields(logrus.Fields{
		"blocksRange": fmt.Sprintf("%d-%d", state.blocks+1, state.node),
		"dataRange":   fmt.Sprintf("%d-%d", state.data+1, state.node),
	})
	// clear balance cache
	defer service.cache.Clear(service.chainID)

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
		if err := service.indexer.Index(service.chainID, startBlock, endBlock); err != nil {
			logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("error indexing")
			continue
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

func (service *IndexerService) indexBalances(state syncState) {
	service.muBalance.Lock()
	defer service.muBalance.Unlock()

	for {
		logger := logger.WithFields(state.Fields())
		start := time.Now()

		total, err := service.store.CountBalanceUpdates(service.chainID)
		if err != nil {
			logger.WithField("error", err).Error("error while updating balances")
			continue
		}
		logger = logger.WithField("pending", total)
		if total == 0 {
			logger.Info("finished updating balances")
			return
		}
		balances, err := service.balanceUpdater.UpdateBalances(service.config.BalanceUpdaterBatchSize)
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

func (service *IndexerService) indexENS(state syncState) {
	if service.ensImporter == nil {
		return
	}
	service.muENS.Lock()
	defer service.muENS.Unlock()

	for {
		logger := logger.WithFields(state.Fields())
		start := time.Now()

		total, err := service.store.CountEnsUpdates(service.chainID)
		if err != nil {
			logger.WithField("error", err).Error("error while importing ens")
			continue
		}
		logger = logger.WithField("pending", total)
		if total == 0 {
			logger.Info("finished importing ens")
			return
		}
		res, err := service.ensImporter.Import(service.chainID, service.config.ENSImportBatchSize)
		if err != nil {
			logger.WithField("error", err).Error("error while importing balances")
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

func (service *IndexerService) state() (syncState, error) {
	lastBlock, err := service.client.BlockNumber(context.Background())
	if err != nil {
		return syncState{}, fmt.Errorf("get chain head: %w", err)
	}
	lastBlockFromBlocksTable, err := service.lastBlockStore.GetInBlocksTable(service.chainID)
	if err != nil {
		return syncState{}, fmt.Errorf("get last block from blocks table: %w", err)
	}
	lastBlockFromDataTable, err := service.lastBlockStore.GetInDataTable(service.chainID)
	if err != nil {
		return syncState{}, fmt.Errorf("get last block from data table: %w", err)
	}
	return syncState{
		node:   lastBlock,
		blocks: lastBlockFromBlocksTable,
		data:   lastBlockFromDataTable,
	}, nil
}

type syncState struct {
	node   uint64
	blocks uint64
	data   uint64
}

func (state syncState) Fields() map[string]interface{} {
	return map[string]interface{}{
		"node":   state.node,
		"blocks": state.blocks,
		"data":   state.data,
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
