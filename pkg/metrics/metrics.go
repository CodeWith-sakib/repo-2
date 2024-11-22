package metrics

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
)

type Registry struct {
	mu       sync.RWMutex
	counters map[string]*int64
	gauges   map[string]*int64
}

var DefaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{
		counters: make(map[string]*int64),
		gauges:   make(map[string]*int64),
	}
}

func (r *Registry) IncCounter(name string, delta int64) {
	r.mu.Lock()
	ptr, exists := r.counters[name]
	if !exists {
		var val int64
		ptr = &val
		r.counters[name] = ptr
	}
	r.mu.Unlock()

	atomic.AddInt64(ptr, delta)
}

func (r *Registry) SetGauge(name string, val int64) {
	r.mu.Lock()
	ptr, exists := r.gauges[name]
	if !exists {
		var v int64
		ptr = &v
		r.gauges[name] = ptr
	}
	r.mu.Unlock()

	atomic.StoreInt64(ptr, val)
}

func (r *Registry) WritePrometheus(w io.Writer) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for k, ptr := range r.counters {
		val := atomic.LoadInt64(ptr)
		fmt.Fprintf(w, "# TYPE %s counter\n%s %d\n", k, k, val)
	}

	for k, ptr := range r.gauges {
		val := atomic.LoadInt64(ptr)
		fmt.Fprintf(w, "# TYPE %s gauge\n%s %d\n", k, k, val)
	}
}

func Handler(r *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		r.WritePrometheus(w)
	}
}
