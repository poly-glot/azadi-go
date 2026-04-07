package settlement

import (
	"context"
	"fmt"
	"time"

	"azadi-go/internal/model"
)

const (
	earlySettlementFeeBPS  = 200
	settlementValidityDays = 28
)

type AgreementGetter interface {
	GetAgreement(ctx context.Context, customerID string, agreementID int64) (*model.Agreement, error)
}

type settlementRepo interface {
	Save(ctx context.Context, s *model.SettlementFigure) (*model.SettlementFigure, error)
	FindByCustomerID(ctx context.Context, customerID string) ([]*model.SettlementFigure, error)
}

type Service struct {
	repo         settlementRepo
	agreementSvc AgreementGetter
}

func NewService(repo settlementRepo, agreementSvc AgreementGetter) *Service {
	return &Service{repo: repo, agreementSvc: agreementSvc}
}

func (s *Service) CalculateSettlement(ctx context.Context, customerID string, agreementID int64) (*model.SettlementFigure, error) {
	agreement, err := s.agreementSvc.GetAgreement(ctx, customerID, agreementID)
	if err != nil {
		return nil, fmt.Errorf("getting agreement: %w", err)
	}

	fee := agreement.BalancePence * earlySettlementFeeBPS / 10_000
	total := agreement.BalancePence + fee

	figure := &model.SettlementFigure{
		AgreementID:  agreementID,
		CustomerID:   customerID,
		AmountPence:  total,
		CalculatedAt: time.Now(),
		ValidUntil:   time.Now().AddDate(0, 0, settlementValidityDays),
	}
	saved, err := s.repo.Save(ctx, figure)
	if err != nil {
		return nil, fmt.Errorf("saving settlement: %w", err)
	}
	return saved, nil
}

func (s *Service) GetSettlementsForCustomer(ctx context.Context, customerID string) ([]*model.SettlementFigure, error) {
	return s.repo.FindByCustomerID(ctx, customerID)
}
