package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_LoginLimit(t *testing.T) {
	rl := NewRateLimiter(context.Background())
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d: got %d, want 200", i+1, w.Code)
		}
	}

	// 6th should be rate limited
	req := httptest.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("6th request: got %d, want 429", w.Code)
	}

	retryAfter := w.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Error("missing Retry-After header")
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(context.Background())
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// Different IP should not be limited
	req := httptest.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "5.6.7.8:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("different IP should not be limited, got %d", w.Code)
	}
}

func TestRateLimiter_GeneralLimit(t *testing.T) {
	rl := NewRateLimiter(context.Background())
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 60; i++ {
		req := httptest.NewRequest("GET", "/my-account", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// 61st should be limited
	req := httptest.NewRequest("GET", "/my-account", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("61st request: got %d, want 429", w.Code)
	}
}

func TestRateLimiter_ClearAll(t *testing.T) {
	rl := NewRateLimiter(context.Background())
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	rl.ClearAll()

	req := httptest.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("after ClearAll: got %d, want 200", w.Code)
	}
}

func TestCleanMap(t *testing.T) {
	m := map[string]*rateBucket{
		"expired": {resetAt: time.Now().Add(-time.Hour)},
		"active":  {resetAt: time.Now().Add(time.Hour)},
	}
	cleanMap(m, time.Now())
	if _, ok := m["expired"]; ok {
		t.Error("expired entry should be removed")
	}
	if _, ok := m["active"]; !ok {
		t.Error("active entry should remain")
	}
}

func TestExtractClientIP(t *testing.T) {
	tests := []struct {
		name      string
		forwarded string
		remote    string
		want      string
	}{
		{"no forwarded", "", "1.2.3.4:1234", "1.2.3.4:1234"},
		{"with forwarded", "10.0.0.1", "1.2.3.4:1234", "10.0.0.1"},
		{"forwarded chain", "10.0.0.1, 10.0.0.2", "1.2.3.4:1234", "10.0.0.1"},
		{"no port", "", "1.2.3.4", "1.2.3.4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remote
			if tt.forwarded != "" {
				req.Header.Set("X-Forwarded-For", tt.forwarded)
			}
			got := extractClientIP(req)
			if got != tt.want {
				t.Errorf("extractClientIP = %q, want %q", got, tt.want)
			}
		})
	}
}
