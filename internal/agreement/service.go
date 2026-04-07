package agreement

import (
	"context"
	"errors"
	"fmt"

	"azadi-go/internal/model"
)

var ErrNotFound = errors.New("agreement not found")
var ErrAccessDenied = errors.New("access denied")

type agreementRepo interface {
	FindByCustomerID(ctx context.Context, customerID string) ([]*model.Agreement, error)
	FindByID(ctx context.Context, id int64) (*model.Agreement, error)
}

type Service struct {
	repo agreementRepo
}

func NewService(repo agreementRepo) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAgreementsForCustomer(ctx context.Context, customerID string) ([]*model.Agreement, error) {
	return s.repo.FindByCustomerID(ctx, customerID)
}

func (s *Service) GetAgreement(ctx context.Context, customerID string, agreementID int64) (*model.Agreement, error) {
	a, err := s.repo.FindByID(ctx, agreementID)
	if err != nil {
		return nil, fmt.Errorf("getting agreement: %w", err)
	}
	if a == nil {
		return nil, ErrNotFound
	}
	if a.CustomerID != customerID {
		return nil, ErrAccessDenied
	}
	return a, nil
}
