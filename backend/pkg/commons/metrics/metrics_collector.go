package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricsRepository interface {
	Error(operation string)
	ObserveTaskDuration(operation string, duration time.Duration)
	ObserveClientCallDuration(client, method string, duration time.Duration)
	SetStateMetric(state string, value uint64)
}

type MetricsCollector struct {
	errors             *prometheus.CounterVec
	taskDuration       *prometheus.HistogramVec
	clientCallDuration *prometheus.HistogramVec
	state              *prometheus.GaugeVec
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		errors:             Errors,
		taskDuration:       TaskDuration,
		clientCallDuration: ClientCallDuration,
		state:              State,
	}
}

func (m *MetricsCollector) Error(operation string) {
	m.errors.WithLabelValues(operation).Inc()
}

func (m *MetricsCollector) ObserveTaskDuration(operation string, duration time.Duration) {
	m.taskDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

func (m *MetricsCollector) ObserveClientCallDuration(client, method string, duration time.Duration) {
	m.clientCallDuration.WithLabelValues(client, method).Observe(duration.Seconds())
}

func (m *MetricsCollector) SetStateMetric(state string, value uint64) {
	m.state.WithLabelValues(state).Set(float64(value))
}
