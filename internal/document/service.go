package document

import (
	"context"
	"errors"
	"fmt"

	"azadi-go/internal/model"
)

var ErrNotFound = errors.New("document not found")
var ErrAccessDenied = errors.New("access denied")

type documentRepo interface {
	FindByCustomerID(ctx context.Context, customerID string) ([]*model.Document, error)
	FindByID(ctx context.Context, id int64) (*model.Document, error)
}

type Service struct {
	repo documentRepo
}

func NewService(repo documentRepo) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetDocumentsForCustomer(ctx context.Context, customerID string) ([]*model.Document, error) {
	return s.repo.FindByCustomerID(ctx, customerID)
}

func (s *Service) GetDocument(ctx context.Context, customerID string, documentID int64) (*model.Document, error) {
	d, err := s.repo.FindByID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("finding document: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}
	if d.CustomerID != customerID {
		return nil, ErrAccessDenied
	}
	return d, nil
}
