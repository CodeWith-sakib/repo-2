package http

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipCompressionMiddleware_Compaction(t *testing.T) {
	mw := GzipCompressionMiddleware(CompressionConfig{MinLengthBytes: 100})

	largePayload := strings.Repeat("Repeated JSON telemetry string payload\n", 20)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(largePayload))
	}))

	// Request with gzip support
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding gzip, got %s", rec.Header().Get("Content-Encoding"))
	}

	// Decompress and verify
	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("decompress error: %v", err)
	}

	if string(decompressed) != largePayload {
		t.Error("decompressed output did not match original")
	}
}

func TestGzipCompressionMiddleware_NoAcceptEncoding(t *testing.T) {
	mw := GzipCompressionMiddleware(CompressionConfig{MinLengthBytes: 100})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("plain content"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "" {
		t.Error("expected no Content-Encoding when client does not support gzip")
	}
}
