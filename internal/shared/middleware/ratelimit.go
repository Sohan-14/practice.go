package middleware

import (
	"log/slog"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// clientBucket holds the state for a single client's token bucket.
type clientBucket struct {
	tokens     float64
	lastRefill time.Time
}

// RateLimiter manages token buckets for multiple clients.
type RateLimiter struct {
	mu         sync.Mutex
	clients    map[string]*clientBucket
	capacity   float64
	refillRate float64
}

// NewRateLimiter creates a rate limiter with the given capacity and refill rate.
func NewRateLimiter(capacity float64, refillRate float64) *RateLimiter {
	return &RateLimiter{
		clients:    make(map[string]*clientBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

// Allow checks if a request from the given clientID should be allowed.
func (rl *RateLimiter) Allow(clientID string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.clients[clientID]
	now := time.Now()

	if !exists {
		bucket = &clientBucket{
			tokens:     rl.capacity,
			lastRefill: now,
		}
		rl.clients[clientID] = bucket
	} else {
		elapsed := now.Sub(bucket.lastRefill).Seconds()
		bucket.tokens += elapsed * rl.refillRate

		if bucket.tokens > rl.capacity {
			bucket.tokens = rl.capacity
		}
		bucket.lastRefill = now
	}

	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true, 0
	}

	tokensNeeded := 1.0 - bucket.tokens
	retryAfterSeconds := tokensNeeded / rl.refillRate
	retryAfter := time.Duration(retryAfterSeconds * float64(time.Second))

	return false, retryAfter
}

// Cleanup removes entries that haven't been used for longer than maxAge.
func (rl *RateLimiter) Cleanup(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for clientID, bucket := range rl.clients {
		if now.Sub(bucket.lastRefill) > maxAge {
			delete(rl.clients, clientID)
		}
	}
}

// StartCleanup launches a background goroutine that periodically cleans up stale entries.
func (rl *RateLimiter) StartCleanup(interval, maxAge time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			rl.Cleanup(maxAge)
		}
	}()
}

// getClientIP extracts the IP address from the request.
func getClientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// RateLimit enforces rate limiting per client IP.
func RateLimit(rl *RateLimiter, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			allowed, retryAfter := rl.Allow(clientIP)

			if !allowed {
				retryAfterSeconds := int(math.Ceil(retryAfter.Seconds()))

				w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate_limit_exceeded","message":"Too many requests"}`))

				logger.Warn("rate limit exceeded",
					"client_ip", clientIP,
					"retry_after", retryAfterSeconds,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
