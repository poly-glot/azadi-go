package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testKey = "dev-only-key-change-in-prod-32ch"

func TestSessionStore_CreateAndGet(t *testing.T) {
	store, err := NewSessionStore(context.Background(), testKey, false)
	if err != nil {
		t.Fatal(err)
	}

	data := &SessionData{CustomerID: "CUST-001", CustomerName: "James", AgreementNum: "AGR-001"}
	recorder := httptest.NewRecorder()
	if err := store.Create(recorder, data); err != nil {
		t.Fatal(err)
	}

	// Extract cookie from response
	cookies := recorder.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no session cookie set")
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(cookies[0])

	got, sessionID := store.Get(req)
	if got == nil {
		t.Fatal("session not found")
	}
	if sessionID == "" {
		t.Fatal("session ID empty")
	}
	if got.CustomerID != "CUST-001" {
		t.Errorf("CustomerID = %q, want CUST-001", got.CustomerID)
	}
}

func TestSessionStore_Destroy(t *testing.T) {
	store, err := NewSessionStore(context.Background(), testKey, false)
	if err != nil {
		t.Fatal(err)
	}

	data := &SessionData{CustomerID: "CUST-001", CustomerName: "James"}
	recorder := httptest.NewRecorder()
	store.Create(recorder, data)
	cookies := recorder.Result().Cookies()

	// Now destroy
	destroyRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/logout", nil)
	req.AddCookie(cookies[0])
	store.Destroy(destroyRecorder, req)

	// Should not find session
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.AddCookie(cookies[0])
	got, _ := store.Get(req2)
	if got != nil {
		t.Error("session should be destroyed")
	}
}

func TestSessionStore_MaxOnePerUser(t *testing.T) {
	store, err := NewSessionStore(context.Background(), testKey, false)
	if err != nil {
		t.Fatal(err)
	}

	data := &SessionData{CustomerID: "CUST-001", CustomerName: "James"}

	// Create first session
	r1 := httptest.NewRecorder()
	store.Create(r1, data)
	cookie1 := r1.Result().Cookies()[0]

	// Create second session (same customer)
	r2 := httptest.NewRecorder()
	store.Create(r2, data)

	// First session should be invalidated
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(cookie1)
	got, _ := store.Get(req)
	if got != nil {
		t.Error("first session should be invalidated after new login")
	}
}

func TestSessionStore_InvalidKey(t *testing.T) {
	_, err := NewSessionStore(context.Background(), "short", false)
	if err == nil {
		t.Error("should fail with short key")
	}
}

func TestSessionStore_CookieName(t *testing.T) {
	store, _ := NewSessionStore(context.Background(), testKey, false)
	data := &SessionData{CustomerID: "CUST-001"}
	recorder := httptest.NewRecorder()
	store.Create(recorder, data)

	cookies := recorder.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "__session" {
			found = true
			if c.HttpOnly != true {
				t.Error("cookie should be HttpOnly")
			}
		}
	}
	if !found {
		t.Error("__session cookie not found")
	}
}

func TestNewFlashStore(t *testing.T) {
	fs := NewFlashStore()
	if fs == nil {
		t.Fatal("NewFlashStore returned nil")
	}
}

func TestFlashStore_SetThenGet(t *testing.T) {
	fs := NewFlashStore()
	fs.Set("sess-1", "success", "Payment completed")

	got := fs.Get("sess-1", "success")
	if got != "Payment completed" {
		t.Errorf("Get() = %q, want %q", got, "Payment completed")
	}
}

func TestFlashStore_GetRemovesValue(t *testing.T) {
	fs := NewFlashStore()
	fs.Set("sess-1", "info", "Hello")

	// First get returns the value
	first := fs.Get("sess-1", "info")
	if first != "Hello" {
		t.Errorf("first Get() = %q, want %q", first, "Hello")
	}

	// Second get returns empty (consumed)
	second := fs.Get("sess-1", "info")
	if second != "" {
		t.Errorf("second Get() = %q, want empty", second)
	}
}

func TestFlashStore_GetNonexistent(t *testing.T) {
	fs := NewFlashStore()

	// Unknown session
	if got := fs.Get("unknown-sess", "key"); got != "" {
		t.Errorf("Get(unknown session) = %q, want empty", got)
	}

	// Known session, unknown key
	fs.Set("sess-1", "exists", "value")
	if got := fs.Get("sess-1", "missing"); got != "" {
		t.Errorf("Get(missing key) = %q, want empty", got)
	}
}

func TestFlashStore_MultipleKeys(t *testing.T) {
	fs := NewFlashStore()
	fs.Set("sess-1", "success", "Done")
	fs.Set("sess-1", "error", "Oops")

	if got := fs.Get("sess-1", "success"); got != "Done" {
		t.Errorf("Get(success) = %q, want %q", got, "Done")
	}
	if got := fs.Get("sess-1", "error"); got != "Oops" {
		t.Errorf("Get(error) = %q, want %q", got, "Oops")
	}
}

func TestSessionID(t *testing.T) {
	store, err := NewSessionStore(context.Background(), testKey, false)
	if err != nil {
		t.Fatal(err)
	}

	// No cookie returns empty
	req := httptest.NewRequest("GET", "/", nil)
	if got := store.SessionID(req); got != "" {
		t.Errorf("SessionID(no cookie) = %q, want empty", got)
	}

	// With valid session returns non-empty
	data := &SessionData{CustomerID: "CUST-001"}
	recorder := httptest.NewRecorder()
	store.Create(recorder, data)
	cookies := recorder.Result().Cookies()

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.AddCookie(cookies[0])
	if got := store.SessionID(req2); got == "" {
		t.Error("SessionID(valid session) should not be empty")
	}

	// Invalid cookie value returns empty
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.AddCookie(&http.Cookie{Name: "__session", Value: "garbage"})
	if got := store.SessionID(req3); got != "" {
		t.Errorf("SessionID(invalid cookie) = %q, want empty", got)
	}
}

