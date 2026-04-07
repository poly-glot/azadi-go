package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type mockAuthenticator struct {
	session *SessionData
	err     error
}

func (m *mockAuthenticator) Authenticate(_ context.Context, _, _, _ string) (*SessionData, error) {
	return m.session, m.err
}

func authTestRender(w http.ResponseWriter, _ *http.Request, name string, _ map[string]any) {
	w.Header().Set("X-Template", name)
	w.WriteHeader(http.StatusOK)
}

func newTestHandler(auth *mockAuthenticator) *Handler {
	ss, _ := NewSessionStore(context.Background(), "12345678901234567890123456789012", false)
	return NewHandler(auth, ss, NewFlashStore(), true, authTestRender)
}

func TestLoginPage(t *testing.T) {
	h := newTestHandler(&mockAuthenticator{})
	req := httptest.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()
	h.LoginPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("X-Template"); got != "login.html" {
		t.Errorf("template = %q, want %q", got, "login.html")
	}
}

func TestLogin_Success(t *testing.T) {
	auth := &mockAuthenticator{
		session: &SessionData{CustomerID: "CUST-001", CustomerName: "John", AgreementNum: "AGR-001"},
	}
	h := newTestHandler(auth)

	form := url.Values{
		"username": {"AGR-001"},
		"password": {"1/1/1990|SW1A 1AA"},
	}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc != "/my-account" {
		t.Errorf("redirect = %q, want /my-account", loc)
	}
	// Check session cookie was set
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "__session" {
			found = true
		}
	}
	if !found {
		t.Error("expected __session cookie")
	}
}

func TestLogin_BadPassword(t *testing.T) {
	h := newTestHandler(&mockAuthenticator{})

	form := url.Values{
		"username": {"AGR-001"},
		"password": {"no-pipe-separator"},
	}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc != "/login-error" {
		t.Errorf("redirect = %q, want /login-error", loc)
	}
}

func TestLogin_AuthFailure(t *testing.T) {
	auth := &mockAuthenticator{err: errors.New("bad credentials")}
	h := newTestHandler(auth)

	form := url.Values{
		"username": {"AGR-001"},
		"password": {"1/1/1990|SW1A1AA"},
	}
	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc != "/login-error" {
		t.Errorf("redirect = %q, want /login-error", loc)
	}
}

func TestLoginError(t *testing.T) {
	h := newTestHandler(&mockAuthenticator{})
	req := httptest.NewRequest("GET", "/login-error", nil)
	w := httptest.NewRecorder()
	h.LoginError(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestLogout(t *testing.T) {
	h := newTestHandler(&mockAuthenticator{})
	req := httptest.NewRequest("GET", "/logout", nil)
	w := httptest.NewRecorder()
	h.Logout(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc != "/login" {
		t.Errorf("redirect = %q, want /login", loc)
	}
}

func TestSplitOnce(t *testing.T) {
	tests := []struct {
		input string
		sep   string
		want  []string
	}{
		{"a|b", "|", []string{"a", "b"}},
		{"1/1/1990|SW1A1AA", "|", []string{"1/1/1990", "SW1A1AA"}},
		{"no separator", "|", []string{"no separator"}},
		{"a|b|c", "|", []string{"a", "b|c"}},
		{"", "|", []string{""}},
	}
	for _, tt := range tests {
		got := splitOnce(tt.input, tt.sep)
		if len(got) != len(tt.want) {
			t.Errorf("splitOnce(%q, %q) = %v, want %v", tt.input, tt.sep, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitOnce(%q, %q)[%d] = %q, want %q", tt.input, tt.sep, i, got[i], tt.want[i])
			}
		}
	}
}

// FlashStore and session tests are in session_test.go
