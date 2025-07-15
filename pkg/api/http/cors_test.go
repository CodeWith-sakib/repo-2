package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomCORSMiddleware(t *testing.T) {
	handler := CustomCORSMiddleware(CORSSettings{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("OPTIONS", "/api/v1/workflows", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected allow origin *")
	}
}
