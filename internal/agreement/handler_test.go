package agreement

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"azadi-go/internal/auth"
	"azadi-go/internal/model"
)

type mockService struct {
	agreements []*model.Agreement
	agreement  *model.Agreement
	err        error
}

func (m *mockService) GetAgreementsForCustomer(_ context.Context, _ string) ([]*model.Agreement, error) {
	return m.agreements, m.err
}

func (m *mockService) GetAgreement(_ context.Context, _ string, _ int64) (*model.Agreement, error) {
	return m.agreement, m.err
}

func testRender(w http.ResponseWriter, _ *http.Request, name string, _ map[string]any) {
	w.Header().Set("X-Template", name)
	w.WriteHeader(http.StatusOK)
}

func withSession(r *http.Request) *http.Request {
	ctx := auth.ContextWithSession(r.Context(), &auth.SessionData{CustomerID: "CUST-001", CustomerName: "Test User"}, "sess-123")
	return r.WithContext(ctx)
}

func TestMyAccount_Success(t *testing.T) {
	svc := &mockService{
		agreements: []*model.Agreement{
			{Base: model.Base{ID: 1}, AgreementNumber: "AGR-001", BalancePence: 100000},
		},
	}
	h := NewHandler(svc, testRender)

	req := withSession(httptest.NewRequest("GET", "/my-account", nil))
	w := httptest.NewRecorder()
	h.MyAccount(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("X-Template"); got != "my-account.html" {
		t.Errorf("template = %q, want %q", got, "my-account.html")
	}
}

func TestMyAccount_ServiceError(t *testing.T) {
	svc := &mockService{err: errors.New("db error")}
	h := NewHandler(svc, testRender)

	req := withSession(httptest.NewRequest("GET", "/my-account", nil))
	w := httptest.NewRecorder()
	h.MyAccount(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestAgreementDetail_Success(t *testing.T) {
	svc := &mockService{
		agreement: &model.Agreement{Base: model.Base{ID: 42}, AgreementNumber: "AGR-001", CustomerID: "CUST-001"},
	}
	h := NewHandler(svc, testRender)

	req := withSession(httptest.NewRequest("GET", "/agreements/42", nil))
	req.SetPathValue("id", "42")
	w := httptest.NewRecorder()
	h.AgreementDetail(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("X-Template"); got != "agreement-detail.html" {
		t.Errorf("template = %q, want %q", got, "agreement-detail.html")
	}
}

func TestAgreementDetail_BadID(t *testing.T) {
	h := NewHandler(&mockService{}, testRender)

	req := withSession(httptest.NewRequest("GET", "/agreements/abc", nil))
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()
	h.AgreementDetail(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestAgreementDetail_NotFound(t *testing.T) {
	svc := &mockService{err: ErrNotFound}
	h := NewHandler(svc, testRender)

	req := withSession(httptest.NewRequest("GET", "/agreements/99", nil))
	req.SetPathValue("id", "99")
	w := httptest.NewRecorder()
	h.AgreementDetail(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestAgreementDetail_AccessDenied(t *testing.T) {
	svc := &mockService{err: ErrAccessDenied}
	h := NewHandler(svc, testRender)

	req := withSession(httptest.NewRequest("GET", "/agreements/99", nil))
	req.SetPathValue("id", "99")
	w := httptest.NewRecorder()
	h.AgreementDetail(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestNewService(t *testing.T) {
	svc := NewService(nil)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}
