package agreement

import (
	"context"
	"net/http"

	"azadi-go/internal/auth"
	"azadi-go/internal/httputil"
	"azadi-go/internal/model"
)

type handlerService interface {
	GetAgreementsForCustomer(ctx context.Context, customerID string) ([]*model.Agreement, error)
	GetAgreement(ctx context.Context, customerID string, agreementID int64) (*model.Agreement, error)
}

type Handler struct {
	service handlerService
	render  httputil.RenderFunc
}

func NewHandler(service handlerService, render httputil.RenderFunc) *Handler {
	return &Handler{service: service, render: render}
}

func (h *Handler) MyAccount(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	agreements, err := h.service.GetAgreementsForCustomer(r.Context(), customer.CustomerID)
	if err != nil {
		httputil.ServerError(w, "failed to get agreements", err)
		return
	}

	var responses []model.AgreementResponse
	for _, a := range agreements {
		responses = append(responses, model.NewAgreementResponse(a))
	}

	h.render(w, r, "my-account.html", map[string]any{
		"Agreements": responses,
	})
}

func (h *Handler) AgreementDetail(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	id, err := httputil.PathInt64(r, "id")
	if err != nil {
		httputil.BadRequest(w, "Bad Request")
		return
	}

	agreement, err := h.service.GetAgreement(r.Context(), customer.CustomerID, id)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	resp := model.NewAgreementResponse(agreement)
	h.render(w, r, "agreement-detail.html", map[string]any{
		"Agreement": resp,
	})
}
