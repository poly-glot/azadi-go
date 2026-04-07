package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
)

type nonceKey struct{}

func NonceFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(nonceKey{}).(string); ok {
		return v
	}
	return ""
}

func CSP(viteDevURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nonceBytes := make([]byte, 16)
			if _, err := rand.Read(nonceBytes); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			nonce := base64.StdEncoding.EncodeToString(nonceBytes)

			ctx := context.WithValue(r.Context(), nonceKey{}, nonce)

			viteOrigin := ""
			viteWs := ""
			if viteDevURL != "" {
				viteOrigin = " " + viteDevURL
				viteWs = " " + strings.Replace(viteDevURL, "http://", "ws://", 1)
			}

			var csp strings.Builder
			csp.WriteString("default-src 'self'; ")
			csp.WriteString("script-src 'self' 'nonce-" + nonce + "' https://js.stripe.com" + viteOrigin + "; ")
			csp.WriteString("style-src 'self' 'nonce-" + nonce + "' https://fonts.googleapis.com" + viteOrigin + "; ")
			csp.WriteString("font-src 'self' https://fonts.gstatic.com; ")
			csp.WriteString("frame-src https://js.stripe.com; ")
			csp.WriteString("img-src 'self' data:" + viteOrigin + "; ")
			csp.WriteString("connect-src 'self' https://api.stripe.com" + viteOrigin + viteWs + "; ")
			csp.WriteString("base-uri 'self'; ")
			csp.WriteString("form-action 'self'; ")
			csp.WriteString("report-uri /api/csp-report;")

			w.Header().Set("Content-Security-Policy", csp.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
