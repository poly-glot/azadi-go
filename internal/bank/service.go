package bank

import (
	"context"
	"errors"
	"fmt"
	"time"

	"azadi-go/internal/domain"
	"azadi-go/internal/model"
)

type bankRepo interface {
	FindByCustomerID(ctx context.Context, customerID string) (*model.BankDetails, error)
	Save(ctx context.Context, b *model.BankDetails) (*model.BankDetails, error)
}

type bankEncryptor interface {
	Encrypt(plaintext string) (string, error)
}

type Service struct {
	repo         bankRepo
	encryptor    bankEncryptor
	auditService domain.AuditLogger
	emailService interface {
		SendBankDetailsUpdated(customerID string)
	}
}

func NewService(repo bankRepo, encryptor bankEncryptor, auditService domain.AuditLogger,
	emailService interface{ SendBankDetailsUpdated(customerID string) }) *Service {
	return &Service{
		repo:         repo,
		encryptor:    encryptor,
		auditService: auditService,
		emailService: emailService,
	}
}

func (s *Service) GetBankDetails(ctx context.Context, customerID string) (*model.BankDetails, error) {
	return s.repo.FindByCustomerID(ctx, customerID)
}

func (s *Service) UpdateBankDetails(ctx context.Context, customerID, accountHolderName, accountNumber, sortCode, ipAddress, sessionID string) (*model.BankDetails, error) {
	existing, err := s.repo.FindByCustomerID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("finding bank details: %w", err)
	}
	if existing == nil {
		existing = &model.BankDetails{}
	}

	if len(accountNumber) < 4 {
		return nil, errors.New("account number too short")
	}
	if len(sortCode) < 2 {
		return nil, errors.New("sort code too short")
	}

	encAccount, err := s.encryptor.Encrypt(accountNumber)
	if err != nil {
		return nil, fmt.Errorf("encrypting account: %w", err)
	}
	encSort, err := s.encryptor.Encrypt(sortCode)
	if err != nil {
		return nil, fmt.Errorf("encrypting sort code: %w", err)
	}

	existing.CustomerID = customerID
	existing.AccountHolderName = accountHolderName
	existing.EncryptedAccountNumber = encAccount
	existing.EncryptedSortCode = encSort
	existing.LastFourAccount = accountNumber[len(accountNumber)-4:]
	existing.LastTwoSortCode = sortCode[len(sortCode)-2:]
	existing.UpdatedAt = time.Now()

	saved, err := s.repo.Save(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("saving bank details: %w", err)
	}

	s.auditService.LogEvent(customerID, "BANK_DETAILS_UPDATED", ipAddress, sessionID, map[string]string{
		"accountEnding": existing.LastFourAccount,
	})

	s.emailService.SendBankDetailsUpdated(customerID)

	return saved, nil
}
