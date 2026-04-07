package agreement

import (
	"context"
	"errors"
	"testing"

	"azadi-go/internal/model"
)

// --- mocks ---

type mockAgreementRepo struct {
	findByCustomerFn func(ctx context.Context, customerID string) ([]*model.Agreement, error)
	findByIDFn       func(ctx context.Context, id int64) (*model.Agreement, error)
}

func (m *mockAgreementRepo) FindByCustomerID(ctx context.Context, customerID string) ([]*model.Agreement, error) {
	return m.findByCustomerFn(ctx, customerID)
}

func (m *mockAgreementRepo) FindByID(ctx context.Context, id int64) (*model.Agreement, error) {
	return m.findByIDFn(ctx, id)
}

// --- tests ---

func TestGetAgreementsForCustomer(t *testing.T) {
	tests := []struct {
		name       string
		customerID string
		agreements []*model.Agreement
		repoErr    error
		wantErr    bool
		wantCount  int
	}{
		{
			name:       "returns agreements",
			customerID: "cust-1",
			agreements: []*model.Agreement{{Base: model.Base{ID: 1}}, {Base: model.Base{ID: 2}}},
			wantCount:  2,
		},
		{
			name:       "empty",
			customerID: "cust-2",
			agreements: nil,
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
			repo := &mockAgreementRepo{
				findByCustomerFn: func(_ context.Context, _ string) ([]*model.Agreement, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return tt.agreements, nil
				},
			}
			svc := NewService(repo)
			results, err := svc.GetAgreementsForCustomer(context.Background(), tt.customerID)
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

func TestGetAgreement(t *testing.T) {
	tests := []struct {
		name        string
		customerID  string
		agreementID int64
		agreement   *model.Agreement
		repoErr     error
		wantErr     error
	}{
		{
			name:        "success",
			customerID:  "cust-1",
			agreementID: 10,
			agreement:   &model.Agreement{Base: model.Base{ID: 10}, CustomerID: "cust-1", VehicleModel: "Golf"},
		},
		{
			name:        "not found - nil",
			customerID:  "cust-1",
			agreementID: 99,
			agreement:   nil,
			wantErr:     ErrNotFound,
		},
		{
			name:        "access denied - wrong customer",
			customerID:  "cust-1",
			agreementID: 10,
			agreement:   &model.Agreement{Base: model.Base{ID: 10}, CustomerID: "cust-other"},
			wantErr:     ErrAccessDenied,
		},
		{
			name:        "repo error",
			customerID:  "cust-1",
			agreementID: 10,
			repoErr:     errors.New("db error"),
			wantErr:     errors.New("getting agreement"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAgreementRepo{
				findByIDFn: func(_ context.Context, _ int64) (*model.Agreement, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return tt.agreement, nil
				},
			}
			svc := NewService(repo)
			result, err := svc.GetAgreement(context.Background(), tt.customerID, tt.agreementID)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if errors.Is(tt.wantErr, ErrNotFound) || errors.Is(tt.wantErr, ErrAccessDenied) {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("got error %v, want %v", err, tt.wantErr)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ID != tt.agreementID {
				t.Errorf("ID = %d, want %d", result.ID, tt.agreementID)
			}
			if result.CustomerID != tt.customerID {
				t.Errorf("CustomerID = %q, want %q", result.CustomerID, tt.customerID)
			}
		})
	}
}
