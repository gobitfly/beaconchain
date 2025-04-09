package executionlayer

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
)

var indexingMetrics = prometheusMetrics{}

type prometheusMetrics struct{}

func (p prometheusMetrics) BlockDifference(chainID string, difference uint64) {
	metrics.IndexingBlockDifference.WithLabelValues(chainID).Set(float64(difference))
}

func (p prometheusMetrics) IndexingTime(chainID string, t time.Duration) {
	metrics.IndexingTime.WithLabelValues(chainID).Observe(t.Seconds())
}

func (p prometheusMetrics) TransformerProcessingTime(chainID string, transformer string, t time.Duration) {
	metrics.IndexingTransformerProcessingTime.WithLabelValues(chainID, transformer).Observe(t.Seconds())
}

func (p prometheusMetrics) PendingBalanceUpdate(chainID string, count int64) {
	metrics.IndexingPendingBalanceUpdate.WithLabelValues(chainID).Set(float64(count))
}

func (p prometheusMetrics) PendingENSUpdate(chainID string, count int64) {
	metrics.IndexingPendingENSUpdate.WithLabelValues(chainID).Set(float64(count))
}

func (p prometheusMetrics) ReorgBlockTotal(chainID string, depth uint64) {
	metrics.IndexingReorgTotal.WithLabelValues(chainID).Add(float64(depth))
}
