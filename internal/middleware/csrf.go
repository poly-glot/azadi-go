package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

type csrfContextKey struct{}

const csrfCookieName = "__xsrf-token"
const csrfHeaderName = "X-CSRF-TOKEN"

func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for static assets — prevents token overwrites on parallel requests
		if strings.HasPrefix(r.URL.Path, "/assets/") || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Generate CSRF token if not present
		cookie, err := r.Cookie(csrfCookieName)
		if err != nil || cookie.Value == "" {
			token := generateCSRFToken()
			http.SetCookie(w, &http.Cookie{
				Name:     csrfCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: false, // JS needs to read it
				SameSite: http.SameSiteLaxMode,
			})
			cookie = &http.Cookie{Value: token}
		}

		// Validate CSRF on POST, except webhook
		if r.Method == "POST" && r.URL.Path != "/api/stripe/webhook" {
			headerToken := r.Header.Get(csrfHeaderName)
			formToken := r.FormValue("_csrf")
			token := headerToken
			if token == "" {
				token = formToken
			}
			if token != cookie.Value {
				http.Error(w, "Forbidden - CSRF token mismatch", http.StatusForbidden)
				return
			}
		}

		// Store the token in context so the render function can access it
		ctx := context.WithValue(r.Context(), csrfContextKey{}, cookie.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CSRFTokenFromRequest(r *http.Request) string {
	// First try context (always set by middleware, even on first visit)
	if token, ok := r.Context().Value(csrfContextKey{}).(string); ok && token != "" {
		return token
	}
	if c, err := r.Cookie(csrfCookieName); err == nil {
		return c.Value
	}
	return ""
}

func generateCSRFToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
