package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"azadi-go/internal/domain"
	"azadi-go/internal/model"

	"github.com/stripe/stripe-go/v82"
)

type webhookRepo interface {
	FindByWebhookEventID(ctx context.Context, eventID string) (*model.PaymentRecord, error)
	FindByStripePaymentIntentID(ctx context.Context, intentID string) (*model.PaymentRecord, error)
	Save(ctx context.Context, p *model.PaymentRecord) (*model.PaymentRecord, error)
}

type webhookStripe interface {
	ConstructWebhookEvent(payload []byte, sigHeader string) (stripe.Event, error)
}

type WebhookHandler struct {
	repo          webhookRepo
	auditService  domain.AuditLogger
	stripeService webhookStripe
	emailService  interface {
		SendPaymentConfirmation(customerID string, amountPence int64)
	}
}

func NewWebhookHandler(repo webhookRepo, auditSvc domain.AuditLogger, stripeSvc webhookStripe,
	emailSvc interface{ SendPaymentConfirmation(customerID string, amountPence int64) }) *WebhookHandler {
	return &WebhookHandler{
		repo:          repo,
		auditService:  auditSvc,
		stripeService: stripeSvc,
		emailService:  emailSvc,
	}
}

func (h *WebhookHandler) HandleEvent(payload []byte, sigHeader, ipAddress string) (int, string) {
	event, err := h.stripeService.ConstructWebhookEvent(payload, sigHeader)
	if err != nil {
		slog.Warn("invalid webhook signature", "error", err)
		return 400, "Invalid signature"
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			slog.Error("unmarshalling payment intent from webhook", "error", err)
			return 400, "Bad payload"
		}
		h.handlePaymentSuccess(pi.ID, event.ID, ipAddress)
	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			slog.Error("unmarshalling payment intent from webhook", "error", err)
			return 400, "Bad payload"
		}
		h.handlePaymentFailure(pi.ID, event.ID, ipAddress)
	default:
		slog.Info("unhandled webhook event type", "type", event.Type)
	}

	return 200, "ok"
}

func (h *WebhookHandler) handlePaymentSuccess(paymentIntentID, webhookEventID, ipAddress string) {
	ctx := context.TODO()

	// Idempotency check
	existing, err := h.repo.FindByWebhookEventID(ctx, webhookEventID)
	if err != nil {
		slog.Error("idempotency check failed", "eventID", webhookEventID, "error", err)
		return
	}
	if existing != nil {
		slog.Info("duplicate webhook event ignored", "eventID", webhookEventID)
		return
	}

	record, err := h.repo.FindByStripePaymentIntentID(ctx, paymentIntentID)
	if err != nil || record == nil {
		slog.Warn("no PaymentRecord found for PaymentIntent", "piID", paymentIntentID)
		return
	}

	record.Status = model.PaymentStatusCompleted
	record.CompletedAt = time.Now()
	record.WebhookEventID = webhookEventID
	if _, err := h.repo.Save(ctx, record); err != nil {
		slog.Error("failed to save payment record", "error", err)
		return
	}

	h.auditService.LogEvent(record.CustomerID, "PAYMENT_COMPLETED", ipAddress, "", map[string]string{
		"amount":          fmt.Sprintf("%d", record.AmountPence),
		"paymentIntentId": paymentIntentID,
	})

	h.emailService.SendPaymentConfirmation(record.CustomerID, record.AmountPence)
}

func (h *WebhookHandler) handlePaymentFailure(paymentIntentID, webhookEventID, ipAddress string) {
	ctx := context.TODO()

	record, err := h.repo.FindByStripePaymentIntentID(ctx, paymentIntentID)
	if err != nil || record == nil {
		slog.Warn("no PaymentRecord found for PaymentIntent", "piID", paymentIntentID)
		return
	}

	record.Status = model.PaymentStatusFailed
	record.CompletedAt = time.Now()
	record.WebhookEventID = webhookEventID
	if _, err := h.repo.Save(ctx, record); err != nil {
		slog.Error("failed to save payment record", "error", err)
		return
	}

	h.auditService.LogEvent(record.CustomerID, "PAYMENT_FAILED", ipAddress, "", map[string]string{
		"paymentIntentId": paymentIntentID,
	})
}
