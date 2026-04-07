package payment

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"azadi-go/internal/domain"
	"azadi-go/internal/model"

	"github.com/stripe/stripe-go/v82"
)

type AgreementGetter interface {
	GetAgreement(ctx context.Context, customerID string, agreementID int64) (*model.Agreement, error)
}

type paymentStripe interface {
	CreatePaymentIntent(amountPence int64, agreementNumber, customerEmail string) (*stripe.PaymentIntent, error)
}

type paymentRepo interface {
	Save(ctx context.Context, p *model.PaymentRecord) (*model.PaymentRecord, error)
}

type customerFinder interface {
	FindByCustomerID(ctx context.Context, customerID string) (*model.Customer, error)
}

type Service struct {
	stripeService paymentStripe
	repo          paymentRepo
	auditService  domain.AuditLogger
	customerRepo  customerFinder
	agreementSvc  AgreementGetter
}

func NewService(stripeSvc paymentStripe, repo paymentRepo, auditSvc domain.AuditLogger,
	customerRepo customerFinder, agreementSvc AgreementGetter) *Service {
	return &Service{
		stripeService: stripeSvc,
		repo:          repo,
		auditService:  auditSvc,
		customerRepo:  customerRepo,
		agreementSvc:  agreementSvc,
	}
}

func (s *Service) InitiatePayment(ctx context.Context, customerID string, agreementID, amountPence int64, ipAddress, sessionID string) (*stripe.PaymentIntent, error) {
	agreement, err := s.agreementSvc.GetAgreement(ctx, customerID, agreementID)
	if err != nil {
		return nil, fmt.Errorf("getting agreement: %w", err)
	}

	customerEmail := ""
	customer, err := s.customerRepo.FindByCustomerID(ctx, customerID)
	if err == nil && customer != nil && customer.Email != "" {
		customerEmail = customer.Email
	} else {
		slog.Warn("no email found for customer", "customerID", customerID)
	}

	pi, err := s.stripeService.CreatePaymentIntent(amountPence, agreement.AgreementNumber, customerEmail)
	if err != nil {
		return nil, fmt.Errorf("creating payment intent: %w", err)
	}

	record := &model.PaymentRecord{
		AgreementID:           agreementID,
		CustomerID:            customerID,
		AmountPence:           amountPence,
		StripePaymentIntentID: pi.ID,
		Status:                model.PaymentStatusPending,
		CreatedAt:             time.Now(),
	}
	if _, err := s.repo.Save(ctx, record); err != nil {
		return nil, fmt.Errorf("saving payment record: %w", err)
	}

	s.auditService.LogEvent(customerID, "PAYMENT_INITIATED", ipAddress, sessionID, map[string]string{
		"amount":          fmt.Sprintf("%d", amountPence),
		"agreementNumber": agreement.AgreementNumber,
	})

	return pi, nil
}
