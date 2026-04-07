package settlement

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"azadi-go/internal/auth"
	"azadi-go/internal/domain"
	"azadi-go/internal/httputil"
	"azadi-go/internal/model"
)

type settlementService interface {
	CalculateSettlement(ctx context.Context, customerID string, agreementID int64) (*model.SettlementFigure, error)
	GetSettlementsForCustomer(ctx context.Context, customerID string) ([]*model.SettlementFigure, error)
}

type Handler struct {
	service          settlementService
	agreementService domain.AgreementLister
	render           httputil.RenderFunc
}

func NewHandler(service settlementService, agreementSvc domain.AgreementLister,
	render httputil.RenderFunc) *Handler {
	return &Handler{service: service, agreementService: agreementSvc, render: render}
}

func (h *Handler) SettlementPage(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	agreements, err := h.agreementService.GetAgreementsForCustomer(r.Context(), customer.CustomerID)
	if err != nil {
		httputil.ServerError(w, "failed to get agreements", err)
		return
	}
	settlements, err := h.service.GetSettlementsForCustomer(r.Context(), customer.CustomerID)
	if err != nil {
		slog.Error("failed to get settlements", "error", err)
	}

	var responses []model.SettlementResponse
	for _, s := range settlements {
		responses = append(responses, model.NewSettlementResponse(s))
	}

	h.render(w, r, "finance/settlement-figure.html", map[string]any{
		"Agreements":  agreements,
		"Settlements": responses,
	})
}

func (h *Handler) CalculateSettlement(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())

	if !httputil.RequireForm(w, r) {
		return
	}

	// Check if SMS stub
	if r.FormValue("mobileNumber") != "" {
		http.Redirect(w, r, "/finance/settlement-figure", http.StatusFound)
		return
	}

	agreementID, err := strconv.ParseInt(r.FormValue("agreementId"), 10, 64)
	if err != nil {
		httputil.BadRequest(w, "Bad Request")
		return
	}

	agreements, _ := h.agreementService.GetAgreementsForCustomer(r.Context(), customer.CustomerID)
	figure, err := h.service.CalculateSettlement(r.Context(), customer.CustomerID, agreementID)
	if err != nil {
		httputil.ServerError(w, "failed to calculate settlement", err)
		return
	}

	resp := model.NewSettlementResponse(figure)
	h.render(w, r, "finance/settlement-figure.html", map[string]any{
		"Agreements": agreements,
		"Settlement": resp,
	})
}
