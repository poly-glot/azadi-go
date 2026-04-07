package contact

import (
	"context"
	"errors"
	"fmt"

	"azadi-go/internal/domain"
	"azadi-go/internal/model"
)

var ErrNotFound = errors.New("customer not found")

type contactCustomerRepo interface {
	FindByCustomerID(ctx context.Context, customerID string) (*model.Customer, error)
	Save(ctx context.Context, c *model.Customer) (*model.Customer, error)
}

type Service struct {
	customerRepo contactCustomerRepo
	auditService domain.AuditLogger
}

func NewService(customerRepo contactCustomerRepo, auditService domain.AuditLogger) *Service {
	return &Service{customerRepo: customerRepo, auditService: auditService}
}

func (s *Service) GetCustomer(ctx context.Context, customerID string) (*model.Customer, error) {
	c, err := s.customerRepo.FindByCustomerID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("finding customer: %w", err)
	}
	if c == nil {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *Service) UpdateContactDetails(ctx context.Context, customerID string, phone, mobilePhone, email, addressLine1, addressLine2, city, postcode, ipAddress, sessionID string) error {
	customer, err := s.GetCustomer(ctx, customerID)
	if err != nil {
		return err
	}

	customer.Phone = phone
	customer.MobilePhone = mobilePhone
	customer.Email = email
	customer.AddressLine1 = addressLine1
	customer.AddressLine2 = addressLine2
	customer.City = city
	customer.Postcode = postcode

	if _, err := s.customerRepo.Save(ctx, customer); err != nil {
		return fmt.Errorf("saving customer: %w", err)
	}

	s.auditService.LogEvent(customerID, "CONTACT_DETAILS_UPDATED", ipAddress, sessionID, map[string]string{
		"email": email,
	})
	return nil
}
