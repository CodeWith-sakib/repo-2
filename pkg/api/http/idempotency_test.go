package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIdempotencyMiddleware_Replay(t *testing.T) {
	store := NewIdempotencyStore(time.Hour)
	mw := IdempotencyMiddleware(store)

	callCount := 0
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"res-123"}`))
	}))

	// Request 1
	req1 := httptest.NewRequest(http.MethodPost, "/api/resource", bytes.NewReader([]byte(`{"name":"test"}`)))
	req1.Header.Set("Idempotency-Key", "key-001")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusCreated {
		t.Fatalf("req1 code: got %d, want 201", rec1.Code)
	}
	if callCount != 1 {
		t.Fatalf("expected handler called once, got %d", callCount)
	}

	// Request 2 (identical key and body -> should replay from cache without calling inner handler)
	req2 := httptest.NewRequest(http.MethodPost, "/api/resource", bytes.NewReader([]byte(`{"name":"test"}`)))
	req2.Header.Set("Idempotency-Key", "key-001")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusCreated {
		t.Errorf("req2 code: got %d, want 201", rec2.Code)
	}
	if rec2.Body.String() != `{"id":"res-123"}` {
		t.Errorf("unexpected body: %s", rec2.Body.String())
	}
	if callCount != 1 {
		t.Errorf("handler called again! callCount=%d", callCount)
	}
	if rec2.Header().Get("X-Cache-Lookup") != "HIT-IDEMPOTENT" {
		t.Error("expected HIT-IDEMPOTENT header")
	}
}

func TestIdempotencyMiddleware_PayloadMismatch(t *testing.T) {
	store := NewIdempotencyStore(time.Hour)
	mw := IdempotencyMiddleware(store)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Initial request
	req1 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{"foo":"bar"}`)))
	req1.Header.Set("Idempotency-Key", "key-002")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	// Second request with different body but same key -> 422 Unprocessable Entity
	req2 := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte(`{"foo":"different"}`)))
	req2.Header.Set("Idempotency-Key", "key-002")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 on payload mismatch, got %d", rec2.Code)
	}
}
