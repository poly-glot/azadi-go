package paymentdate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"azadi-go/internal/domain"
	"azadi-go/internal/model"
)

var (
	ErrAlreadyChanged = errors.New("payment date has already been changed for this agreement")
	ErrInvalidDay     = errors.New("payment day must be between 1 and 28")
)

type AgreementGetter interface {
	GetAgreement(ctx context.Context, customerID string, agreementID int64) (*model.Agreement, error)
}

type AgreementSaver interface {
	Save(ctx context.Context, a *model.Agreement) (*model.Agreement, error)
}

type Service struct {
	agreementSvc  AgreementGetter
	agreementRepo AgreementSaver
	auditService  domain.AuditLogger
}

func NewService(agreementSvc AgreementGetter, agreementRepo AgreementSaver, auditSvc domain.AuditLogger) *Service {
	return &Service{
		agreementSvc:  agreementSvc,
		agreementRepo: agreementRepo,
		auditService:  auditSvc,
	}
}

func (s *Service) ChangePaymentDate(ctx context.Context, customerID string, agreementID int64, newDay int, ipAddress, sessionID string) error {
	agreement, err := s.agreementSvc.GetAgreement(ctx, customerID, agreementID)
	if err != nil {
		return fmt.Errorf("getting agreement: %w", err)
	}

	if agreement.PaymentDateChanged {
		return ErrAlreadyChanged
	}

	if newDay < 1 || newDay > 28 {
		return ErrInvalidDay
	}

	if !agreement.NextPaymentDate.IsZero() {
		y, m, _ := agreement.NextPaymentDate.Date()
		agreement.NextPaymentDate = time.Date(y, m, newDay, 0, 0, 0, 0, agreement.NextPaymentDate.Location())
	}
	agreement.PaymentDateChanged = true

	if _, err := s.agreementRepo.Save(ctx, agreement); err != nil {
		return fmt.Errorf("saving agreement: %w", err)
	}

	s.auditService.LogEvent(customerID, "PAYMENT_DATE_CHANGED", ipAddress, sessionID, map[string]string{
		"agreementId": fmt.Sprintf("%d", agreementID),
		"newDay":      fmt.Sprintf("%d", newDay),
	})
	return nil
}

func GetCurrentPaymentDay(agreement *model.Agreement) int {
	if !agreement.NextPaymentDate.IsZero() {
		return agreement.NextPaymentDate.Day()
	}
	return 1
}
