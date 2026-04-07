package statement

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

type mockStatementSvc struct {
	statements  []*model.StatementRequest
	statement   *model.StatementRequest
	resolvedID  *int64
	err         error
}

func (m *mockStatementSvc) GetStatementsForCustomer(_ context.Context, _ string) ([]*model.StatementRequest, error) {
	return m.statements, m.err
}

func (m *mockStatementSvc) ResolveAgreementID(_ context.Context, _ string, _ *int64) *int64 {
	return m.resolvedID
}

func (m *mockStatementSvc) RequestStatement(_ context.Context, _ string, _ int64, _, _ string) (*model.StatementRequest, error) {
	return m.statement, m.err
}

type mockAgrLister struct {
	agreements []*model.Agreement
	err        error
}

func (m *mockAgrLister) GetAgreementsForCustomer(_ context.Context, _ string) ([]*model.Agreement, error) {
	return m.agreements, m.err
}

type mockCustGetter struct {
	customer *model.Customer
	err      error
}

func (m *mockCustGetter) GetCustomer(_ context.Context, _ string) (*model.Customer, error) {
	return m.customer, m.err
}

func testRender(w http.ResponseWriter, _ *http.Request, name string, _ map[string]any) {
	w.Header().Set("X-Template", name)
	w.WriteHeader(http.StatusOK)
}

func withSession(r *http.Request) *http.Request {
	ctx := auth.ContextWithSession(r.Context(), &auth.SessionData{CustomerID: "CUST-001", CustomerName: "Test User"}, "sess-123")
	return r.WithContext(ctx)
}

func TestStatementPage_Success(t *testing.T) {
	stmtSvc := &mockStatementSvc{
		statements: []*model.StatementRequest{{Base: model.Base{ID: 1}, Status: model.StatementStatusPending}},
	}
	agrSvc := &mockAgrLister{
		agreements: []*model.Agreement{{Base: model.Base{ID: 1}, AgreementNumber: "AGR-001"}},
	}
	custSvc := &mockCustGetter{
		customer: &model.Customer{Email: "test@test.com", AddressLine1: "123 Main St"},
	}
	flashes := auth.NewFlashStore()
	h := NewHandler(stmtSvc, agrSvc, custSvc, nil, flashes, testRender)

	req := withSession(httptest.NewRequest("GET", "/finance/request-a-statement", nil))
	w := httptest.NewRecorder()
	h.StatementPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("X-Template"); got != "finance/request-a-statement.html" {
		t.Errorf("template = %q, want %q", got, "finance/request-a-statement.html")
	}
}

func TestStatementPage_NilCustomer(t *testing.T) {
	stmtSvc := &mockStatementSvc{}
	agrSvc := &mockAgrLister{}
	custSvc := &mockCustGetter{err: errors.New("not found")}
	h := NewHandler(stmtSvc, agrSvc, custSvc, nil, auth.NewFlashStore(), testRender)

	req := withSession(httptest.NewRequest("GET", "/finance/request-a-statement", nil))
	w := httptest.NewRecorder()
	h.StatementPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestRequestStatement_Success(t *testing.T) {
	id := int64(1)
	stmtSvc := &mockStatementSvc{
		resolvedID: &id,
		statement:  &model.StatementRequest{Base: model.Base{ID: 1}, Status: model.StatementStatusPending, RequestedAt: time.Now()},
	}
	flashes := auth.NewFlashStore()
	h := NewHandler(stmtSvc, &mockAgrLister{}, &mockCustGetter{}, nil, flashes, testRender)

	form := url.Values{"agreementId": {"1"}}
	req := withSession(httptest.NewRequest("POST", "/finance/request-a-statement", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.RequestStatement(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
}

func TestRequestStatement_NoAgreementID(t *testing.T) {
	stmtSvc := &mockStatementSvc{resolvedID: nil}
	flashes := auth.NewFlashStore()
	h := NewHandler(stmtSvc, &mockAgrLister{}, &mockCustGetter{}, nil, flashes, testRender)

	form := url.Values{}
	req := withSession(httptest.NewRequest("POST", "/finance/request-a-statement", strings.NewReader(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.RequestStatement(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status = %d, want 302", w.Code)
	}
}
