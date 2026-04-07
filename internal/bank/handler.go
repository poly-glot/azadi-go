package bank

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"

	"azadi-go/internal/auth"
	"azadi-go/internal/httputil"
	"azadi-go/internal/model"
)

var (
	accountNumberPattern = regexp.MustCompile(`^\d{8}$`)
	sortCodePattern      = regexp.MustCompile(`^\d{2}-\d{2}-\d{2}$`)
)

type handlerService interface {
	GetBankDetails(ctx context.Context, customerID string) (*model.BankDetails, error)
	UpdateBankDetails(ctx context.Context, customerID, accountHolderName, accountNumber, sortCode, ipAddress, sessionID string) (*model.BankDetails, error)
}

type Handler struct {
	service  handlerService
	sessions *auth.SessionStore
	flashes  *auth.FlashStore
	render   httputil.RenderFunc
}

func NewHandler(service handlerService, sessions *auth.SessionStore, flashes *auth.FlashStore,
	render httputil.RenderFunc) *Handler {
	return &Handler{service: service, sessions: sessions, flashes: flashes, render: render}
}

func (h *Handler) BankDetailsPage(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	details, err := h.service.GetBankDetails(r.Context(), customer.CustomerID)
	if err != nil {
		slog.Error("failed to get bank details", "error", err)
	}
	data := map[string]any{}
	if details != nil {
		data["BankDetails"] = details
	}
	h.render(w, r, "finance/update-bank-details.html", data)
}

func (h *Handler) UpdateBankDetails(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	sessionID := auth.SessionIDFromContext(r.Context())

	if !httputil.RequireForm(w, r) {
		return
	}

	accountHolderName := r.FormValue("accountHolderName")
	accountNumber := r.FormValue("accountNumber")
	sortCode := r.FormValue("sortCode")

	// Validate
	var errors []string
	if accountHolderName == "" || len(accountHolderName) > 70 {
		errors = append(errors, "Account holder name is required (max 70 characters)")
	}
	if !accountNumberPattern.MatchString(accountNumber) {
		errors = append(errors, "Account number must be 8 digits")
	}
	if !sortCodePattern.MatchString(sortCode) {
		errors = append(errors, "Sort code must be in XX-XX-XX format")
	}

	if len(errors) > 0 {
		h.render(w, r, "finance/update-bank-details.html", map[string]any{
			"Errors": errors,
		})
		return
	}

	_, err := h.service.UpdateBankDetails(r.Context(), customer.CustomerID,
		accountHolderName, accountNumber, sortCode, r.RemoteAddr, sessionID)
	if err != nil {
		httputil.ServerError(w, "failed to update bank details", err)
		return
	}

	h.flashes.Set(sessionID, "success", "Bank details updated successfully.")
	http.Redirect(w, r, "/finance/update-bank-details", http.StatusFound)
}
