package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddlewareChain_PriorityOrder(t *testing.T) {
	chain := NewMiddlewareChain()

	var execOrder []string

	makeMiddleware := func(name string) MiddlewareFunc {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				execOrder = append(execOrder, name+":before")
				next.ServeHTTP(w, r)
				execOrder = append(execOrder, name+":after")
			})
		}
	}

	chain.Use("auth", 10, makeMiddleware("auth"))
	chain.Use("logging", 1, makeMiddleware("logging"))
	chain.Use("tracing", 5, makeMiddleware("tracing"))

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		execOrder = append(execOrder, "handler")
	})

	h := chain.Build(base)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	// Priority 1 (logging) should be outermost
	if len(execOrder) != 7 {
		t.Errorf("expected 7 execution events, got %d: %v", len(execOrder), execOrder)
	}
	if execOrder[0] != "logging:before" {
		t.Errorf("logging should be outermost, got %s", execOrder[0])
	}
}

func TestMiddlewareChain_PanicRecovery(t *testing.T) {
	chain := NewMiddlewareChain()
	chain.Use("recovery", 1, func(next http.Handler) http.Handler {
		return PanicRecoveryMiddleware(next)
	})

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("intentional panic")
	})

	h := chain.Build(base)
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestMiddlewareChain_TimeoutMiddleware(t *testing.T) {
	chain := NewMiddlewareChain()
	chain.Use("timeout", 1, TimeoutMiddleware(50*time.Millisecond))

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fast handler - should not time out
		w.WriteHeader(http.StatusOK)
	})

	h := chain.Build(base)
	req := httptest.NewRequest("GET", "/fast", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestMiddlewareChain_RemoveMiddleware(t *testing.T) {
	chain := NewMiddlewareChain()
	chain.Use("mw-a", 1, RequestIDMiddleware)
	chain.Use("mw-b", 2, RequestIDMiddleware)

	if chain.Len() != 2 {
		t.Fatalf("expected 2 middlewares, got %d", chain.Len())
	}

	removed := chain.Remove("mw-a")
	if !removed {
		t.Error("expected successful removal")
	}
	if chain.Len() != 1 {
		t.Errorf("expected 1 middleware after removal, got %d", chain.Len())
	}
}
