package middleware

import (
	"net/http"
)

// InterceptorDeps is the tiny API the interceptor needs.
type InterceptorDeps interface {
	Inc(method, code string)
}

// MetricsHTTPMiddleware records one cumulative count per request
// with labels (method, code).
func MetricsHTTPMiddleware(w InterceptorDeps) func(http.Handler) http.Handler {
	if w == nil {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(rw, r)
			})
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: rw, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			w.Inc(r.Method, http.StatusText(rec.status))
		})
	}
}

// statusRecorder wraps http.ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
