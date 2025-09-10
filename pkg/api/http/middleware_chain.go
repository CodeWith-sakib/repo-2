package http

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// MiddlewareFunc is a function that wraps an http.Handler.
type MiddlewareFunc func(http.Handler) http.Handler

// MiddlewareEntry wraps a middleware with metadata for chaining and telemetry.
type MiddlewareEntry struct {
	Name       string
	Priority   int
	Middleware MiddlewareFunc
}

// MiddlewareChain manages an ordered list of HTTP middleware with telemetry hooks.
type MiddlewareChain struct {
	mu          sync.RWMutex
	middlewares []MiddlewareEntry
	metrics     *ChainMetrics
}

// ChainMetrics tracks per-middleware execution statistics.
type ChainMetrics struct {
	mu         sync.Mutex
	execCounts map[string]int64
	totalDur   map[string]time.Duration
	errCounts  map[string]int64
}

// NewChainMetrics creates an empty metrics tracker.
func NewChainMetrics() *ChainMetrics {
	return &ChainMetrics{
		execCounts: make(map[string]int64),
		totalDur:   make(map[string]time.Duration),
		errCounts:  make(map[string]int64),
	}
}

// Record adds an execution observation.
func (m *ChainMetrics) Record(name string, dur time.Duration, hadError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.execCounts[name]++
	m.totalDur[name] += dur
	if hadError {
		m.errCounts[name]++
	}
}

// Snapshot returns a copy of all metrics.
func (m *ChainMetrics) Snapshot() map[string]MiddlewareStats {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]MiddlewareStats, len(m.execCounts))
	for name, count := range m.execCounts {
		avgDur := time.Duration(0)
		if count > 0 {
			avgDur = m.totalDur[name] / time.Duration(count)
		}
		out[name] = MiddlewareStats{
			Name:      name,
			ExecCount: count,
			ErrCount:  m.errCounts[name],
			AvgDur:    avgDur,
		}
	}
	return out
}

// MiddlewareStats is a point-in-time snapshot of per-middleware metrics.
type MiddlewareStats struct {
	Name      string
	ExecCount int64
	ErrCount  int64
	AvgDur    time.Duration
}

// NewMiddlewareChain initializes an empty middleware chain.
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		metrics: NewChainMetrics(),
	}
}

// Use registers a middleware at the given priority (lower number = outer wrap).
func (c *MiddlewareChain) Use(name string, priority int, mw MiddlewareFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := MiddlewareEntry{Name: name, Priority: priority, Middleware: mw}

	// Insert in sorted order (ascending priority = outermost first)
	inserted := false
	for i, existing := range c.middlewares {
		if priority < existing.Priority {
			tail := make([]MiddlewareEntry, len(c.middlewares[i:]))
			copy(tail, c.middlewares[i:])
			c.middlewares = append(c.middlewares[:i], entry)
			c.middlewares = append(c.middlewares, tail...)
			inserted = true
			break
		}
	}
	if !inserted {
		c.middlewares = append(c.middlewares, entry)
	}
}

// Remove deregisters a named middleware.
func (c *MiddlewareChain) Remove(name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, m := range c.middlewares {
		if m.Name == name {
			c.middlewares = append(c.middlewares[:i], c.middlewares[i+1:]...)
			return true
		}
	}
	return false
}

// Build compiles the current chain into a single http.Handler wrapping the base handler.
func (c *MiddlewareChain) Build(base http.Handler) http.Handler {
	c.mu.RLock()
	mws := make([]MiddlewareEntry, len(c.middlewares))
	copy(mws, c.middlewares)
	c.mu.RUnlock()

	h := base
	for i := len(mws) - 1; i >= 0; i-- {
		entry := mws[i]
		name := entry.Name
		mw := entry.Middleware
		inner := h
		metrics := c.metrics
		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			defer func() {
				metrics.Record(name, time.Since(start), false)
			}()
			mw(inner).ServeHTTP(w, r)
		})
	}
	return h
}

// Metrics returns the chain metrics tracker.
func (c *MiddlewareChain) Metrics() *ChainMetrics {
	return c.metrics
}

// Len returns the number of registered middlewares.
func (c *MiddlewareChain) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.middlewares)
}

// RequestIDMiddleware injects a unique request ID into the request context.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = fmt.Sprintf("req-%d", time.Now().UnixNano())
		}
		ctx := context.WithValue(r.Context(), contextKeyRequestID{}, reqID)
		w.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type contextKeyRequestID struct{}

// TimeoutMiddleware adds a request processing deadline.
func TimeoutMiddleware(timeout time.Duration) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// PanicRecoveryMiddleware catches panics and returns 500.
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// MaxBodySizeMiddleware limits request body size.
func MaxBodySizeMiddleware(maxBytes int64) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
