package settlement

import (
	"context"
	"errors"
	"testing"

	"azadi-go/internal/model"
)

// --- mocks ---

type mockSettlementRepo struct {
	saveFn           func(ctx context.Context, s *model.SettlementFigure) (*model.SettlementFigure, error)
	findByCustomerFn func(ctx context.Context, customerID string) ([]*model.SettlementFigure, error)
}

func (m *mockSettlementRepo) Save(ctx context.Context, s *model.SettlementFigure) (*model.SettlementFigure, error) {
	return m.saveFn(ctx, s)
}

func (m *mockSettlementRepo) FindByCustomerID(ctx context.Context, customerID string) ([]*model.SettlementFigure, error) {
	return m.findByCustomerFn(ctx, customerID)
}

type mockAgreementGetter struct {
	getAgreementFn func(ctx context.Context, customerID string, agreementID int64) (*model.Agreement, error)
}

func (m *mockAgreementGetter) GetAgreement(ctx context.Context, customerID string, agreementID int64) (*model.Agreement, error) {
	return m.getAgreementFn(ctx, customerID, agreementID)
}

// --- tests ---

func TestSettlementCalculation(t *testing.T) {
	tests := []struct {
		name      string
		balance   int64
		wantTotal int64
	}{
		{"standard", 1000000, 1020000},
		{"zero", 0, 0},
		{"small", 10000, 10200},
		{"large", 10000000, 10200000},
		{"one penny", 100, 102},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fee := tt.balance * earlySettlementFeeBPS / 10_000
			total := tt.balance + fee
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

func TestSettlementFeeRate(t *testing.T) {
	if earlySettlementFeeBPS != 200 {
		t.Errorf("earlySettlementFeeBPS = %d, want 200", earlySettlementFeeBPS)
	}
}

func TestSettlementValidityDays(t *testing.T) {
	if settlementValidityDays != 28 {
		t.Errorf("settlementValidityDays = %d, want 28", settlementValidityDays)
	}
}

func TestCalculateSettlement(t *testing.T) {
	tests := []struct {
		name        string
		balance     int64
		agreementID int64
		customerID  string
		agreeErr    error
		saveErr     error
		wantErr     bool
		wantAmount  int64
	}{
		{
			name:        "success",
			balance:     500000,
			agreementID: 42,
			customerID:  "cust-1",
			wantAmount:  510000, // 500000 + 2% fee
		},
		{
			name:        "zero balance",
			balance:     0,
			agreementID: 1,
			customerID:  "cust-1",
			wantAmount:  0,
		},
		{
			name:        "agreement error",
			agreementID: 99,
			customerID:  "cust-1",
			agreeErr:    errors.New("not found"),
			wantErr:     true,
		},
		{
			name:        "save error",
			balance:     100000,
			agreementID: 1,
			customerID:  "cust-1",
			saveErr:     errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockSettlementRepo{
				saveFn: func(_ context.Context, s *model.SettlementFigure) (*model.SettlementFigure, error) {
					if tt.saveErr != nil {
						return nil, tt.saveErr
					}
					s.ID = 1
					return s, nil
				},
			}
			agSvc := &mockAgreementGetter{
				getAgreementFn: func(_ context.Context, _ string, _ int64) (*model.Agreement, error) {
					if tt.agreeErr != nil {
						return nil, tt.agreeErr
					}
					return &model.Agreement{Base: model.Base{ID: tt.agreementID}, BalancePence: tt.balance, CustomerID: tt.customerID}, nil
				},
			}
			svc := NewService(repo, agSvc)
			result, err := svc.CalculateSettlement(context.Background(), tt.customerID, tt.agreementID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.AmountPence != tt.wantAmount {
				t.Errorf("AmountPence = %d, want %d", result.AmountPence, tt.wantAmount)
			}
			if result.AgreementID != tt.agreementID {
				t.Errorf("AgreementID = %d, want %d", result.AgreementID, tt.agreementID)
			}
			if result.CustomerID != tt.customerID {
				t.Errorf("CustomerID = %q, want %q", result.CustomerID, tt.customerID)
			}
			if result.ValidUntil.Before(result.CalculatedAt) {
				t.Error("ValidUntil should be after CalculatedAt")
			}
		})
	}
}

func TestGetSettlementsForCustomer(t *testing.T) {
	tests := []struct {
		name       string
		customerID string
		figures    []*model.SettlementFigure
		repoErr    error
		wantErr    bool
		wantCount  int
	}{
		{
			name:       "returns figures",
			customerID: "cust-1",
			figures:    []*model.SettlementFigure{{Base: model.Base{ID: 1}}, {Base: model.Base{ID: 2}}},
			wantCount:  2,
		},
		{
			name:       "empty",
			customerID: "cust-2",
			figures:    nil,
			wantCount:  0,
		},
		{
			name:       "repo error",
			customerID: "cust-3",
			repoErr:    errors.New("db error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockSettlementRepo{
				findByCustomerFn: func(_ context.Context, _ string) ([]*model.SettlementFigure, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return tt.figures, nil
				},
			}
			svc := NewService(repo, nil)
			results, err := svc.GetSettlementsForCustomer(context.Background(), tt.customerID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(results) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(results), tt.wantCount)
			}
		})
	}
}
