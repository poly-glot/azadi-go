package statement

import (
	"context"
	"errors"
	"testing"

	"azadi-go/internal/model"
)

// --- mocks ---

type mockStatementRepo struct {
	saveFn           func(ctx context.Context, s *model.StatementRequest) (*model.StatementRequest, error)
	findByCustomerFn func(ctx context.Context, customerID string) ([]*model.StatementRequest, error)
}

func (m *mockStatementRepo) Save(ctx context.Context, s *model.StatementRequest) (*model.StatementRequest, error) {
	return m.saveFn(ctx, s)
}

func (m *mockStatementRepo) FindByCustomerID(ctx context.Context, customerID string) ([]*model.StatementRequest, error) {
	return m.findByCustomerFn(ctx, customerID)
}

type mockAgreementGetter struct {
	getAgreementsFn func(ctx context.Context, customerID string) ([]*model.Agreement, error)
}

func (m *mockAgreementGetter) GetAgreementsForCustomer(ctx context.Context, customerID string) ([]*model.Agreement, error) {
	return m.getAgreementsFn(ctx, customerID)
}

type mockLogEventer struct {
	called    bool
	eventType string
}

func (m *mockLogEventer) LogEvent(customerID, eventType, ipAddress, sessionID string, details map[string]string) {
	m.called = true
	m.eventType = eventType
}

// --- tests ---

func TestResolveAgreementID(t *testing.T) {
	id42 := int64(42)
	tests := []struct {
		name        string
		customerID  string
		agreementID *int64
		agreements  []*model.Agreement
		agreeErr    error
		wantNil     bool
		wantID      int64
	}{
		{
			name:        "provided ID returned as-is",
			customerID:  "cust-1",
			agreementID: &id42,
			wantID:      42,
		},
		{
			name:       "nil resolves to first agreement",
			customerID: "cust-1",
			agreements: []*model.Agreement{{Base: model.Base{ID: 10}}, {Base: model.Base{ID: 20}}},
			wantID:     10,
		},
		{
			name:       "nil with no agreements returns nil",
			customerID: "cust-1",
			agreements: nil,
			wantNil:    true,
		},
		{
			name:       "nil with agreement error returns nil",
			customerID: "cust-1",
			agreeErr:   errors.New("db error"),
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agSvc := &mockAgreementGetter{
				getAgreementsFn: func(_ context.Context, _ string) ([]*model.Agreement, error) {
					if tt.agreeErr != nil {
						return nil, tt.agreeErr
					}
					return tt.agreements, nil
				},
			}
			svc := NewService(nil, nil, agSvc)
			result := svc.ResolveAgreementID(context.Background(), tt.customerID, tt.agreementID)
			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %d", *result)
				}
				return
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if *result != tt.wantID {
				t.Errorf("got %d, want %d", *result, tt.wantID)
			}
		})
	}
}

func TestRequestStatement(t *testing.T) {
	tests := []struct {
		name        string
		customerID  string
		agreementID int64
		saveErr     error
		wantErr     bool
	}{
		{
			name:        "success",
			customerID:  "cust-1",
			agreementID: 42,
		},
		{
			name:        "save error",
			customerID:  "cust-1",
			agreementID: 42,
			saveErr:     errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockStatementRepo{
				saveFn: func(_ context.Context, s *model.StatementRequest) (*model.StatementRequest, error) {
					if tt.saveErr != nil {
						return nil, tt.saveErr
					}
					s.ID = 1
					return s, nil
				},
			}
			audit := &mockLogEventer{}
			svc := NewService(repo, audit, nil)
			result, err := svc.RequestStatement(context.Background(), tt.customerID, tt.agreementID, "127.0.0.1", "sess-1")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.CustomerID != tt.customerID {
				t.Errorf("CustomerID = %q, want %q", result.CustomerID, tt.customerID)
			}
			if result.AgreementID != tt.agreementID {
				t.Errorf("AgreementID = %d, want %d", result.AgreementID, tt.agreementID)
			}
			if result.Status != model.StatementStatusPending {
				t.Errorf("Status = %q, want %q", result.Status, model.StatementStatusPending)
			}
			if !audit.called {
				t.Error("expected audit LogEvent to be called")
			}
			if audit.eventType != "STATEMENT_REQUESTED" {
				t.Errorf("audit eventType = %q, want %q", audit.eventType, "STATEMENT_REQUESTED")
			}
		})
	}
}

func TestGetStatementsForCustomer(t *testing.T) {
	tests := []struct {
		name       string
		customerID string
		stmts      []*model.StatementRequest
		repoErr    error
		wantErr    bool
		wantCount  int
	}{
		{
			name:       "returns statements",
			customerID: "cust-1",
			stmts:      []*model.StatementRequest{{Base: model.Base{ID: 1}}, {Base: model.Base{ID: 2}}},
			wantCount:  2,
		},
		{
			name:       "empty",
			customerID: "cust-2",
			stmts:      nil,
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
			repo := &mockStatementRepo{
				findByCustomerFn: func(_ context.Context, _ string) ([]*model.StatementRequest, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return tt.stmts, nil
				},
			}
			svc := NewService(repo, nil, nil)
			results, err := svc.GetStatementsForCustomer(context.Background(), tt.customerID)
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
