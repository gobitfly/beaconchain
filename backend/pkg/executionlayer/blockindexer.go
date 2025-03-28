package executionlayer

import (
	"context"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/gobitfly/beaconchain/pkg/commons/db2"
	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/gobitfly/beaconchain/pkg/commons/types"
)

type Client interface {
	GetBlock(number uint64, traceMode string) (*types.Eth1Block, *types.GetBlockTimings, error)
	GetChainID() *big.Int
}

type clientWithMetrics struct {
	client Client
}

func (c clientWithMetrics) GetBlock(number uint64, traceMode string) (*types.Eth1Block, *types.GetBlockTimings, error) {
	start := time.Now()
	defer func(start time.Time) {
		metrics.ClientGetBlock.WithLabelValues(c.client.GetChainID().String()).Observe(time.Since(start).Seconds())
	}(start)
	return c.client.GetBlock(number, traceMode)
}

func (c clientWithMetrics) GetChainID() *big.Int {
	return c.client.GetChainID()
}

type Store interface {
	AddIndexedBlock(block db2.IndexedBlock) error
	GetBlocksRange(chainID string, start, end uint64) ([]*types.Eth1Block, error)
	SaveBlock(chainID string, block *types.Eth1Block) error
}

type BlockIndexerConfig struct {
	Concurrency uint64
	TraceMode   string
}

type BlockIndexer struct {
	store          Store
	lastBlockStore db2.LastBlocksStore
	transformers   []Transformer
	client         Client
	config         BlockIndexerConfig
}

func NewBlockIndexer(store Store, lastBlockStore db2.LastBlocksStore, config BlockIndexerConfig, client Client, transformers ...Transformer) *BlockIndexer {
	return &BlockIndexer{
		store:          store,
		lastBlockStore: lastBlockStore,
		transformers:   transformers,
		client:         clientWithMetrics{client},
		config:         config,
	}
}

// IndexNode retrieve types.Eth1Block from the client and save them into the store
func (indexer *BlockIndexer) IndexNode(chainID string, start, end uint64) error {
	g, gCtx := errgroup.WithContext(context.Background())
	g.SetLimit(int(indexer.config.Concurrency))

	startTs := time.Now()
	lastTickTs := time.Now()

	reporter := new(reporter)
	for i := start; i <= end; i++ {
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}

			blockStartTs := time.Now()
			block, timings, err := indexer.client.GetBlock(i, indexer.config.TraceMode)
			if err != nil {
				return fmt.Errorf("error getting block: %v from ethereum node err: %w", i, err)
			}

			dbStart := time.Now()
			err = indexer.store.SaveBlock(chainID, block)
			if err != nil {
				return fmt.Errorf("error saving block: %v to bigtable: %w", i, err)
			}

			reporter.report(end-start, block, timings, &lastTickTs, startTs, blockStartTs, dbStart)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	lastBlockInCache, err := indexer.lastBlockStore.GetInBlocksTable(chainID)
	if err != nil {
		return err
	}

	if end > lastBlockInCache {
		if err := indexer.lastBlockStore.SetInBlocksTable(chainID, end); err != nil {
			return err
		}
	}
	return nil
}

// IndexEvents retrieve read the types.Eth1Block from the store and apply the transformers to them
func (indexer *BlockIndexer) IndexEvents(chainID string, start, end uint64) error {
	retrieval := new(errgroup.Group)
	retrieval.SetLimit(int(indexer.config.Concurrency))

	indexing := new(errgroup.Group)
	indexing.SetLimit(int(indexer.config.Concurrency * indexer.config.Concurrency))

	batchSize := uint64(1000)
	for i := start; i <= end; i += batchSize {
		firstBlock := i
		lastBlock := firstBlock + batchSize - 1
		if lastBlock > end {
			lastBlock = end
		}

		retrieval.Go(func() error {
			high := lastBlock
			low := firstBlock

			blocks, err := indexer.store.GetBlocksRange(chainID, low, high)
			if err != nil {
				log.Error(err, "error getting blocks descending", 0, map[string]interface{}{"high": high, "low": low})
				return err
			}

			for _, block := range blocks {
				indexing.Go(func() error {
					return indexer.indexBlock(chainID, block)
				})
			}
			return nil
		})
	}

	if err := retrieval.Wait(); err != nil {
		log.Error(err, "retrieval wait group error", 0)
		return err
	}
	if err := indexing.Wait(); err != nil {
		log.Error(err, "indexing wait group error", 0)
		return err
	}

	lastBlockInCache, err := indexer.lastBlockStore.GetInDataTable(chainID)
	if err != nil {
		return err
	}

	if end > lastBlockInCache {
		if err := indexer.lastBlockStore.SetInDataTable(chainID, end); err != nil {
			return err
		}
	}
	return nil
}

// Index retrieve the types.Eth1Block from the client and apply the transformers to them
func (indexer *BlockIndexer) Index(chainID string, start, end uint64) error {
	if err := indexer.IndexNode(chainID, start, end); err != nil {
		return fmt.Errorf("blocks from node: %w", err)
	}
	if err := indexer.IndexEvents(chainID, start, end); err != nil {
		return fmt.Errorf("events: %w", err)
	}
	return nil
}

func (indexer *BlockIndexer) indexBlock(chainID string, block *types.Eth1Block) error {
	res := db2.IndexedBlock{
		ChainID: chainID,
		Number:  block.Number,
		Hash:    block.Hash,
	}
	for _, transformer := range indexer.transformers {
		start := time.Now()
		err := transformer.fn(chainID, block, &res)
		if err != nil {
			return fmt.Errorf("error transforming [%v] block [%v]", transformer.name, block.Number)
		}
		indexingMetrics.TransformerProcessingTime(chainID, transformer.name, time.Since(start))
	}
	if err := indexer.store.AddIndexedBlock(res); err != nil {
		return fmt.Errorf("error saving block [%v]: %w", block.Number, err)
	}
	return nil
}

type reporter struct {
	count atomic.Uint64
}

func (r *reporter) report(total uint64, block *types.Eth1Block, timings *types.GetBlockTimings, lastTickTs *time.Time, startTs, blockStartTs, dbStart time.Time) {
	count := r.count.Add(1)
	if count%100 == 0 {
		if total == 0 {
			total = 1
		}
		perc := float64(count) * 100 / float64(total)
		log.Infof("retrieved & saved block %v (0x%x) in %v (header: %v, receipts: %v, traces: %v, db: %v)", block.Number, block.Hash, time.Since(blockStartTs), timings.Headers, timings.Receipts, timings.Traces, time.Since(dbStart))
		log.Infof("processed %v blocks in %v (%.1f blocks / sec); sync is %.1f%% complete", block.Number, time.Since(startTs), float64(block.Number)/time.Since(*lastTickTs).Seconds(), perc)

		*lastTickTs = time.Now()
	}
}
