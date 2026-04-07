package bank

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

type mockBankService struct {
	details  *model.BankDetails
	err      error
	updateOK bool
}

func (m *mockBankService) GetBankDetails(_ context.Context, _ string) (*model.BankDetails, error) {
	return m.details, m.err
}

func (m *mockBankService) UpdateBankDetails(_ context.Context, _, _, _, _, _, _ string) (*model.BankDetails, error) {
	if m.updateOK {
		return &model.BankDetails{}, nil
	}
	return nil, m.err
}

func testRender(w http.ResponseWriter, _ *http.Request, name string, _ map[string]any) {
	w.Header().Set("X-Template", name)
	w.WriteHeader(http.StatusOK)
}

func withSession(r *http.Request) *http.Request {
	ctx := auth.ContextWithSession(r.Context(), &auth.SessionData{CustomerID: "CUST-001", CustomerName: "Test User"}, "sess-123")
	return r.WithContext(ctx)
}

func TestBankDetailsPage_Success(t *testing.T) {
	svc := &mockBankService{
		details: &model.BankDetails{LastFourAccount: "1234", LastTwoSortCode: "56"},
	}
	h := NewHandler(svc, nil, auth.NewFlashStore(), testRender)

	req := withSession(httptest.NewRequest("GET", "/finance/update-bank-details", nil))
	w := httptest.NewRecorder()
	h.BankDetailsPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("X-Template"); got != "finance/update-bank-details.html" {
		t.Errorf("template = %q, want %q", got, "finance/update-bank-details.html")
	}
}

func TestBankDetailsPage_NoDetails(t *testing.T) {
	svc := &mockBankService{err: errors.New("not found")}
	h := NewHandler(svc, nil, auth.NewFlashStore(), testRender)

	req := withSession(httptest.NewRequest("GET", "/finance/update-bank-details", nil))
	w := httptest.NewRecorder()
	h.BankDetailsPage(w, req)

	// Handler logs error but still renders the page
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestUpdateBankDetails_Success(t *testing.T) {
	svc := &mockBankService{updateOK: true}
	flashes := auth.NewFlashStore()
	h := NewHandler(svc, nil, flashes, testRender)

	form := url.Values{
		"accountHolderName": {"John Smith"},
		"accountNumber":     {"12345678"},
		"sortCode":          {"12-34-56"},
	}
	req := withSession(httptest.NewRequest("POST", "/finance/update-bank-details", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.UpdateBankDetails(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
}

func TestUpdateBankDetails_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		holder      string
		account     string
		sortCode    string
	}{
		{"empty holder", "", "12345678", "12-34-56"},
		{"bad account", "John", "1234", "12-34-56"},
		{"bad sort code", "John", "12345678", "123456"},
		{"all bad", "", "abc", "xyz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&mockBankService{}, nil, auth.NewFlashStore(), testRender)

			form := url.Values{
				"accountHolderName": {tt.holder},
				"accountNumber":     {tt.account},
				"sortCode":          {tt.sortCode},
			}
			req := withSession(httptest.NewRequest("POST", "/finance/update-bank-details", strings.NewReader(form.Encode())))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			h.UpdateBankDetails(w, req)

			// Validation errors re-render the page
			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want 200 (re-render with errors)", w.Code)
			}
		})
	}
}

func TestUpdateBankDetails_ServiceError(t *testing.T) {
	svc := &mockBankService{err: errors.New("encrypt error")}
	h := NewHandler(svc, nil, auth.NewFlashStore(), testRender)

	form := url.Values{
		"accountHolderName": {"John Smith"},
		"accountNumber":     {"12345678"},
		"sortCode":          {"12-34-56"},
	}
	req := withSession(httptest.NewRequest("POST", "/finance/update-bank-details", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.UpdateBankDetails(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestAccountNumberPattern(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12345678", true},
		{"00000000", true},
		{"1234567", false},
		{"123456789", false},
		{"1234abcd", false},
		{"", false},
	}
	for _, tt := range tests {
		got := accountNumberPattern.MatchString(tt.input)
		if got != tt.valid {
			t.Errorf("accountNumber %q: got %v, want %v", tt.input, got, tt.valid)
		}
	}
}

func TestSortCodePattern(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12-34-56", true},
		{"00-00-00", true},
		{"123456", false},
		{"12-34-5", false},
		{"ab-cd-ef", false},
		{"", false},
	}
	for _, tt := range tests {
		got := sortCodePattern.MatchString(tt.input)
		if got != tt.valid {
			t.Errorf("sortCode %q: got %v, want %v", tt.input, got, tt.valid)
		}
	}
}
