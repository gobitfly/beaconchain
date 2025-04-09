package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	IndexingBlockDifference = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "indexing_block_difference",
		Help: "Difference between the latest on-chain block and the last indexed block",
	}, []string{"chain_id"})

	IndexingTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "indexing_time_seconds",
		Help: "Time taken to index the new blocks",
	}, []string{"chain_id"})

	IndexingTransformerProcessingTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "indexing_transformer_processing_time_seconds",
		Help: "Time taken by a transformer to index a block",
	}, []string{"chain_id", "transformer"})

	IndexingPendingBalanceUpdate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "indexing_pending_balance_update",
		Help: "Pending balances to update",
	}, []string{"chain_id"})

	IndexingPendingENSUpdate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "indexing_pending_ens_update",
		Help: "Pending ENS to update",
	}, []string{"chain_id"})

	IndexingReorgTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "indexing_reorg_total",
		Help: "Total number of block impacted by chain reorganizations",
	}, []string{"chain_id"})

	ClientGetBlock = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "client_get_block_time_seconds",
		Help: "Time taken to retrieve a block from the blockchain and parse it",
	}, []string{"chain_id"})
)
