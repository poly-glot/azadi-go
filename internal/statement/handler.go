package statement

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

type statementHandlerService interface {
	GetStatementsForCustomer(ctx context.Context, customerID string) ([]*model.StatementRequest, error)
	ResolveAgreementID(ctx context.Context, customerID string, agreementID *int64) *int64
	RequestStatement(ctx context.Context, customerID string, agreementID int64, ipAddress, sessionID string) (*model.StatementRequest, error)
}

type customerGetter interface {
	GetCustomer(ctx context.Context, customerID string) (*model.Customer, error)
}

type Handler struct {
	service          statementHandlerService
	agreementService domain.AgreementLister
	contactService   customerGetter
	sessions         *auth.SessionStore
	flashes          *auth.FlashStore
	render           httputil.RenderFunc
}

func NewHandler(service statementHandlerService, agreementSvc domain.AgreementLister, contactSvc customerGetter,
	sessions *auth.SessionStore, flashes *auth.FlashStore,
	render httputil.RenderFunc) *Handler {
	return &Handler{
		service: service, agreementService: agreementSvc, contactService: contactSvc,
		sessions: sessions, flashes: flashes, render: render,
	}
}

func (h *Handler) StatementPage(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	agreements, err := h.agreementService.GetAgreementsForCustomer(r.Context(), customer.CustomerID)
	if err != nil {
		slog.Error("failed to get agreements", "error", err)
	}
	statements, err := h.service.GetStatementsForCustomer(r.Context(), customer.CustomerID)
	if err != nil {
		slog.Error("failed to get statements", "error", err)
	}
	cust, _ := h.contactService.GetCustomer(r.Context(), customer.CustomerID)

	data := map[string]any{
		"Agreements": agreements,
		"Statements": statements,
	}
	if cust != nil {
		data["Email"] = cust.Email
		data["Address"] = cust.AddressLine1
	}
	h.render(w, r, "finance/request-a-statement.html", data)
}

func (h *Handler) RequestStatement(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	sessionID := auth.SessionIDFromContext(r.Context())

	if !httputil.RequireForm(w, r) {
		return
	}

	var agreementIDPtr *int64
	if idStr := r.FormValue("agreementId"); idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			agreementIDPtr = &id
		}
	}

	resolved := h.service.ResolveAgreementID(r.Context(), customer.CustomerID, agreementIDPtr)
	if resolved != nil {
		if _, err := h.service.RequestStatement(r.Context(), customer.CustomerID, *resolved, r.RemoteAddr, sessionID); err != nil {
			slog.Error("failed to request statement", "error", err)
		}
	}

	h.flashes.Set(sessionID, "success", "Statement requested successfully. You will receive it by email.")
	http.Redirect(w, r, "/finance/request-a-statement", http.StatusFound)
}
