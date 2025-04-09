package executionlayer

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/executionlayer/evm"
)

type Config struct {
	Service      ServiceConfig
	Indexer      IndexerConfig
	BlockIndexer BlockIndexerConfig
	Reorg        ReorgConfig
	Batcher      evm.BatcherConfig
}

var DefaultConfig = Config{
	Service: ServiceConfig{
		TokenPriceFrequency: time.Hour,
		BlockFrequency:      12 * time.Second,
	},
	Indexer: IndexerConfig{
		BalanceUpdaterBatchSize: 1000,
		EnableENS:               false,
		ENSImportBatchSize:      200,
		Bulk:                    8000,
	},
	BlockIndexer: BlockIndexerConfig{
		Concurrency: 30,
		TraceMode:   "geth",
	},
	Reorg: ReorgConfig{
		Depth: 20,
	},
	Batcher: evm.BatcherConfig{},
}
