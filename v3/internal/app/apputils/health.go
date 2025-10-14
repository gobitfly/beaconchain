package apputils

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type HealthHandler struct {
	pingable map[string]Pingable
}

type healthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
}

type readyStatus struct {
	Ready     bool              `json:"ready"`
	Timestamp time.Time         `json:"timestamp,omitempty"`
	Details   map[string]string `json:"details,omitempty"`
}

type Pingable func(ctx context.Context) error

func (f Pingable) PingContext(ctx context.Context) error {
	return f(ctx)
}

func NewHealthHandler(pingable map[string]Pingable) *HealthHandler {
	return &HealthHandler{
		pingable: pingable,
	}
}

// ServeHealth handles /healthz: basic liveness / “am I alive?”
func (h *HealthHandler) ServeHealth(w http.ResponseWriter, r *http.Request) {
	resp := healthStatus{
		Status:    "SERVING",
		Timestamp: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// ServeReady handles /readyz: readiness / “am I ready to serve?”
func (h *HealthHandler) ServeReady(w http.ResponseWriter, r *http.Request) {
	ready := true
	details := make(map[string]string)
	ctx := r.Context()

	// todo: parallelize checks
	for name, p := range h.pingable {
		if err := p(ctx); err != nil {
			ready = false
			details[name] = err.Error()
		} else {
			details[name] = "ok"
		}
	}

	resp := readyStatus{
		Ready:     ready,
		Timestamp: time.Now().UTC(),
		Details:   details,
	}

	w.Header().Set("Content-Type", "application/json")
	if !ready {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
