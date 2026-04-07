package payment

import (
	"context"
	"testing"
	"time"

	"azadi-go/internal/model"
	"azadi-go/internal/testutil"
)

func TestPaymentRepo_SaveAndFindByCustomerID(t *testing.T) {
	client := testutil.EmulatorClient(t, "PaymentRecord")
	repo := NewRepository(client)
	ctx := context.Background()
	now := time.Now().Truncate(time.Millisecond)

	r1 := &model.PaymentRecord{
		AgreementID: 100,
		CustomerID:  "cust-1",
		AmountPence: 5000,
		Status:      model.PaymentStatusPending,
		CreatedAt:   now,
	}
	r2 := &model.PaymentRecord{
		AgreementID: 200,
		CustomerID:  "cust-1",
		AmountPence: 3000,
		Status:      model.PaymentStatusCompleted,
		CreatedAt:   now,
	}
	r3 := &model.PaymentRecord{
		AgreementID: 300,
		CustomerID:  "cust-other",
		AmountPence: 1000,
		Status:      model.PaymentStatusPending,
		CreatedAt:   now,
	}

	for _, r := range []*model.PaymentRecord{r1, r2, r3} {
		if _, err := repo.Save(ctx, r); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	results, err := repo.FindByCustomerID(ctx, "cust-1")
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 records, got %d", len(results))
	}
	for _, r := range results {
		if r.CustomerID != "cust-1" {
			t.Errorf("expected customerID cust-1, got %s", r.CustomerID)
		}
		if r.ID == 0 {
			t.Error("expected non-zero ID")
		}
	}
}

func TestPaymentRepo_FindByStripePaymentIntentID(t *testing.T) {
	client := testutil.EmulatorClient(t, "PaymentRecord")
	repo := NewRepository(client)
	ctx := context.Background()
	now := time.Now().Truncate(time.Millisecond)

	rec := &model.PaymentRecord{
		CustomerID:            "cust-pi",
		AmountPence:           2500,
		StripePaymentIntentID: "pi_test_123",
		Status:                model.PaymentStatusPending,
		CreatedAt:             now,
	}
	saved, err := repo.Save(ctx, rec)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := repo.FindByStripePaymentIntentID(ctx, "pi_test_123")
	if err != nil {
		t.Fatalf("FindByStripePaymentIntentID: %v", err)
	}
	if found == nil {
		t.Fatal("expected record, got nil")
	}
	if found.ID != saved.ID {
		t.Errorf("expected ID %d, got %d", saved.ID, found.ID)
	}
	if found.StripePaymentIntentID != "pi_test_123" {
		t.Errorf("expected pi_test_123, got %s", found.StripePaymentIntentID)
	}
	if found.AmountPence != 2500 {
		t.Errorf("expected 2500, got %d", found.AmountPence)
	}
}

func TestPaymentRepo_FindByStripePaymentIntentID_NotFound(t *testing.T) {
	client := testutil.EmulatorClient(t, "PaymentRecord")
	repo := NewRepository(client)
	ctx := context.Background()

	found, err := repo.FindByStripePaymentIntentID(ctx, "pi_nonexistent")
	if err != nil {
		t.Fatalf("FindByStripePaymentIntentID: %v", err)
	}
	if found != nil {
		t.Errorf("expected nil, got %+v", found)
	}
}

func TestPaymentRepo_FindByWebhookEventID(t *testing.T) {
	client := testutil.EmulatorClient(t, "PaymentRecord")
	repo := NewRepository(client)
	ctx := context.Background()
	now := time.Now().Truncate(time.Millisecond)

	rec := &model.PaymentRecord{
		CustomerID:     "cust-wh",
		AmountPence:    7700,
		WebhookEventID: "evt_abc123",
		Status:         model.PaymentStatusCompleted,
		CreatedAt:      now,
		CompletedAt:    now,
	}
	saved, err := repo.Save(ctx, rec)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := repo.FindByWebhookEventID(ctx, "evt_abc123")
	if err != nil {
		t.Fatalf("FindByWebhookEventID: %v", err)
	}
	if found == nil {
		t.Fatal("expected record, got nil")
	}
	if found.ID != saved.ID {
		t.Errorf("expected ID %d, got %d", saved.ID, found.ID)
	}
	if found.WebhookEventID != "evt_abc123" {
		t.Errorf("expected evt_abc123, got %s", found.WebhookEventID)
	}
	if found.AmountPence != 7700 {
		t.Errorf("expected 7700, got %d", found.AmountPence)
	}
}

func TestPaymentRepo_FindByWebhookEventID_NotFound(t *testing.T) {
	client := testutil.EmulatorClient(t, "PaymentRecord")
	repo := NewRepository(client)
	ctx := context.Background()

	found, err := repo.FindByWebhookEventID(ctx, "evt_nonexistent")
	if err != nil {
		t.Fatalf("FindByWebhookEventID: %v", err)
	}
	if found != nil {
		t.Errorf("expected nil, got %+v", found)
	}
}

func TestPaymentRepo_SaveAll(t *testing.T) {
	client := testutil.EmulatorClient(t, "PaymentRecord")
	repo := NewRepository(client)
	ctx := context.Background()
	now := time.Now().Truncate(time.Millisecond)

	records := []*model.PaymentRecord{
		{
			CustomerID:  "cust-batch",
			AmountPence: 1000,
			Status:      model.PaymentStatusPending,
			CreatedAt:   now,
		},
		{
			CustomerID:  "cust-batch",
			AmountPence: 2000,
			Status:      model.PaymentStatusPending,
			CreatedAt:   now,
		},
	}

	if err := repo.SaveAll(ctx, records); err != nil {
		t.Fatalf("SaveAll: %v", err)
	}

	for i, r := range records {
		if r.ID == 0 {
			t.Errorf("record[%d]: expected non-zero ID", i)
		}
	}
	if records[0].ID == records[1].ID {
		t.Error("expected different IDs for the two records")
	}
}

func TestPaymentRepo_Save_Update(t *testing.T) {
	client := testutil.EmulatorClient(t, "PaymentRecord")
	repo := NewRepository(client)
	ctx := context.Background()
	now := time.Now().Truncate(time.Millisecond)

	rec := &model.PaymentRecord{
		CustomerID:  "cust-update",
		AmountPence: 4200,
		Status:      model.PaymentStatusPending,
		CreatedAt:   now,
	}
	saved, err := repo.Save(ctx, rec)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	originalID := saved.ID

	saved.Status = model.PaymentStatusCompleted
	saved.CompletedAt = now
	updated, err := repo.Save(ctx, saved)
	if err != nil {
		t.Fatalf("Save (update): %v", err)
	}
	if updated.ID != originalID {
		t.Errorf("expected same ID %d, got %d", originalID, updated.ID)
	}

	// Verify via query
	results, err := repo.FindByCustomerID(ctx, "cust-update")
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 record, got %d", len(results))
	}
	if results[0].Status != model.PaymentStatusCompleted {
		t.Errorf("expected COMPLETED, got %s", results[0].Status)
	}
}
