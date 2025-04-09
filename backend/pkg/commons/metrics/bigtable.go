package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	BigtableMetric = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "bigtable_request_time_seconds",
		Help: "Time taken by request to execute on a Bigtable table",
	}, []string{"table", "request"})
)
