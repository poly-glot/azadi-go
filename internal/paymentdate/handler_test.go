package paymentdate

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"azadi-go/internal/auth"
	"azadi-go/internal/model"
)

type mockPaymentDateSvc struct {
	err error
}

func (m *mockPaymentDateSvc) ChangePaymentDate(_ context.Context, _ string, _ int64, _ int, _, _ string) error {
	return m.err
}

type mockAgrLister struct {
	agreements []*model.Agreement
	err        error
}

func (m *mockAgrLister) GetAgreementsForCustomer(_ context.Context, _ string) ([]*model.Agreement, error) {
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

func TestChangeDatePage_Success(t *testing.T) {
	agrSvc := &mockAgrLister{
		agreements: []*model.Agreement{
			{Base: model.Base{ID: 1}, AgreementNumber: "AGR-001", NextPaymentDate: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)},
		},
	}
	flashes := auth.NewFlashStore()
	h := NewHandler(&mockPaymentDateSvc{}, agrSvc, nil, flashes, testRender)

	req := withSession(httptest.NewRequest("GET", "/finance/change-payment-date", nil))
	w := httptest.NewRecorder()
	h.ChangeDatePage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("X-Template"); got != "finance/change-payment-date.html" {
		t.Errorf("template = %q, want %q", got, "finance/change-payment-date.html")
	}
}

func TestChangeDatePage_NoAgreements(t *testing.T) {
	agrSvc := &mockAgrLister{agreements: []*model.Agreement{}}
	h := NewHandler(&mockPaymentDateSvc{}, agrSvc, nil, auth.NewFlashStore(), testRender)

	req := withSession(httptest.NewRequest("GET", "/finance/change-payment-date", nil))
	w := httptest.NewRecorder()
	h.ChangeDatePage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestChangeDatePage_AgreementError(t *testing.T) {
	agrSvc := &mockAgrLister{err: errors.New("db error")}
	h := NewHandler(&mockPaymentDateSvc{}, agrSvc, nil, auth.NewFlashStore(), testRender)

	req := withSession(httptest.NewRequest("GET", "/finance/change-payment-date", nil))
	w := httptest.NewRecorder()
	h.ChangeDatePage(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestChangeDate_Success(t *testing.T) {
	flashes := auth.NewFlashStore()
	h := NewHandler(&mockPaymentDateSvc{}, &mockAgrLister{}, nil, flashes, testRender)

	form := url.Values{"newPaymentDate": {"15"}, "agreementId": {"1"}}
	req := withSession(httptest.NewRequest("POST", "/finance/change-payment-date", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ChangeDate(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
}

func TestChangeDate_BadDay(t *testing.T) {
	h := NewHandler(&mockPaymentDateSvc{}, &mockAgrLister{}, nil, auth.NewFlashStore(), testRender)

	form := url.Values{"newPaymentDate": {"abc"}, "agreementId": {"1"}}
	req := withSession(httptest.NewRequest("POST", "/finance/change-payment-date", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ChangeDate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestChangeDate_BadAgreementID(t *testing.T) {
	h := NewHandler(&mockPaymentDateSvc{}, &mockAgrLister{}, nil, auth.NewFlashStore(), testRender)

	form := url.Values{"newPaymentDate": {"15"}, "agreementId": {"abc"}}
	req := withSession(httptest.NewRequest("POST", "/finance/change-payment-date", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ChangeDate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestChangeDate_ServiceError(t *testing.T) {
	svc := &mockPaymentDateSvc{err: ErrAlreadyChanged}
	flashes := auth.NewFlashStore()
	h := NewHandler(svc, &mockAgrLister{}, nil, flashes, testRender)

	form := url.Values{"newPaymentDate": {"15"}, "agreementId": {"1"}}
	req := withSession(httptest.NewRequest("POST", "/finance/change-payment-date", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ChangeDate(w, req)

	// Error gets stored as flash message, still redirects
	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
}
