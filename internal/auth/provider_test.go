package auth

import (
	"context"
	"testing"
	"time"

	"azadi-go/internal/model"
)

// Mock agreement repository satisfies the interface in Provider.
type mockAgreementRepo struct {
	agreements map[string]*model.Agreement
}

func (m *mockAgreementRepo) FindByAgreementNumber(_ context.Context, number string) (*model.Agreement, error) {
	if a, ok := m.agreements[number]; ok {
		return a, nil
	}
	return nil, nil
}

func TestParseDOB(t *testing.T) {
	tests := []struct {
		input string
		valid bool
		day   int
		month int
		year  int
	}{
		{"15/3/1985", true, 15, 3, 1985},
		{"1/1/2000", true, 1, 1, 2000},
		{"01/01/2000", true, 1, 1, 2000},
		{"invalid", false, 0, 0, 0},
		{"32/13/2000", false, 0, 0, 0},
		{"", false, 0, 0, 0},
		{"15/03/1985", true, 15, 3, 1985},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parsed, err := parseDOB(tt.input)
			if tt.valid {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if parsed.Day() != tt.day || int(parsed.Month()) != tt.month || parsed.Year() != tt.year {
					t.Errorf("got %v, want %d/%d/%d", parsed, tt.day, tt.month, tt.year)
				}
			} else if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestProvider_Authenticate_BadDOB(t *testing.T) {
	tracker := NewLoginAttemptTracker(context.Background())
	agrRepo := &mockAgreementRepo{agreements: map[string]*model.Agreement{
		"AGR-001": {AgreementNumber: "AGR-001", CustomerID: "CUST-001"},
	}}
	provider := &Provider{
		agreementRepo: agrRepo,
		customerRepo:  nil,
		tracker:       tracker,
	}

	_, err := provider.Authenticate(context.Background(), "AGR-001", "not-a-date", "SW1A 1AA")
	if err != ErrBadCredentials {
		t.Errorf("got %v, want ErrBadCredentials", err)
	}
}

func TestProvider_Authenticate_UnknownAgreement(t *testing.T) {
	tracker := NewLoginAttemptTracker(context.Background())
	agrRepo := &mockAgreementRepo{agreements: map[string]*model.Agreement{}}
	provider := &Provider{
		agreementRepo: agrRepo,
		customerRepo:  nil,
		tracker:       tracker,
	}

	_, err := provider.Authenticate(context.Background(), "UNKNOWN", "1/1/2000", "SW1A 1AA")
	if err != ErrBadCredentials {
		t.Errorf("got %v, want ErrBadCredentials", err)
	}
}

func TestProvider_Authenticate_LockedAccount(t *testing.T) {
	tracker := NewLoginAttemptTracker(context.Background())
	agrRepo := &mockAgreementRepo{agreements: map[string]*model.Agreement{
		"AGR-001": {AgreementNumber: "AGR-001", CustomerID: "CUST-001"},
	}}
	provider := &Provider{
		agreementRepo: agrRepo,
		customerRepo:  nil,
		tracker:       tracker,
	}

	// Exhaust attempts to trigger lock
	for i := 0; i < 10; i++ {
		tracker.RecordFailure("AGR-001")
	}

	_, err := provider.Authenticate(context.Background(), "AGR-001", "1/1/2000", "SW1A 1AA")
	if err != ErrAccountLocked {
		t.Errorf("got %v, want ErrAccountLocked", err)
	}
}

type mockCustomerRepo struct {
	customers map[string]*model.Customer
}

func (m *mockCustomerRepo) FindByCustomerID(_ context.Context, customerID string) (*model.Customer, error) {
	if c, ok := m.customers[customerID]; ok {
		return c, nil
	}
	return nil, nil
}

func TestProvider_Authenticate_Success(t *testing.T) {
	tracker := NewLoginAttemptTracker(context.Background())
	agrRepo := &mockAgreementRepo{agreements: map[string]*model.Agreement{
		"AGR-001": {AgreementNumber: "AGR-001", CustomerID: "CUST-001"},
	}}
	custRepo := &mockCustomerRepo{customers: map[string]*model.Customer{
		"CUST-001": {
			CustomerID: "CUST-001",
			FullName:   "James Smith",
			DOB:        time.Date(1985, 3, 15, 0, 0, 0, 0, time.UTC),
			Postcode:   "SW1A1AA",
		},
	}}
	provider := NewProvider(agrRepo, custRepo, tracker)

	session, err := provider.Authenticate(context.Background(), "AGR-001", "15/3/1985", "SW1A 1AA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.CustomerID != "CUST-001" {
		t.Errorf("customerID = %q", session.CustomerID)
	}
	if session.CustomerName != "James Smith" {
		t.Errorf("name = %q", session.CustomerName)
	}
	if session.AgreementNum != "AGR-001" {
		t.Errorf("agreement = %q", session.AgreementNum)
	}
}

func TestProvider_Authenticate_WrongDOB(t *testing.T) {
	agrRepo := &mockAgreementRepo{agreements: map[string]*model.Agreement{
		"AGR-001": {AgreementNumber: "AGR-001", CustomerID: "CUST-001"},
	}}
	custRepo := &mockCustomerRepo{customers: map[string]*model.Customer{
		"CUST-001": {
			CustomerID: "CUST-001",
			DOB:        time.Date(1985, 3, 15, 0, 0, 0, 0, time.UTC),
			Postcode:   "SW1A1AA",
		},
	}}
	provider := NewProvider(agrRepo, custRepo, NewLoginAttemptTracker(context.Background()))

	_, err := provider.Authenticate(context.Background(), "AGR-001", "1/1/2000", "SW1A 1AA")
	if err != ErrBadCredentials {
		t.Errorf("got %v, want ErrBadCredentials", err)
	}
}

func TestProvider_Authenticate_WrongPostcode(t *testing.T) {
	agrRepo := &mockAgreementRepo{agreements: map[string]*model.Agreement{
		"AGR-001": {AgreementNumber: "AGR-001", CustomerID: "CUST-001"},
	}}
	custRepo := &mockCustomerRepo{customers: map[string]*model.Customer{
		"CUST-001": {
			CustomerID: "CUST-001",
			DOB:        time.Date(1985, 3, 15, 0, 0, 0, 0, time.UTC),
			Postcode:   "SW1A1AA",
		},
	}}
	provider := NewProvider(agrRepo, custRepo, NewLoginAttemptTracker(context.Background()))

	_, err := provider.Authenticate(context.Background(), "AGR-001", "15/3/1985", "E1 6AN")
	if err != ErrBadCredentials {
		t.Errorf("got %v, want ErrBadCredentials", err)
	}
}

func TestProvider_Authenticate_CustomerNotFound(t *testing.T) {
	agrRepo := &mockAgreementRepo{agreements: map[string]*model.Agreement{
		"AGR-001": {AgreementNumber: "AGR-001", CustomerID: "CUST-MISSING"},
	}}
	custRepo := &mockCustomerRepo{customers: map[string]*model.Customer{}}
	provider := NewProvider(agrRepo, custRepo, NewLoginAttemptTracker(context.Background()))

	_, err := provider.Authenticate(context.Background(), "AGR-001", "15/3/1985", "SW1A 1AA")
	if err != ErrBadCredentials {
		t.Errorf("got %v, want ErrBadCredentials", err)
	}
}

// Verify parseDOB produces a time.Time that compares correctly for the
// DOB-matching logic in Authenticate.
func TestParseDOB_ComparisonLogic(t *testing.T) {
	parsed, err := parseDOB("15/3/1985")
	if err != nil {
		t.Fatal(err)
	}

	customerDOB := time.Date(1985, 3, 15, 10, 30, 0, 0, time.UTC)
	if parsed.Year() != customerDOB.Year() || parsed.Month() != customerDOB.Month() || parsed.Day() != customerDOB.Day() {
		t.Errorf("DOB mismatch: parsed=%v customer=%v", parsed, customerDOB)
	}
}
