package http

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// CompressionConfig specifies response compression boundaries.
type CompressionConfig struct {
	MinLengthBytes int // minimum body size to compress
}

// DefaultCompressionConfig returns standard settings.
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		MinLengthBytes: 512,
	}
}

// GzipCompressionMiddleware provides dynamic response compression.
func GzipCompressionMiddleware(cfg CompressionConfig) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Add("Vary", "Accept-Encoding")

			bufWriter := &bufferingResponseWriter{
				ResponseWriter: w,
				buf:            &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(bufWriter, r)

			bodyBytes := bufWriter.buf.Bytes()

			// Check if should compress
			contentType := bufWriter.Header().Get("Content-Type")
			shouldCompress := len(bodyBytes) >= cfg.MinLengthBytes &&
				!strings.HasPrefix(contentType, "image/") &&
				!strings.HasPrefix(contentType, "video/") &&
				!strings.Contains(contentType, "zip")

			if !shouldCompress {
				// Flush raw
				w.WriteHeader(bufWriter.statusCode)
				_, _ = w.Write(bodyBytes)
				return
			}

			// Compress body
			var compressed bytes.Buffer
			gw := gzip.NewWriter(&compressed)
			_, _ = gw.Write(bodyBytes)
			_ = gw.Close()

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length") // length changed
			w.WriteHeader(bufWriter.statusCode)
			_, _ = io.Copy(w, &compressed)
		})
	}
}

type bufferingResponseWriter struct {
	http.ResponseWriter
	buf        *bytes.Buffer
	statusCode int
}

func (b *bufferingResponseWriter) WriteHeader(code int) {
	b.statusCode = code
}

func (b *bufferingResponseWriter) Write(data []byte) (int, error) {
	return b.buf.Write(data)
}
