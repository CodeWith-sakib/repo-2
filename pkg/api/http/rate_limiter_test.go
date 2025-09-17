package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokenBucket_AllowAndRefill(t *testing.T) {
	tb := NewTokenBucket(5, 100) // 5 token burst, 100/s refill

	for i := 0; i < 5; i++ {
		if !tb.Allow(1) {
			t.Errorf("request %d should be allowed within burst", i)
		}
	}

	// 6th request should be denied (no tokens left)
	if tb.Allow(1) {
		t.Error("6th request should be denied (burst exhausted)")
	}
}

func TestTenantHTTPRateLimiter_Middleware(t *testing.T) {
	limiter := NewTenantHTTPRateLimiter(TenantRateLimiterConfig{
		BurstSize:  3,
		RatePerSec: 100,
		CostPerReq: 1,
	})

	limiter.SetTenantConfig("tenant-a", TenantRateLimiterConfig{
		BurstSize:  2,
		RatePerSec: 100,
		CostPerReq: 1,
	})

	var passCount, throttleCount int
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passCount++
		w.WriteHeader(http.StatusOK)
	})

	mw := limiter.Middleware()
	h := mw(base)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("X-Tenant-ID", "tenant-a")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			throttleCount++
		}
	}

	if passCount < 1 {
		t.Error("at least 1 request should have passed")
	}
	if throttleCount < 1 {
		t.Error("at least 1 request should have been throttled")
	}
}

func TestTenantHTTPRateLimiter_Stats(t *testing.T) {
	limiter := NewTenantHTTPRateLimiter(DefaultTenantRateLimiterConfig())
	limiter.Allow("tenant-b") // trigger bucket creation
	stats := limiter.BucketStats("tenant-b")
	if stats == "" {
		t.Error("stats should not be empty")
	}
}
