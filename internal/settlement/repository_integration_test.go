package settlement

import (
	"context"
	"testing"
	"time"

	"azadi-go/internal/model"
	"azadi-go/internal/testutil"
)

func TestRepo_SaveAndFindByCustomerID(t *testing.T) {
	client := testutil.EmulatorClient(t, "SettlementFigure")
	repo := NewRepository(client)
	ctx := context.Background()

	now := time.Now().Truncate(time.Millisecond)
	custID := "CUST-100"

	s1 := &model.SettlementFigure{
		AgreementID:  1001,
		CustomerID:   custID,
		AmountPence:  1500000,
		CalculatedAt: now,
		ValidUntil:   now.Add(7 * 24 * time.Hour),
	}
	s2 := &model.SettlementFigure{
		AgreementID:  1002,
		CustomerID:   custID,
		AmountPence:  2500000,
		CalculatedAt: now,
		ValidUntil:   now.Add(14 * 24 * time.Hour),
	}

	if _, err := repo.Save(ctx, s1); err != nil {
		t.Fatalf("Save s1: %v", err)
	}
	if _, err := repo.Save(ctx, s2); err != nil {
		t.Fatalf("Save s2: %v", err)
	}

	results, err := repo.FindByCustomerID(ctx, custID)
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}

	agreementIDs := map[int64]bool{}
	for _, r := range results {
		if r.ID == 0 {
			t.Error("expected non-zero ID")
		}
		if r.CustomerID != custID {
			t.Errorf("CustomerID = %q, want %q", r.CustomerID, custID)
		}
		agreementIDs[r.AgreementID] = true
	}
	if !agreementIDs[1001] || !agreementIDs[1002] {
		t.Errorf("expected agreement IDs 1001 and 1002, got %v", agreementIDs)
	}
}

func TestRepo_Save_GeneratesID(t *testing.T) {
	client := testutil.EmulatorClient(t, "SettlementFigure")
	repo := NewRepository(client)
	ctx := context.Background()

	now := time.Now().Truncate(time.Millisecond)
	s := &model.SettlementFigure{
		AgreementID:  2001,
		CustomerID:   "CUST-200",
		AmountPence:  750000,
		CalculatedAt: now,
		ValidUntil:   now.Add(7 * 24 * time.Hour),
	}

	if s.ID != 0 {
		t.Fatal("expected ID to be 0 before save")
	}

	saved, err := repo.Save(ctx, s)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.ID == 0 {
		t.Fatal("expected non-zero ID after save")
	}

	// Verify the fields were persisted by querying back
	results, err := repo.FindByCustomerID(ctx, "CUST-200")
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	got := results[0]
	if got.ID != saved.ID {
		t.Errorf("ID = %d, want %d", got.ID, saved.ID)
	}
	if got.AgreementID != 2001 {
		t.Errorf("AgreementID = %d, want 2001", got.AgreementID)
	}
	if got.AmountPence != 750000 {
		t.Errorf("AmountPence = %d, want 750000", got.AmountPence)
	}
	if !got.CalculatedAt.Equal(now) {
		t.Errorf("CalculatedAt = %v, want %v", got.CalculatedAt, now)
	}
	if !got.ValidUntil.Equal(now.Add(7 * 24 * time.Hour)) {
		t.Errorf("ValidUntil = %v, want %v", got.ValidUntil, now.Add(7*24*time.Hour))
	}
}
