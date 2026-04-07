package payment

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"azadi-go/internal/auth"
	"azadi-go/internal/model"

	"github.com/stripe/stripe-go/v82"
)

type mockPaymentSvc struct {
	pi  *stripe.PaymentIntent
	err error
}

func (m *mockPaymentSvc) InitiatePayment(_ context.Context, _ string, _, _ int64, _, _ string) (*stripe.PaymentIntent, error) {
	return m.pi, m.err
}

type mockWebhookProc struct {
	status int
	msg    string
}

func (m *mockWebhookProc) HandleEvent(_ []byte, _, _ string) (int, string) {
	return m.status, m.msg
}

type mockPayAgrLister struct {
	agreements []*model.Agreement
	err        error
}

func (m *mockPayAgrLister) GetAgreementsForCustomer(_ context.Context, _ string) ([]*model.Agreement, error) {
	return m.agreements, m.err
}

func testRender(w http.ResponseWriter, _ *http.Request, name string, _ map[string]any) {
	w.Header().Set("X-Template", name)
	w.WriteHeader(http.StatusOK)
}

func withSession(r *http.Request) *http.Request {
	ctx := auth.ContextWithSession(r.Context(), &auth.SessionData{CustomerID: "CUST-001", CustomerName: "Test User"}, "sess-123")
	return r.WithContext(ctx)
}

func TestPaymentPage_Success(t *testing.T) {
	agrSvc := &mockPayAgrLister{
		agreements: []*model.Agreement{{Base: model.Base{ID: 1}, AgreementNumber: "AGR-001"}},
	}
	h := NewHandler(&mockPaymentSvc{}, &mockWebhookProc{}, agrSvc, nil, "pk_test_xxx", testRender)

	req := withSession(httptest.NewRequest("GET", "/finance/make-a-payment", nil))
	w := httptest.NewRecorder()
	h.PaymentPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("X-Template"); got != "finance/make-a-payment.html" {
		t.Errorf("template = %q, want %q", got, "finance/make-a-payment.html")
	}
}

func TestPaymentPage_AgreementError(t *testing.T) {
	agrSvc := &mockPayAgrLister{err: errors.New("db error")}
	h := NewHandler(&mockPaymentSvc{}, &mockWebhookProc{}, agrSvc, nil, "pk_test_xxx", testRender)

	req := withSession(httptest.NewRequest("GET", "/finance/make-a-payment", nil))
	w := httptest.NewRecorder()
	h.PaymentPage(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestMakePayment_Success(t *testing.T) {
	svc := &mockPaymentSvc{
		pi: &stripe.PaymentIntent{ID: "pi_test", ClientSecret: "pi_test_secret_xxx"},
	}
	h := NewHandler(svc, &mockWebhookProc{}, &mockPayAgrLister{}, nil, "pk_test_xxx", testRender)

	body := `{"amountPence": 15000, "agreementId": 1}`
	req := withSession(httptest.NewRequest("POST", "/finance/make-a-payment", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.MakePayment(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "clientSecret") {
		t.Error("expected clientSecret in response")
	}
}

func TestMakePayment_BadJSON(t *testing.T) {
	h := NewHandler(&mockPaymentSvc{}, &mockWebhookProc{}, &mockPayAgrLister{}, nil, "pk_test_xxx", testRender)

	req := withSession(httptest.NewRequest("POST", "/finance/make-a-payment", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.MakePayment(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestMakePayment_MinimumAmount(t *testing.T) {
	h := NewHandler(&mockPaymentSvc{}, &mockWebhookProc{}, &mockPayAgrLister{}, nil, "pk_test_xxx", testRender)

	body := `{"amountPence": 50, "agreementId": 1}`
	req := withSession(httptest.NewRequest("POST", "/finance/make-a-payment", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.MakePayment(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestMakePayment_ServiceError(t *testing.T) {
	svc := &mockPaymentSvc{err: errors.New("stripe error")}
	h := NewHandler(svc, &mockWebhookProc{}, &mockPayAgrLister{}, nil, "pk_test_xxx", testRender)

	body := `{"amountPence": 15000, "agreementId": 1}`
	req := withSession(httptest.NewRequest("POST", "/finance/make-a-payment", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.MakePayment(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", w.Code)
	}
}

func TestHandleWebhook_Success(t *testing.T) {
	wh := &mockWebhookProc{status: 200, msg: "ok"}
	h := NewHandler(&mockPaymentSvc{}, wh, &mockPayAgrLister{}, nil, "pk_test_xxx", testRender)

	req := httptest.NewRequest("POST", "/api/stripe/webhook", strings.NewReader(`{}`))
	req.Header.Set("Stripe-Signature", "test_sig")
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHandleWebhook_BadSignature(t *testing.T) {
	wh := &mockWebhookProc{status: 400, msg: "Invalid signature"}
	h := NewHandler(&mockPaymentSvc{}, wh, &mockPayAgrLister{}, nil, "pk_test_xxx", testRender)

	req := httptest.NewRequest("POST", "/api/stripe/webhook", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)

	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
