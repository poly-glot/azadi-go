package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRF_SetsCookie(t *testing.T) {
	handler := CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	found := false
	for _, c := range w.Result().Cookies() {
		if c.Name == "__xsrf-token" {
			found = true
			if c.HttpOnly {
				t.Error("CSRF cookie should NOT be HttpOnly")
			}
		}
	}
	if !found {
		t.Error("__xsrf-token cookie not set")
	}
}

func TestCSRF_BlocksPostWithoutToken(t *testing.T) {
	handler := CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("POST", "/my-contact-details", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "__xsrf-token", Value: "valid-token"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", w.Code)
	}
}

func TestCSRF_AllowsPostWithMatchingToken(t *testing.T) {
	handler := CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("POST", "/my-contact-details", strings.NewReader("_csrf=valid-token"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "__xsrf-token", Value: "valid-token"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
}

func TestCSRF_AllowsPostWithHeader(t *testing.T) {
	handler := CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("POST", "/finance/make-a-payment", nil)
	req.Header.Set("X-CSRF-TOKEN", "valid-token")
	req.AddCookie(&http.Cookie{Name: "__xsrf-token", Value: "valid-token"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
}

func TestCSRF_SkipsWebhook(t *testing.T) {
	handler := CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("POST", "/api/stripe/webhook", nil)
	req.AddCookie(&http.Cookie{Name: "__xsrf-token", Value: "token"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("webhook should skip CSRF, got %d", w.Code)
	}
}

func TestCSRF_GetDoesNotValidate(t *testing.T) {
	handler := CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/any-page", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET should not validate CSRF, got %d", w.Code)
	}
}

func TestCSRF_SkipsStaticAssets(t *testing.T) {
	handler := CSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/assets/img/logo.svg", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	for _, c := range w.Result().Cookies() {
		if c.Name == "__xsrf-token" {
			t.Error("static asset requests should not set __xsrf-token cookie")
		}
	}
}

func TestCSRFTokenFromRequest(t *testing.T) {
	tests := []struct {
		name   string
		cookie *http.Cookie
		want   string
	}{
		{"with cookie", &http.Cookie{Name: "__xsrf-token", Value: "abc123"}, "abc123"},
		{"no cookie", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			got := CSRFTokenFromRequest(req)
			if got != tt.want {
				t.Errorf("CSRFTokenFromRequest() = %q, want %q", got, tt.want)
			}
		})
	}
}
