package document

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"azadi-go/internal/auth"
	"azadi-go/internal/model"
)

type mockDocService struct {
	docs []*model.Document
	doc  *model.Document
	err  error
}

func (m *mockDocService) GetDocumentsForCustomer(_ context.Context, _ string) ([]*model.Document, error) {
	return m.docs, m.err
}

func (m *mockDocService) GetDocument(_ context.Context, _ string, _ int64) (*model.Document, error) {
	return m.doc, m.err
}

func testRender(w http.ResponseWriter, _ *http.Request, name string, _ map[string]any) {
	w.Header().Set("X-Template", name)
	w.WriteHeader(http.StatusOK)
}

func withSession(r *http.Request) *http.Request {
	ctx := auth.ContextWithSession(r.Context(), &auth.SessionData{CustomerID: "CUST-001", CustomerName: "Test User"}, "sess-123")
	return r.WithContext(ctx)
}

func TestDocumentsPage_Success(t *testing.T) {
	svc := &mockDocService{
		docs: []*model.Document{{Base: model.Base{ID: 1}, Title: "Statement", FileName: "stmt.pdf"}},
	}
	h := NewHandler(svc, testRender)

	req := withSession(httptest.NewRequest("GET", "/my-documents", nil))
	w := httptest.NewRecorder()
	h.DocumentsPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("X-Template"); got != "my-documents.html" {
		t.Errorf("template = %q, want %q", got, "my-documents.html")
	}
}

func TestDocumentsPage_Error(t *testing.T) {
	svc := &mockDocService{err: errors.New("db error")}
	h := NewHandler(svc, testRender)

	req := withSession(httptest.NewRequest("GET", "/my-documents", nil))
	w := httptest.NewRecorder()
	h.DocumentsPage(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestDownload_Success(t *testing.T) {
	svc := &mockDocService{
		doc: &model.Document{Base: model.Base{ID: 1}, FileName: "statement.pdf", CustomerID: "CUST-001"},
	}
	h := NewHandler(svc, testRender)

	req := withSession(httptest.NewRequest("GET", "/documents/1/download", nil))
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.Download(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want 501", w.Code)
	}
	if got := w.Header().Get("Content-Disposition"); got == "" {
		t.Error("expected Content-Disposition header")
	}
}

func TestDownload_BadID(t *testing.T) {
	h := NewHandler(&mockDocService{}, testRender)

	req := withSession(httptest.NewRequest("GET", "/documents/abc/download", nil))
	req.SetPathValue("id", "abc")
	w := httptest.NewRecorder()
	h.Download(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestDownload_NotFound(t *testing.T) {
	svc := &mockDocService{err: ErrNotFound}
	h := NewHandler(svc, testRender)

	req := withSession(httptest.NewRequest("GET", "/documents/99/download", nil))
	req.SetPathValue("id", "99")
	w := httptest.NewRecorder()
	h.Download(w, req)

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
