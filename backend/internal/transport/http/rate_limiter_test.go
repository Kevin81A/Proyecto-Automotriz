package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterBlocksAfterExceedingLimit(t *testing.T) {
	now := time.Now()
	clock := func() time.Time { return now }
	limiter := NewRateLimiter(5, 15*time.Minute, clock)

	handler := limiter.Middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// First 5 attempts should succeed
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/session", nil)
		req.RemoteAddr = "192.168.1.100:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("attempt %d should be allowed, got status %d", i+1, rec.Code)
		}
	}

	// 6th attempt should be rejected with 429
	req := httptest.NewRequest(http.MethodPost, "/api/session", nil)
	req.RemoteAddr = "192.168.1.100:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("6th attempt must be rejected with 429 Too Many Requests, got %d", rec.Code)
	}

	// A different IP should still be allowed
	diffReq := httptest.NewRequest(http.MethodPost, "/api/session", nil)
	diffReq.RemoteAddr = "192.168.1.101:1234"
	diffRec := httptest.NewRecorder()
	handler.ServeHTTP(diffRec, diffReq)

	if diffRec.Code != http.StatusOK {
		t.Fatalf("different IP should be allowed, got %d", diffRec.Code)
	}

	// After window expires, previous IP should be allowed again
	now = now.Add(16 * time.Minute)
	expiredReq := httptest.NewRequest(http.MethodPost, "/api/session", nil)
	expiredReq.RemoteAddr = "192.168.1.100:1234"
	expiredRec := httptest.NewRecorder()
	handler.ServeHTTP(expiredRec, expiredReq)

	if expiredRec.Code != http.StatusOK {
		t.Fatalf("attempt after window expires should be allowed, got %d", expiredRec.Code)
	}
}
