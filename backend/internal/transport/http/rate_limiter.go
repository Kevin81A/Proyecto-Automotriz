package http

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter tracks login attempts per IP address to prevent brute force attacks.
type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
	now      func() time.Time
}

// NewRateLimiter creates an in-memory rate limiter.
func NewRateLimiter(limit int, window time.Duration, now func() time.Time) *RateLimiter {
	if now == nil {
		now = time.Now
	}
	return &RateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
		now:      now,
	}
}

// Allow returns true if the given IP has not exceeded the attempt limit in the current window.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	current := rl.now()
	cutoff := current.Add(-rl.window)

	valid := rl.attempts[ip][:0]
	for _, moment := range rl.attempts[ip] {
		if moment.After(cutoff) {
			valid = append(valid, moment)
		}
	}

	if len(valid) >= rl.limit {
		rl.attempts[ip] = valid
		return false
	}

	rl.attempts[ip] = append(valid, current)
	return true
}

// Middleware wraps an HTTP handler with rate limiting by client IP.
func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ip := clientIP(request)
		if !rl.Allow(ip) {
			respond(writer, http.StatusTooManyRequests, errorPayload{
				Code:    "too_many_requests",
				Message: "Demasiados intentos de inicio de sesion. Intente de nuevo mas tarde.",
			})
			return
		}
		next(writer, request)
	}
}

// clientIP extracts the client IP address from request headers or remote address.
func clientIP(request *http.Request) string {
	forwarded := request.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		first := strings.TrimSpace(parts[0])
		if first != "" {
			return first
		}
	}
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if request.RemoteAddr != "" {
		return request.RemoteAddr
	}
	return "unknown"
}
