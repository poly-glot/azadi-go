package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAuth_Unauthenticated(t *testing.T) {
	store, err := NewSessionStore(context.Background(), "dev-only-key-change-in-prod-32ch", false)
	if err != nil {
		t.Fatal(err)
	}
	handler := RequireAuth(store, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/my-account", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("got %d, want 302 redirect", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/login" {
		t.Errorf("redirect to %q, want /login", loc)
	}
}

func TestRequireAuth_Authenticated(t *testing.T) {
	store, err := NewSessionStore(context.Background(), "dev-only-key-change-in-prod-32ch", false)
	if err != nil {
		t.Fatal(err)
	}

	// Create session
	data := &SessionData{CustomerID: "CUST-001", CustomerName: "James"}
	rec := httptest.NewRecorder()
	if err := store.Create(rec, data); err != nil {
		t.Fatal(err)
	}
	cookie := rec.Result().Cookies()[0]

	var gotCustomer *SessionData
	handler := RequireAuth(store, func(w http.ResponseWriter, r *http.Request) {
		gotCustomer = CustomerFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/my-account", nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
	if gotCustomer == nil {
		t.Fatal("customer not found in context")
	}
	if gotCustomer.CustomerID != "CUST-001" {
		t.Errorf("CustomerID = %q, want CUST-001", gotCustomer.CustomerID)
	}
}

func TestCustomerFromContext_Nil(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if c := CustomerFromContext(req.Context()); c != nil {
		t.Error("should be nil for unauthenticated request")
	}
}

func TestSessionIDFromContext_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	if s := SessionIDFromContext(req.Context()); s != "" {
		t.Errorf("expected empty string, got %q", s)
	}
}
