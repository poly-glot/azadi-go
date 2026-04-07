package settlement

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

type mockSettlementSvc struct {
	figure      *model.SettlementFigure
	settlements []*model.SettlementFigure
	err         error
}

func (m *mockSettlementSvc) CalculateSettlement(_ context.Context, _ string, _ int64) (*model.SettlementFigure, error) {
	return m.figure, m.err
}

func (m *mockSettlementSvc) GetSettlementsForCustomer(_ context.Context, _ string) ([]*model.SettlementFigure, error) {
	return m.settlements, m.err
}

type mockAgreementLister struct {
	agreements []*model.Agreement
	err        error
}

func (m *mockAgreementLister) GetAgreementsForCustomer(_ context.Context, _ string) ([]*model.Agreement, error) {
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

func TestSettlementPage_Success(t *testing.T) {
	svc := &mockSettlementSvc{
		settlements: []*model.SettlementFigure{
			{Base: model.Base{ID: 1}, AmountPence: 1020000, CalculatedAt: time.Now(), ValidUntil: time.Now().AddDate(0, 0, 28)},
		},
	}
	agrSvc := &mockAgreementLister{
		agreements: []*model.Agreement{{Base: model.Base{ID: 1}, AgreementNumber: "AGR-001"}},
	}
	h := NewHandler(svc, agrSvc, testRender)

	req := withSession(httptest.NewRequest("GET", "/finance/settlement-figure", nil))
	w := httptest.NewRecorder()
	h.SettlementPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("X-Template"); got != "finance/settlement-figure.html" {
		t.Errorf("template = %q, want %q", got, "finance/settlement-figure.html")
	}
}

func TestSettlementPage_AgreementError(t *testing.T) {
	agrSvc := &mockAgreementLister{err: errors.New("db error")}
	h := NewHandler(&mockSettlementSvc{}, agrSvc, testRender)

	req := withSession(httptest.NewRequest("GET", "/finance/settlement-figure", nil))
	w := httptest.NewRecorder()
	h.SettlementPage(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestCalculateSettlement_Success(t *testing.T) {
	svc := &mockSettlementSvc{
		figure: &model.SettlementFigure{Base: model.Base{ID: 1}, AmountPence: 1020000, CalculatedAt: time.Now(), ValidUntil: time.Now().AddDate(0, 0, 28)},
	}
	agrSvc := &mockAgreementLister{
		agreements: []*model.Agreement{{Base: model.Base{ID: 1}}},
	}
	h := NewHandler(svc, agrSvc, testRender)

	form := url.Values{"agreementId": {"1"}}
	req := withSession(httptest.NewRequest("POST", "/finance/settlement-figure", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.CalculateSettlement(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestCalculateSettlement_BadAgreementID(t *testing.T) {
	h := NewHandler(&mockSettlementSvc{}, &mockAgreementLister{}, testRender)

	form := url.Values{"agreementId": {"abc"}}
	req := withSession(httptest.NewRequest("POST", "/finance/settlement-figure", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.CalculateSettlement(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestCalculateSettlement_SMSStub(t *testing.T) {
	h := NewHandler(&mockSettlementSvc{}, &mockAgreementLister{}, testRender)

	form := url.Values{"mobileNumber": {"07123456789"}, "agreementId": {"1"}}
	req := withSession(httptest.NewRequest("POST", "/finance/settlement-figure", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.CalculateSettlement(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
}

func TestCalculateSettlement_ServiceError(t *testing.T) {
	svc := &mockSettlementSvc{err: errors.New("calc error")}
	agrSvc := &mockAgreementLister{}
	h := NewHandler(svc, agrSvc, testRender)

	form := url.Values{"agreementId": {"1"}}
	req := withSession(httptest.NewRequest("POST", "/finance/settlement-figure", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.CalculateSettlement(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}
