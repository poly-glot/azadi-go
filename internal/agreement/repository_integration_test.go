package agreement

import (
	"context"
	"testing"
	"time"

	"azadi-go/internal/model"
	"azadi-go/internal/testutil"
)

func TestRepo_SaveAndFindByID(t *testing.T) {
	client := testutil.EmulatorClient(t, "Agreement")
	repo := NewRepository(client)
	ctx := context.Background()

	now := time.Now().Truncate(time.Millisecond)
	a := &model.Agreement{
		AgreementNumber:         "AGR-001",
		CustomerID:              "CUST-100",
		Type:                    "PCP",
		BalancePence:            1500000,
		APR:                     "6.9",
		OriginalTermMonths:      48,
		ContractMileage:         40000,
		ExcessPricePerMilePence: 12,
		VehicleModel:            "Focus",
		Registration:            "AB12 CDE",
		LastPaymentPence:        25000,
		LastPaymentDate:         now.Add(-30 * 24 * time.Hour),
		NextPaymentPence:        25000,
		NextPaymentDate:         now.Add(24 * time.Hour),
		PaymentsRemaining:       12,
		FinalPaymentDate:        now.Add(365 * 24 * time.Hour),
		PaymentDateChanged:      true,
	}

	saved, err := repo.Save(ctx, a)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.ID == 0 {
		t.Fatal("expected non-zero ID after save")
	}

	got, err := repo.FindByID(ctx, saved.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got == nil {
		t.Fatal("FindByID returned nil")
	}
	if got.ID != saved.ID {
		t.Errorf("ID = %d, want %d", got.ID, saved.ID)
	}
	if got.AgreementNumber != "AGR-001" {
		t.Errorf("AgreementNumber = %q, want %q", got.AgreementNumber, "AGR-001")
	}
	if got.CustomerID != "CUST-100" {
		t.Errorf("CustomerID = %q, want %q", got.CustomerID, "CUST-100")
	}
	if got.Type != "PCP" {
		t.Errorf("Type = %q, want %q", got.Type, "PCP")
	}
	if got.BalancePence != 1500000 {
		t.Errorf("BalancePence = %d, want %d", got.BalancePence, 1500000)
	}
	if got.APR != "6.9" {
		t.Errorf("APR = %q, want %q", got.APR, "6.9")
	}
	if got.OriginalTermMonths != 48 {
		t.Errorf("OriginalTermMonths = %d, want %d", got.OriginalTermMonths, 48)
	}
	if got.ContractMileage != 40000 {
		t.Errorf("ContractMileage = %d, want %d", got.ContractMileage, 40000)
	}
	if got.ExcessPricePerMilePence != 12 {
		t.Errorf("ExcessPricePerMilePence = %d, want %d", got.ExcessPricePerMilePence, 12)
	}
	if got.VehicleModel != "Focus" {
		t.Errorf("VehicleModel = %q, want %q", got.VehicleModel, "Focus")
	}
	if got.Registration != "AB12 CDE" {
		t.Errorf("Registration = %q, want %q", got.Registration, "AB12 CDE")
	}
	if got.PaymentsRemaining != 12 {
		t.Errorf("PaymentsRemaining = %d, want %d", got.PaymentsRemaining, 12)
	}
	if got.PaymentDateChanged != true {
		t.Error("PaymentDateChanged = false, want true")
	}
	if !got.LastPaymentDate.Equal(a.LastPaymentDate) {
		t.Errorf("LastPaymentDate = %v, want %v", got.LastPaymentDate, a.LastPaymentDate)
	}
	if !got.NextPaymentDate.Equal(a.NextPaymentDate) {
		t.Errorf("NextPaymentDate = %v, want %v", got.NextPaymentDate, a.NextPaymentDate)
	}
	if !got.FinalPaymentDate.Equal(a.FinalPaymentDate) {
		t.Errorf("FinalPaymentDate = %v, want %v", got.FinalPaymentDate, a.FinalPaymentDate)
	}
}

func TestRepo_FindByCustomerID(t *testing.T) {
	client := testutil.EmulatorClient(t, "Agreement")
	repo := NewRepository(client)
	ctx := context.Background()

	now := time.Now().Truncate(time.Millisecond)
	custID := "CUST-200"

	a1 := &model.Agreement{
		AgreementNumber: "AGR-010",
		CustomerID:      custID,
		Type:            "PCP",
		BalancePence:    100000,
		NextPaymentDate: now,
	}
	a2 := &model.Agreement{
		AgreementNumber: "AGR-011",
		CustomerID:      custID,
		Type:            "HP",
		BalancePence:    200000,
		NextPaymentDate: now,
	}

	if _, err := repo.Save(ctx, a1); err != nil {
		t.Fatalf("Save a1: %v", err)
	}
	if _, err := repo.Save(ctx, a2); err != nil {
		t.Fatalf("Save a2: %v", err)
	}

	results, err := repo.FindByCustomerID(ctx, custID)
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}

	numbers := map[string]bool{}
	for _, r := range results {
		if r.ID == 0 {
			t.Error("expected non-zero ID")
		}
		numbers[r.AgreementNumber] = true
	}
	if !numbers["AGR-010"] || !numbers["AGR-011"] {
		t.Errorf("expected AGR-010 and AGR-011 in results, got %v", numbers)
	}
}

func TestRepo_FindByAgreementNumber(t *testing.T) {
	client := testutil.EmulatorClient(t, "Agreement")
	repo := NewRepository(client)
	ctx := context.Background()

	now := time.Now().Truncate(time.Millisecond)
	a := &model.Agreement{
		AgreementNumber: "AGR-UNIQUE-300",
		CustomerID:      "CUST-300",
		Type:            "Lease",
		BalancePence:    500000,
		NextPaymentDate: now,
	}

	saved, err := repo.Save(ctx, a)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByAgreementNumber(ctx, "AGR-UNIQUE-300")
	if err != nil {
		t.Fatalf("FindByAgreementNumber: %v", err)
	}
	if got == nil {
		t.Fatal("FindByAgreementNumber returned nil")
	}
	if got.ID != saved.ID {
		t.Errorf("ID = %d, want %d", got.ID, saved.ID)
	}
	if got.AgreementNumber != "AGR-UNIQUE-300" {
		t.Errorf("AgreementNumber = %q, want %q", got.AgreementNumber, "AGR-UNIQUE-300")
	}
	if got.CustomerID != "CUST-300" {
		t.Errorf("CustomerID = %q, want %q", got.CustomerID, "CUST-300")
	}
}

func TestRepo_FindByAgreementNumber_NotFound(t *testing.T) {
	client := testutil.EmulatorClient(t, "Agreement")
	repo := NewRepository(client)
	ctx := context.Background()

	got, err := repo.FindByAgreementNumber(ctx, "NONEXISTENT-999")
	if err != nil {
		t.Fatalf("FindByAgreementNumber: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestRepo_FindByID_NotFound(t *testing.T) {
	client := testutil.EmulatorClient(t, "Agreement")
	repo := NewRepository(client)
	ctx := context.Background()

	got, err := repo.FindByID(ctx, 999999999)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestRepo_Save_Update(t *testing.T) {
	client := testutil.EmulatorClient(t, "Agreement")
	repo := NewRepository(client)
	ctx := context.Background()

	now := time.Now().Truncate(time.Millisecond)
	a := &model.Agreement{
		AgreementNumber: "AGR-UPD-400",
		CustomerID:      "CUST-400",
		Type:            "PCP",
		BalancePence:    100000,
		NextPaymentDate: now,
	}

	saved, err := repo.Save(ctx, a)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	originalID := saved.ID

	// Modify and save again with same ID
	saved.BalancePence = 80000
	saved.Type = "HP"
	saved.PaymentDateChanged = true

	updated, err := repo.Save(ctx, saved)
	if err != nil {
		t.Fatalf("Save (update): %v", err)
	}
	if updated.ID != originalID {
		t.Errorf("ID changed: got %d, want %d", updated.ID, originalID)
	}

	got, err := repo.FindByID(ctx, originalID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got.BalancePence != 80000 {
		t.Errorf("BalancePence = %d, want 80000", got.BalancePence)
	}
	if got.Type != "HP" {
		t.Errorf("Type = %q, want %q", got.Type, "HP")
	}
	if got.PaymentDateChanged != true {
		t.Error("PaymentDateChanged = false, want true")
	}
	if got.AgreementNumber != "AGR-UPD-400" {
		t.Errorf("AgreementNumber = %q, want %q", got.AgreementNumber, "AGR-UPD-400")
	}
}
