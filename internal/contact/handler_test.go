package contact

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"azadi-go/internal/auth"
	"azadi-go/internal/model"
)

type mockContactService struct {
	customer *model.Customer
	err      error
	updateOK bool
}

func (m *mockContactService) GetCustomer(_ context.Context, _ string) (*model.Customer, error) {
	return m.customer, m.err
}

func (m *mockContactService) UpdateContactDetails(_ context.Context, _, _, _, _, _, _, _, _, _ string, _ string) error {
	if m.updateOK {
		return nil
	}
	return m.err
}

func testRender(w http.ResponseWriter, _ *http.Request, name string, _ map[string]any) {
	w.Header().Set("X-Template", name)
	w.WriteHeader(http.StatusOK)
}

func withSession(r *http.Request) *http.Request {
	ctx := auth.ContextWithSession(r.Context(), &auth.SessionData{CustomerID: "CUST-001", CustomerName: "Test User"}, "sess-123")
	return r.WithContext(ctx)
}

func TestContactPage_Success(t *testing.T) {
	svc := &mockContactService{
		customer: &model.Customer{CustomerID: "CUST-001", Email: "test@test.com"},
	}
	h := NewHandler(svc, nil, auth.NewFlashStore(), testRender)

	req := withSession(httptest.NewRequest("GET", "/my-contact-details", nil))
	w := httptest.NewRecorder()
	h.ContactPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("X-Template"); got != "my-contact-details.html" {
		t.Errorf("template = %q, want %q", got, "my-contact-details.html")
	}
}

func TestContactPage_Error(t *testing.T) {
	svc := &mockContactService{err: errors.New("db error")}
	h := NewHandler(svc, nil, auth.NewFlashStore(), testRender)

	req := withSession(httptest.NewRequest("GET", "/my-contact-details", nil))
	w := httptest.NewRecorder()
	h.ContactPage(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestUpdateContact_Success(t *testing.T) {
	svc := &mockContactService{
		customer: &model.Customer{CustomerID: "CUST-001", Email: "old@test.com", Phone: "01onal"},
		updateOK: true,
	}
	flashes := auth.NewFlashStore()
	h := NewHandler(svc, nil, flashes, testRender)

	form := url.Values{
		"email":       {"new@test.com"},
		"homePhone":   {"01onal"},
		"mobilePhone": {"07111222333"},
	}
	req := withSession(httptest.NewRequest("POST", "/my-contact-details", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.UpdateContact(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
}

func TestUpdateContact_GetCustomerError(t *testing.T) {
	svc := &mockContactService{err: errors.New("not found")}
	h := NewHandler(svc, nil, auth.NewFlashStore(), testRender)

	form := url.Values{"email": {"new@test.com"}}
	req := withSession(httptest.NewRequest("POST", "/my-contact-details", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.UpdateContact(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestNonEmpty(t *testing.T) {
	tests := []struct {
		value, fallback, want string
	}{
		{"new", "old", "new"},
		{"", "old", "old"},
		{"", "", ""},
		{"x", "", "x"},
	}
	for _, tt := range tests {
		got := nonEmpty(tt.value, tt.fallback)
		if got != tt.want {
			t.Errorf("nonEmpty(%q, %q) = %q, want %q", tt.value, tt.fallback, got, tt.want)
		}
	}
}
