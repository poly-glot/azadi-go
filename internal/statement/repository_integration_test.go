package statement

import (
	"context"
	"testing"
	"time"

	"azadi-go/internal/model"
	"azadi-go/internal/testutil"
)

func TestStatementRepo_SaveAndFindByCustomerID(t *testing.T) {
	client := testutil.EmulatorClient(t, "StatementRequest")
	repo := NewRepository(client)
	ctx := context.Background()
	now := time.Now().Truncate(time.Millisecond)

	s1 := &model.StatementRequest{
		CustomerID:  "cust-stmt-1",
		AgreementID: 10,
		Status:      model.StatementStatusPending,
		RequestedAt: now,
	}
	s2 := &model.StatementRequest{
		CustomerID:  "cust-stmt-1",
		AgreementID: 20,
		Status:      model.StatementStatusSent,
		RequestedAt: now,
		FulfilledAt: now,
	}
	s3 := &model.StatementRequest{
		CustomerID:  "cust-stmt-other",
		AgreementID: 30,
		Status:      model.StatementStatusPending,
		RequestedAt: now,
	}

	for _, s := range []*model.StatementRequest{s1, s2, s3} {
		if _, err := repo.Save(ctx, s); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	results, err := repo.FindByCustomerID(ctx, "cust-stmt-1")
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 records, got %d", len(results))
	}
	for _, r := range results {
		if r.CustomerID != "cust-stmt-1" {
			t.Errorf("expected customerID cust-stmt-1, got %s", r.CustomerID)
		}
		if r.ID == 0 {
			t.Error("expected non-zero ID")
		}
	}
}

func TestStatementRepo_FindByCustomerID_Empty(t *testing.T) {
	client := testutil.EmulatorClient(t, "StatementRequest")
	repo := NewRepository(client)
	ctx := context.Background()

	results, err := repo.FindByCustomerID(ctx, "cust-nonexistent")
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 records, got %d", len(results))
	}
}

func TestStatementRepo_Save_GeneratesID(t *testing.T) {
	client := testutil.EmulatorClient(t, "StatementRequest")
	repo := NewRepository(client)
	ctx := context.Background()
	now := time.Now().Truncate(time.Millisecond)

	s := &model.StatementRequest{
		CustomerID:  "cust-gen-id",
		AgreementID: 42,
		Status:      model.StatementStatusPending,
		RequestedAt: now,
	}
	if s.ID != 0 {
		t.Fatal("expected zero ID before save")
	}

	saved, err := repo.Save(ctx, s)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.ID == 0 {
		t.Error("expected non-zero ID after save")
	}
	if saved.CustomerID != "cust-gen-id" {
		t.Errorf("expected cust-gen-id, got %s", saved.CustomerID)
	}
	if saved.AgreementID != 42 {
		t.Errorf("expected agreementID 42, got %d", saved.AgreementID)
	}
}
