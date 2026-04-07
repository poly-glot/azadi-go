package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"azadi-go/internal/model"
)

var (
	ErrBadCredentials = errors.New("invalid agreement number, date of birth, or postcode")
	ErrAccountLocked  = errors.New("account temporarily locked due to too many failed attempts")
)

type agreementFinder interface {
	FindByAgreementNumber(ctx context.Context, number string) (*model.Agreement, error)
}

type customerFinder interface {
	FindByCustomerID(ctx context.Context, customerID string) (*model.Customer, error)
}

type Provider struct {
	agreementRepo agreementFinder
	customerRepo  customerFinder
	tracker       *LoginAttemptTracker
}

func NewProvider(agreementRepo agreementFinder, customerRepo customerFinder, tracker *LoginAttemptTracker) *Provider {
	return &Provider{
		agreementRepo: agreementRepo,
		customerRepo:  customerRepo,
		tracker:       tracker,
	}
}

func (p *Provider) Authenticate(ctx context.Context, agreementNumber, dobStr, postcode string) (*SessionData, error) {
	if p.tracker.IsBlocked(agreementNumber) {
		return nil, ErrAccountLocked
	}

	// Parse DOB in d/M/yyyy format
	dob, err := parseDOB(dobStr)
	if err != nil {
		p.tracker.RecordFailure(agreementNumber)
		return nil, ErrBadCredentials
	}

	agreement, err := p.agreementRepo.FindByAgreementNumber(ctx, agreementNumber)
	if err != nil {
		return nil, fmt.Errorf("looking up agreement: %w", err)
	}
	if agreement == nil {
		p.tracker.RecordFailure(agreementNumber)
		return nil, ErrBadCredentials
	}

	customer, err := p.customerRepo.FindByCustomerID(ctx, agreement.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("looking up customer: %w", err)
	}
	if customer == nil {
		p.tracker.RecordFailure(agreementNumber)
		return nil, ErrBadCredentials
	}

	// Compare DOB (date only)
	customerDOB := customer.DOB
	if dob.Year() != customerDOB.Year() || dob.Month() != customerDOB.Month() || dob.Day() != customerDOB.Day() {
		p.tracker.RecordFailure(agreementNumber)
		return nil, ErrBadCredentials
	}

	// Compare postcode (normalized: uppercase, no spaces)
	normalizedInput := strings.ToUpper(strings.ReplaceAll(postcode, " ", ""))
	normalizedStored := strings.ToUpper(strings.ReplaceAll(customer.Postcode, " ", ""))
	if normalizedInput != normalizedStored {
		p.tracker.RecordFailure(agreementNumber)
		return nil, ErrBadCredentials
	}

	p.tracker.RecordSuccess(agreementNumber)

	return &SessionData{
		CustomerID:   customer.CustomerID,
		CustomerName: customer.FullName,
		AgreementNum: agreementNumber,
	}, nil
}

func parseDOB(s string) (time.Time, error) {
	// Try d/M/yyyy format
	t, err := time.Parse("2/1/2006", s)
	if err != nil {
		// Also try dd/MM/yyyy
		t, err = time.Parse("02/01/2006", s)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid DOB format: %s", s)
		}
	}
	return t, nil
}
