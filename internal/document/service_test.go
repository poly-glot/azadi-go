package document

import (
	"context"
	"errors"
	"testing"

	"azadi-go/internal/model"
)

// --- mocks ---

type mockDocumentRepo struct {
	findByCustomerFn func(ctx context.Context, customerID string) ([]*model.Document, error)
	findByIDFn       func(ctx context.Context, id int64) (*model.Document, error)
}

func (m *mockDocumentRepo) FindByCustomerID(ctx context.Context, customerID string) ([]*model.Document, error) {
	return m.findByCustomerFn(ctx, customerID)
}

func (m *mockDocumentRepo) FindByID(ctx context.Context, id int64) (*model.Document, error) {
	return m.findByIDFn(ctx, id)
}

// --- tests ---

func TestGetDocumentsForCustomer(t *testing.T) {
	tests := []struct {
		name       string
		customerID string
		docs       []*model.Document
		repoErr    error
		wantErr    bool
		wantCount  int
	}{
		{
			name:       "returns documents",
			customerID: "cust-1",
			docs:       []*model.Document{{Base: model.Base{ID: 1}, CustomerID: "cust-1"}, {Base: model.Base{ID: 2}, CustomerID: "cust-1"}},
			wantCount:  2,
		},
		{
			name:       "empty",
			customerID: "cust-2",
			docs:       nil,
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
			repo := &mockDocumentRepo{
				findByCustomerFn: func(_ context.Context, _ string) ([]*model.Document, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return tt.docs, nil
				},
			}
			svc := NewService(repo)
			results, err := svc.GetDocumentsForCustomer(context.Background(), tt.customerID)
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

func TestGetDocument(t *testing.T) {
	tests := []struct {
		name       string
		customerID string
		docID      int64
		doc        *model.Document
		repoErr    error
		wantErr    error
	}{
		{
			name:       "success",
			customerID: "cust-1",
			docID:      10,
			doc:        &model.Document{Base: model.Base{ID: 10}, CustomerID: "cust-1", Title: "Statement"},
		},
		{
			name:       "not found - nil doc",
			customerID: "cust-1",
			docID:      99,
			doc:        nil,
			wantErr:    ErrNotFound,
		},
		{
			name:       "access denied - wrong customer",
			customerID: "cust-1",
			docID:      10,
			doc:        &model.Document{Base: model.Base{ID: 10}, CustomerID: "cust-other"},
			wantErr:    ErrAccessDenied,
		},
		{
			name:       "repo error",
			customerID: "cust-1",
			docID:      10,
			repoErr:    errors.New("db error"),
			wantErr:    errors.New("finding document"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDocumentRepo{
				findByIDFn: func(_ context.Context, _ int64) (*model.Document, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return tt.doc, nil
				},
			}
			svc := NewService(repo)
			result, err := svc.GetDocument(context.Background(), tt.customerID, tt.docID)

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
			if result.ID != tt.docID {
				t.Errorf("ID = %d, want %d", result.ID, tt.docID)
			}
			if result.CustomerID != tt.customerID {
				t.Errorf("CustomerID = %q, want %q", result.CustomerID, tt.customerID)
			}
		})
	}
}
