package httpserver

import (
	"net/http"
	"sync/atomic"
)

// Health reports liveness and readiness.
type Health struct{ ready atomic.Bool }

// SetReady flips readiness; false during drain.
func (h *Health) SetReady(v bool) { h.ready.Store(v) }

// Register mounts /healthz and /readyz.
func (h *Health) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !h.ready.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
