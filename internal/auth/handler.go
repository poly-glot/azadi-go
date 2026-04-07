package auth

import (
	"context"
	"log/slog"
	"net/http"

	"azadi-go/internal/httputil"
)

type authenticator interface {
	Authenticate(ctx context.Context, agreementNumber, dobStr, postcode string) (*SessionData, error)
}

type Handler struct {
	provider authenticator
	sessions *SessionStore
	flashes  *FlashStore
	demoMode bool
	render   httputil.RenderFunc
}

func NewHandler(provider authenticator, sessions *SessionStore, flashes *FlashStore, demoMode bool,
	render httputil.RenderFunc) *Handler {
	return &Handler{
		provider: provider,
		sessions: sessions,
		flashes:  flashes,
		demoMode: demoMode,
		render:   render,
	}
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "login.html", map[string]any{
		"DemoMode": h.demoMode,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/login-error", http.StatusFound)
		return
	}

	username := r.FormValue("username") // agreement number
	password := r.FormValue("password") // "D/M/YYYY|POSTCODE"

	// Parse password field: "dob|postcode"
	parts := splitOnce(password, "|")
	if len(parts) != 2 {
		http.Redirect(w, r, "/login-error", http.StatusFound)
		return
	}
	dob := parts[0]
	postcode := parts[1]

	sessionData, err := h.provider.Authenticate(r.Context(), username, dob, postcode)
	if err != nil {
		slog.Warn("login failed", "agreement", username, "error", err)
		http.Redirect(w, r, "/login-error", http.StatusFound)
		return
	}

	if err := h.sessions.Create(w, sessionData); err != nil {
		httputil.ServerError(w, "session creation failed", err)
		return
	}

	http.Redirect(w, r, "/my-account", http.StatusFound)
}

func (h *Handler) LoginError(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "login.html", map[string]any{
		"Error":    true,
		"DemoMode": h.demoMode,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.sessions.Destroy(w, r)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func splitOnce(s, sep string) []string {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return []string{s[:i], s[i+len(sep):]}
		}
	}
	return []string{s}
}
