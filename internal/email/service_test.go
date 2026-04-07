package email

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"azadi-go/internal/model"
)

type mockEmailCustRepo struct {
	customer *model.Customer
	err      error
}

func (m *mockEmailCustRepo) FindByCustomerID(_ context.Context, _ string) (*model.Customer, error) {
	return m.customer, m.err
}

func TestNewService(t *testing.T) {
	svc := NewService("key", "from@test.com", &mockEmailCustRepo{})
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if svc.baseURL != "https://api.resend.com" {
		t.Errorf("baseURL = %q", svc.baseURL)
	}
}

func TestSendEmail(t *testing.T) {
	done := make(chan struct{})
	var gotBody string
	var gotAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(200)
		close(done)
	}))
	defer srv.Close()

	svc := &Service{
		apiKey:     "test-api-key",
		fromEmail:  "noreply@azadi.co.uk",
		baseURL:    srv.URL,
		httpClient: srv.Client(),
	}

	svc.SendEmail("user@test.com", "Test Subject", "<p>Hello</p>")

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for email send")
	}

	if !strings.Contains(gotBody, "user@test.com") {
		t.Errorf("body = %q, expected recipient", gotBody)
	}
	if gotAuth != "Bearer test-api-key" {
		t.Errorf("auth = %q", gotAuth)
	}
}

func TestResolveEmailAndSend(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	custRepo := &mockEmailCustRepo{
		customer: &model.Customer{Email: "found@test.com"},
	}
	svc := &Service{
		apiKey:       "key",
		fromEmail:    "noreply@azadi.co.uk",
		baseURL:      srv.URL,
		customerRepo: custRepo,
		httpClient:   srv.Client(),
	}
	// This calls resolveEmailAndSend internally
	svc.SendPaymentConfirmation("CUST-001", 15000)
	time.Sleep(100 * time.Millisecond)
}

func TestResolveEmailAndSend_NoCustomer(t *testing.T) {
	custRepo := &mockEmailCustRepo{customer: nil}
	svc := &Service{
		apiKey:       "key",
		fromEmail:    "noreply@azadi.co.uk",
		customerRepo: custRepo,
	}
	// Should not panic
	svc.resolveAndSend("CUST-UNKNOWN", "Subject", "<p>Body</p>")
}

func TestSendSettlementFigure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	svc := &Service{
		apiKey:       "key",
		fromEmail:    "noreply@azadi.co.uk",
		baseURL:      srv.URL,
		customerRepo: &mockEmailCustRepo{customer: &model.Customer{Email: "a@b.com"}},
		httpClient:   srv.Client(),
	}
	svc.SendSettlementFigure("CUST-001", 1020000, time.Now().AddDate(0, 0, 28))
	time.Sleep(100 * time.Millisecond)
}

func TestSendBankDetailsUpdated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	svc := &Service{
		apiKey:       "key",
		fromEmail:    "noreply@azadi.co.uk",
		baseURL:      srv.URL,
		customerRepo: &mockEmailCustRepo{customer: &model.Customer{Email: "a@b.com"}},
		httpClient:   srv.Client(),
	}
	svc.SendBankDetailsUpdated("CUST-001")
	time.Sleep(100 * time.Millisecond)
}

func TestSendPaymentDateChanged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	svc := &Service{
		apiKey:       "key",
		fromEmail:    "noreply@azadi.co.uk",
		baseURL:      srv.URL,
		customerRepo: &mockEmailCustRepo{customer: &model.Customer{Email: "a@b.com"}},
		httpClient:   srv.Client(),
	}
	svc.SendPaymentDateChanged("CUST-001", "15th")
	time.Sleep(100 * time.Millisecond)
}

func TestSendLoginAlert(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	svc := &Service{
		apiKey:       "key",
		fromEmail:    "noreply@azadi.co.uk",
		baseURL:      srv.URL,
		customerRepo: &mockEmailCustRepo{customer: &model.Customer{Email: "a@b.com"}},
		httpClient:   srv.Client(),
	}
	svc.SendLoginAlert("CUST-001", "192.168.1.1")
	time.Sleep(100 * time.Millisecond)
}

func TestPaymentConfirmationHTML(t *testing.T) {
	html := PaymentConfirmationHTML(15000)
	if !strings.Contains(html, "£150.00") {
		t.Errorf("expected formatted amount, got %s", html)
	}
	if !strings.Contains(html, "Payment Confirmed") {
		t.Error("expected title in HTML")
	}
}

func TestSettlementFigureHTML(t *testing.T) {
	validUntil := time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC)
	html := SettlementFigureHTML(1020000, validUntil)
	if !strings.Contains(html, "£10,200.00") {
		t.Errorf("expected formatted amount, got %s", html)
	}
	if !strings.Contains(html, "15 July 2024") {
		t.Errorf("expected formatted date, got %s", html)
	}
}

func TestBankDetailsUpdatedHTML(t *testing.T) {
	html := BankDetailsUpdatedHTML()
	if !strings.Contains(html, "Bank Details Updated") {
		t.Error("expected title in HTML")
	}
	if !strings.Contains(html, "successfully updated") {
		t.Error("expected confirmation text")
	}
}

func TestPaymentDateChangedHTML(t *testing.T) {
	html := PaymentDateChangedHTML("15th")
	if !strings.Contains(html, "15th") {
		t.Error("expected new date in HTML")
	}
	if !strings.Contains(html, "Payment Date Changed") {
		t.Error("expected title")
	}
}

func TestLoginAlertHTML(t *testing.T) {
	html := LoginAlertHTML("192.168.1.1")
	if !strings.Contains(html, "192.168.1.1") {
		t.Error("expected IP address in HTML")
	}
	if !strings.Contains(html, "Login Alert") {
		t.Error("expected title")
	}
}

func TestPaymentDateChangedHTML_EscapesInput(t *testing.T) {
	html := PaymentDateChangedHTML("<script>alert('xss')</script>")
	if strings.Contains(html, "<script>") {
		t.Error("expected HTML escaping of input")
	}
}

func TestLoginAlertHTML_EscapesInput(t *testing.T) {
	html := LoginAlertHTML("<img onerror=alert(1)>")
	if strings.Contains(html, "<img") {
		t.Error("expected HTML escaping of input")
	}
}

func TestWrapLayout(t *testing.T) {
	html := wrapLayout("Test Title", "<p>Body</p>")
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE")
	}
	if !strings.Contains(html, "Azadi Finance") {
		t.Error("expected brand name")
	}
	if !strings.Contains(html, "<p>Body</p>") {
		t.Error("expected body content")
	}
}
