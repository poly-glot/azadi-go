package auth

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type rateBucket struct {
	count   int
	resetAt time.Time
}

type RateLimiter struct {
	mu             sync.Mutex
	loginBuckets   map[string]*rateBucket
	paymentBuckets map[string]*rateBucket
	generalBuckets map[string]*rateBucket
}

const (
	loginMax   = 5
	paymentMax = 3
	generalMax = 60
)

var (
	loginWindow   = 15 * time.Minute
	paymentWindow = 1 * time.Hour
	generalWindow = 1 * time.Minute
)

func NewRateLimiter(ctx context.Context) *RateLimiter {
	rl := &RateLimiter{
		loginBuckets:   make(map[string]*rateBucket),
		paymentBuckets: make(map[string]*rateBucket),
		generalBuckets: make(map[string]*rateBucket),
	}
	go rl.cleanupLoop(ctx)
	return rl
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		if method == "POST" && path == "/login" {
			ip := extractClientIP(r)
			if rl.isLimited(rl.loginBuckets, "login:"+ip, loginMax, loginWindow) {
				sendTooManyRequests(w, loginWindow)
				return
			}
		}

		if method == "POST" && strings.Contains(path, "/make-a-payment") {
			key := resolveClientKey(r)
			if rl.isLimited(rl.paymentBuckets, "pay:"+key, paymentMax, paymentWindow) {
				sendTooManyRequests(w, paymentWindow)
				return
			}
		}

		key := resolveClientKey(r)
		if rl.isLimited(rl.generalBuckets, "gen:"+key, generalMax, generalWindow) {
			sendTooManyRequests(w, generalWindow)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) isLimited(buckets map[string]*rateBucket, key string, max int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := buckets[key]
	if !ok || time.Now().After(b.resetAt) {
		b = &rateBucket{count: 0, resetAt: time.Now().Add(window)}
		buckets[key] = b
	}
	b.count++
	return b.count > max
}

func (rl *RateLimiter) ClearAll() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.loginBuckets = make(map[string]*rateBucket)
	rl.paymentBuckets = make(map[string]*rateBucket)
	rl.generalBuckets = make(map[string]*rateBucket)
}

func (rl *RateLimiter) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()
			cleanMap(rl.loginBuckets, time.Now())
			cleanMap(rl.paymentBuckets, time.Now())
			cleanMap(rl.generalBuckets, time.Now())
			rl.mu.Unlock()
		}
	}
}

func cleanMap(m map[string]*rateBucket, now time.Time) {
	for k, b := range m {
		if now.After(b.resetAt) {
			delete(m, k)
		}
	}
}

func sendTooManyRequests(w http.ResponseWriter, retryAfter time.Duration) {
	w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
	http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
}

func extractClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.SplitN(forwarded, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	return r.RemoteAddr
}

func resolveClientKey(r *http.Request) string {
	// Try session cookie first, fall back to IP
	if c, err := r.Cookie("__session"); err == nil && c.Value != "" {
		return c.Value[:min(len(c.Value), 32)]
	}
	return extractClientIP(r)
}
