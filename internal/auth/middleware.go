package auth

import (
	"context"
	"net/http"
)

type contextKey string

const (
	customerKey  contextKey = "customer"
	sessionIDKey contextKey = "sessionID"
)

// ContextWithSession injects session data into a context.
// Used by RequireAuth middleware and test helpers.
func ContextWithSession(ctx context.Context, data *SessionData, sessionID string) context.Context {
	ctx = context.WithValue(ctx, customerKey, data)
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)
	return ctx
}

func RequireAuth(sessions *SessionStore, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, sessionID := sessions.Get(r)
		if data == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		ctx := ContextWithSession(r.Context(), data, sessionID)
		next(w, r.WithContext(ctx))
	})
}

// OptionalAuth loads session data if present but does not redirect.
func OptionalAuth(sessions *SessionStore, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, sessionID := sessions.Get(r)
		if data != nil {
			ctx := ContextWithSession(r.Context(), data, sessionID)
			next(w, r.WithContext(ctx))
			return
		}
		next(w, r)
	})
}

func CustomerFromContext(ctx context.Context) *SessionData {
	data, _ := ctx.Value(customerKey).(*SessionData)
	return data
}

func SessionIDFromContext(ctx context.Context) string {
	s, _ := ctx.Value(sessionIDKey).(string)
	return s
}
