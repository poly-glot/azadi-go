package payment

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"azadi-go/internal/auth"
	"azadi-go/internal/domain"
	"azadi-go/internal/httputil"

	"github.com/stripe/stripe-go/v82"
)

type paymentService interface {
	InitiatePayment(ctx context.Context, customerID string, agreementID, amountPence int64, ipAddress, sessionID string) (*stripe.PaymentIntent, error)
}

type webhookProcessor interface {
	HandleEvent(payload []byte, sigHeader, ipAddress string) (int, string)
}

type Handler struct {
	service              paymentService
	webhookHandler       webhookProcessor
	agreementService     domain.AgreementLister
	sessions             *auth.SessionStore
	stripePublishableKey string
	render               httputil.RenderFunc
}

func NewHandler(service paymentService, webhookHandler webhookProcessor, agreementSvc domain.AgreementLister,
	sessions *auth.SessionStore, stripePublishableKey string,
	render httputil.RenderFunc) *Handler {
	return &Handler{
		service:              service,
		webhookHandler:       webhookHandler,
		agreementService:     agreementSvc,
		sessions:             sessions,
		stripePublishableKey: stripePublishableKey,
		render:               render,
	}
}

func (h *Handler) PaymentPage(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	agreements, err := h.agreementService.GetAgreementsForCustomer(r.Context(), customer.CustomerID)
	if err != nil {
		httputil.ServerError(w, "failed to get agreements", err)
		return
	}
	h.render(w, r, "finance/make-a-payment.html", map[string]any{
		"Agreements":           agreements,
		"StripePublishableKey": h.stripePublishableKey,
	})
}

func (h *Handler) MakePayment(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	sessionID := auth.SessionIDFromContext(r.Context())

	var req struct {
		AmountPence int64 `json:"amountPence"`
		AgreementID int64 `json:"agreementId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "Bad Request")
		return
	}

	if req.AmountPence < 100 {
		httputil.BadRequest(w, "Minimum payment is £1.00")
		return
	}

	pi, err := h.service.InitiatePayment(r.Context(), customer.CustomerID,
		req.AgreementID, req.AmountPence, r.RemoteAddr, sessionID)
	if err != nil {
		slog.Error("failed to initiate payment", "error", err)
		http.Error(w, "Payment Service Unavailable", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"clientSecret": pi.ClientSecret,
	}); err != nil {
		slog.Error("encoding payment response", "error", err)
	}
}

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	const maxWebhookBody = 1 << 16 // 64KB
	body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBody))
	if err != nil {
		httputil.BadRequest(w, "Bad Request")
		return
	}
	sigHeader := r.Header.Get("Stripe-Signature")
	status, msg := h.webhookHandler.HandleEvent(body, sigHeader, r.RemoteAddr)
	w.WriteHeader(status)
	if _, err := io.WriteString(w, msg); err != nil {
		slog.Error("writing webhook response", "error", err)
	}
}
