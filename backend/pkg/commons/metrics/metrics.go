package metrics

import (
	"log" //nolint:depguard
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/version"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Version = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "version",
		Help: "Gauge with version-string in label",
	}, []string{"version"})
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of requests by path, method and status_code.",
	}, []string{"path", "method", "status_code"})
	HttpRequestsInFlight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Current requests being served.",
	}, []string{"path", "method"})
	HttpRequestsDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_requests_duration",
		Help: "Duration of HTTP requests in seconds by path and method.",
	}, []string{"path", "method"})
	Tasks = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_counter",
		Help: "Counter of tasks with name in labels",
	}, []string{"name"})
	TaskDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "task_duration",
		Help:    "Duration of tasks",
		Buckets: []float64{.05, .1, .5, 1, 5, 10, 20, 60, 90, 120, 180, 300},
	}, []string{"task"})
	DBSLongRunningQueries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "db_long_running_queries",
		Help: "Counter of long-running-queries with database and query in labels",
	}, []string{"database", "query"})
	Errors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "errors",
		Help: "Counter of errors with name in labels",
	}, []string{"name"})
	NotificationsCollected = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notifications_collected",
		Help: "Counter of notification event type that gets collected",
	}, []string{"event_type"})
	NotificationsQueued = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notifications_queued",
		Help: "Counter of notification channel and event type that gets queued",
	}, []string{"channel", "event_type"})
	NotificationsSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notifications_sent",
		Help: "Counter of notifications sent with the channel and notification type in the label",
	}, []string{"channel", "status"})
	State = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "state",
		Help: "Gauge for various states",
	}, []string{"state"})
	Counter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "counter",
		Help: "Generic counter of events with name in labels",
	}, []string{"name"})
)

func init() {
	Version.WithLabelValues(version.Version).Set(1)
}

// HttpMiddleware implements mux.MiddlewareFunc.
// This middleware uses the path template, so the label value will be /obj/{id} rather than /obj/123 which would risk a cardinality explosion.
// See https://www.robustperception.io/prometheus-middleware-for-gorilla-mux
func HttpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		route := mux.CurrentRoute(r)
		path, err := route.GetPathTemplate()
		if err != nil {
			path = "UNDEFINED"
		}
		method := strings.ToUpper(r.Method)
		HttpRequestsInFlight.WithLabelValues(path, method).Inc()
		defer HttpRequestsInFlight.WithLabelValues(path, method).Dec()
		d := &responseWriterDelegator{ResponseWriter: w}
		next.ServeHTTP(d, r)
		status := strconv.Itoa(d.status)
		HttpRequestsTotal.WithLabelValues(path, method, status).Inc()
		HttpRequestsDuration.WithLabelValues(path, method).Observe(time.Since(start).Seconds())
	})
}

type responseWriterDelegator struct {
	http.ResponseWriter
	status      int
	written     int64
	wroteHeader bool
}

func (r *responseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseWriterDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

// Serve serves prometheus metrics on the given address under /metrics
func Serve(addr string, servePprof bool) error {
	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())
	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
<head><title>prometheus-metrics</title></head>
<body>
<h1>prometheus-metrics</h1>
<p><a href='/metrics'>metrics</a></p>
</body>
</html>`))
		if err != nil {
			log.Println(err, "error writing to response buffer: %v", 0)
		}
	}))

	if servePprof {
		log.Printf("serving pprof on %v/debug/pprof/", addr)
		router.HandleFunc("/debug/pprof/", pprof.Index)
		router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		router.HandleFunc("/debug/pprof/profile", pprof.Profile)
		router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
		router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	}

	srv := &http.Server{
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		Handler:      router,
		Addr:         addr,
	}

	return srv.ListenAndServe()
}
