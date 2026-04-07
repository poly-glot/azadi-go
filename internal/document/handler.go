package document

import (
	"context"
	"net/http"

	"azadi-go/internal/auth"
	"azadi-go/internal/httputil"
	"azadi-go/internal/model"
)

type handlerService interface {
	GetDocumentsForCustomer(ctx context.Context, customerID string) ([]*model.Document, error)
	GetDocument(ctx context.Context, customerID string, documentID int64) (*model.Document, error)
}

type Handler struct {
	service handlerService
	render  httputil.RenderFunc
}

func NewHandler(service handlerService, render httputil.RenderFunc) *Handler {
	return &Handler{service: service, render: render}
}

func (h *Handler) DocumentsPage(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	docs, err := h.service.GetDocumentsForCustomer(r.Context(), customer.CustomerID)
	if err != nil {
		httputil.ServerError(w, "failed to get documents", err)
		return
	}
	h.render(w, r, "my-documents.html", map[string]any{
		"Documents": docs,
	})
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	id, err := httputil.PathInt64(r, "id")
	if err != nil {
		httputil.BadRequest(w, "Bad Request")
		return
	}

	doc, err := h.service.GetDocument(r.Context(), customer.CustomerID, id)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+doc.FileName+"\"")
	w.Header().Set("Content-Type", "application/pdf")
	w.WriteHeader(http.StatusNotImplemented)
}
