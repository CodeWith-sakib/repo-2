package http

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// TokenBucket implements the token-bucket rate limiting algorithm.
type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// NewTokenBucket creates a token bucket with the given burst capacity and refill rate.
func NewTokenBucket(maxTokens, refillPerSecond float64) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillPerSecond,
		lastRefill: time.Now(),
	}
}

// Allow returns true if n tokens can be consumed from the bucket.
func (b *TokenBucket) Allow(n float64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.lastRefill = now

	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}

	if b.tokens >= n {
		b.tokens -= n
		return true
	}
	return false
}

// Available returns the approximate number of tokens currently available.
func (b *TokenBucket) Available() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	elapsed := time.Since(b.lastRefill).Seconds()
	tokens := b.tokens + elapsed*b.refillRate
	if tokens > b.maxTokens {
		tokens = b.maxTokens
	}
	return tokens
}

// TenantRateLimiterConfig specifies per-tenant rate limit parameters.
type TenantRateLimiterConfig struct {
	BurstSize  float64
	RatePerSec float64
	CostPerReq float64
}

// DefaultTenantRateLimiterConfig returns conservative defaults.
func DefaultTenantRateLimiterConfig() TenantRateLimiterConfig {
	return TenantRateLimiterConfig{
		BurstSize:  100,
		RatePerSec: 50,
		CostPerReq: 1,
	}
}

// TenantHTTPRateLimiter applies per-tenant token-bucket rate limiting to HTTP handlers.
type TenantHTTPRateLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*TokenBucket
	defaults TenantRateLimiterConfig
	configs  map[string]TenantRateLimiterConfig
}

// NewTenantHTTPRateLimiter creates a rate limiter with default config for unknown tenants.
func NewTenantHTTPRateLimiter(defaults TenantRateLimiterConfig) *TenantHTTPRateLimiter {
	return &TenantHTTPRateLimiter{
		buckets:  make(map[string]*TokenBucket),
		defaults: defaults,
		configs:  make(map[string]TenantRateLimiterConfig),
	}
}

// SetTenantConfig configures per-tenant rate limits.
func (l *TenantHTTPRateLimiter) SetTenantConfig(tenantID string, cfg TenantRateLimiterConfig) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.configs[tenantID] = cfg
	l.buckets[tenantID] = NewTokenBucket(cfg.BurstSize, cfg.RatePerSec)
}

// Allow checks if a tenant's request should be allowed.
func (l *TenantHTTPRateLimiter) Allow(tenantID string) bool {
	l.mu.RLock()
	bucket, ok := l.buckets[tenantID]
	cfg, cfgOk := l.configs[tenantID]
	l.mu.RUnlock()

	if !ok {
		// Create bucket on demand with defaults
		l.mu.Lock()
		if b, ok2 := l.buckets[tenantID]; ok2 {
			bucket = b
		} else {
			bucket = NewTokenBucket(l.defaults.BurstSize, l.defaults.RatePerSec)
			l.buckets[tenantID] = bucket
		}
		l.mu.Unlock()
	}

	cost := l.defaults.CostPerReq
	if cfgOk {
		cost = cfg.CostPerReq
	}

	return bucket.Allow(cost)
}

// Middleware returns an HTTP middleware that enforces per-tenant rate limits.
// The tenantID is extracted from the X-Tenant-ID header.
func (l *TenantHTTPRateLimiter) Middleware() MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := r.Header.Get("X-Tenant-ID")
			if tenantID == "" {
				tenantID = "default"
			}

			if !l.Allow(tenantID) {
				w.Header().Set("Retry-After", "1")
				w.Header().Set("X-RateLimit-Tenant", tenantID)
				http.Error(w, fmt.Sprintf("rate limit exceeded for tenant %s", tenantID), http.StatusTooManyRequests)
				return
			}

			w.Header().Set("X-RateLimit-Tenant", tenantID)
			next.ServeHTTP(w, r)
		})
	}
}

// BucketStats returns diagnostic info for a tenant's bucket.
func (l *TenantHTTPRateLimiter) BucketStats(tenantID string) string {
	l.mu.RLock()
	bucket, ok := l.buckets[tenantID]
	l.mu.RUnlock()

	if !ok {
		return fmt.Sprintf("tenant=%s: no bucket", tenantID)
	}
	return fmt.Sprintf("tenant=%s tokens=%.2f max=%.2f rate=%.2f/s",
		tenantID, bucket.Available(), bucket.maxTokens, bucket.refillRate)
}
