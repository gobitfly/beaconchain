package metrics

import (
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsRepository interface {
	Error(operation string)
	ObserveDuration(operation string, duration time.Duration)
	ObserveClientCallDuration(client, method string, duration time.Duration)
}

type Metrics struct {
	errors             *prometheus.CounterVec
	duration           *prometheus.HistogramVec
	clientCallDuration *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		errors:             metrics.Errors,
		duration:           metrics.TaskDuration,
		clientCallDuration: metrics.ClientCallDuration,
	}
}

func (m *Metrics) Error(operation string) {
	m.errors.WithLabelValues(operation).Inc()
}

func (m *Metrics) ObserveDuration(operation string, duration time.Duration) {
	m.duration.WithLabelValues(operation).Observe(duration.Seconds())
}

func (m *Metrics) ObserveClientCallDuration(client, method string, duration time.Duration) {
	m.clientCallDuration.WithLabelValues(client, method).Observe(duration.Seconds())
}
