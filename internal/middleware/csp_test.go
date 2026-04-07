package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSP_SetsHeader(t *testing.T) {
	handler := CSP("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	csp := w.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("CSP header not set")
	}
	if !strings.Contains(csp, "default-src 'self'") {
		t.Error("missing default-src")
	}
	if !strings.Contains(csp, "nonce-") {
		t.Error("missing nonce")
	}
	if !strings.Contains(csp, "https://js.stripe.com") {
		t.Error("missing Stripe script source")
	}
}

func TestCSP_DevMode(t *testing.T) {
	handler := CSP("http://localhost:5173")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	csp := w.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "http://localhost:5173") {
		t.Error("missing Vite dev URL in CSP")
	}
	if !strings.Contains(csp, "ws://localhost:5173") {
		t.Error("missing Vite WebSocket URL in CSP")
	}
}

func TestCSP_NonceInContext(t *testing.T) {
	var nonce string
	handler := CSP("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce = NonceFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if nonce == "" {
		t.Error("nonce not found in context")
	}
	if len(nonce) < 16 {
		t.Error("nonce too short")
	}
}

func TestCSP_NonceUnique(t *testing.T) {
	var nonces []string
	handler := CSP("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonces = append(nonces, NonceFromContext(r.Context()))
		w.WriteHeader(http.StatusOK)
	}))
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
	seen := make(map[string]bool)
	for _, n := range nonces {
		if seen[n] {
			t.Errorf("duplicate nonce: %s", n)
		}
		seen[n] = true
	}
}

func TestNonceFromContext_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if n := NonceFromContext(req.Context()); n != "" {
		t.Errorf("expected empty nonce, got %q", n)
	}
}
