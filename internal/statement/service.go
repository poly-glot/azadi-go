package statement

import (
	"context"
	"fmt"
	"time"

	"azadi-go/internal/domain"
	"azadi-go/internal/model"
)

type AgreementGetter interface {
	GetAgreementsForCustomer(ctx context.Context, customerID string) ([]*model.Agreement, error)
}

type statementRepo interface {
	Save(ctx context.Context, s *model.StatementRequest) (*model.StatementRequest, error)
	FindByCustomerID(ctx context.Context, customerID string) ([]*model.StatementRequest, error)
}

type Service struct {
	repo         statementRepo
	auditService domain.AuditLogger
	agreementSvc AgreementGetter
}

func NewService(repo statementRepo, auditSvc domain.AuditLogger, agreementSvc AgreementGetter) *Service {
	return &Service{repo: repo, auditService: auditSvc, agreementSvc: agreementSvc}
}

func (s *Service) ResolveAgreementID(ctx context.Context, customerID string, agreementID *int64) *int64 {
	if agreementID != nil {
		return agreementID
	}
	agreements, err := s.agreementSvc.GetAgreementsForCustomer(ctx, customerID)
	if err != nil || len(agreements) == 0 {
		return nil
	}
	id := agreements[0].ID
	return &id
}

func (s *Service) RequestStatement(ctx context.Context, customerID string, agreementID int64, ipAddress, sessionID string) (*model.StatementRequest, error) {
	req := &model.StatementRequest{
		CustomerID:  customerID,
		AgreementID: agreementID,
		Status:      model.StatementStatusPending,
		RequestedAt: time.Now(),
	}
	saved, err := s.repo.Save(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("saving statement request: %w", err)
	}

	s.auditService.LogEvent(customerID, "STATEMENT_REQUESTED", ipAddress, sessionID, map[string]string{
		"agreementId": fmt.Sprintf("%d", agreementID),
	})
	return saved, nil
}

func (s *Service) GetStatementsForCustomer(ctx context.Context, customerID string) ([]*model.StatementRequest, error) {
	return s.repo.FindByCustomerID(ctx, customerID)
}
