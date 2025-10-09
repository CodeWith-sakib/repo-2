package http

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"sync"
	"time"
)

// IdempotencyStatus represents the processing state of an idempotent request.
type IdempotencyStatus string

const (
	IdempotencyInFlight  IdempotencyStatus = "IN_FLIGHT"
	IdempotencyCompleted IdempotencyStatus = "COMPLETED"
	IdempotencyFailed    IdempotencyStatus = "FAILED"
)

// CachedResponse stores the completed response for replay.
type CachedResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// IdempotencyRecord tracks the lifecycle of an Idempotency-Key.
type IdempotencyRecord struct {
	Key         string
	Fingerprint string
	Status      IdempotencyStatus
	CreatedAt   time.Time
	Response    *CachedResponse
}

// IdempotencyStore manages concurrency and replay cache for idempotency keys.
type IdempotencyStore struct {
	mu      sync.RWMutex
	records map[string]*IdempotencyRecord
	ttl     time.Duration
}

// NewIdempotencyStore creates an in-memory store with given TTL.
func NewIdempotencyStore(ttl time.Duration) *IdempotencyStore {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &IdempotencyStore{
		records: make(map[string]*IdempotencyRecord),
		ttl:     ttl,
	}
}

// IdempotencyMiddleware returns an HTTP middleware enforcing idempotency.
func IdempotencyMiddleware(store *IdempotencyStore) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Idempotency-Key")
			if key == "" || (r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch) {
				next.ServeHTTP(w, r)
				return
			}

			// Read and hash body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			hash := sha256.Sum256(bodyBytes)
			fingerprint := hex.EncodeToString(hash[:])

			store.mu.Lock()
			rec, exists := store.records[key]
			if exists {
				// Check expiration
				if time.Since(rec.CreatedAt) > store.ttl {
					delete(store.records, key)
					exists = false
				}
			}

			if exists {
				if rec.Fingerprint != fingerprint {
					store.mu.Unlock()
					http.Error(w, "idempotency key reused with different payload", http.StatusUnprocessableEntity)
					return
				}

				if rec.Status == IdempotencyInFlight {
					store.mu.Unlock()
					http.Error(w, "concurrent request in progress with same idempotency key", http.StatusConflict)
					return
				}

				if rec.Status == IdempotencyCompleted && rec.Response != nil {
					store.mu.Unlock()
					// Replay cached response
					for k, vv := range rec.Response.Header {
						for _, v := range vv {
							w.Header().Add(k, v)
						}
					}
					w.Header().Set("X-Cache-Lookup", "HIT-IDEMPOTENT")
					w.WriteHeader(rec.Response.StatusCode)
					_, _ = w.Write(rec.Response.Body)
					return
				}
			}

			// Reserve key as in-flight
			store.records[key] = &IdempotencyRecord{
				Key:         key,
				Fingerprint: fingerprint,
				Status:      IdempotencyInFlight,
				CreatedAt:   time.Now(),
			}
			store.mu.Unlock()

			// Intercept recorder
			recW := &responseInterceptor{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(recW, r)

			// Store completed response
			store.mu.Lock()
			if rEntry, ok := store.records[key]; ok {
				rEntry.Status = IdempotencyCompleted
				rEntry.Response = &CachedResponse{
					StatusCode: recW.statusCode,
					Header:     recW.Header().Clone(),
					Body:       recW.body.Bytes(),
				}
			}
			store.mu.Unlock()
		})
	}
}

type responseInterceptor struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseInterceptor) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseInterceptor) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
