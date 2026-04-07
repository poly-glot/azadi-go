package bank

import (
	"context"
	"testing"
	"time"

	"azadi-go/internal/model"
	"azadi-go/internal/testutil"
)

func TestBankRepo_SaveAndFindByCustomerID(t *testing.T) {
	client := testutil.EmulatorClient(t, "BankDetails")
	repo := NewRepository(client)
	ctx := context.Background()

	updatedAt := time.Now().Truncate(time.Millisecond)
	bd := &model.BankDetails{
		CustomerID:             "cust-bank-1",
		AccountHolderName:      "Alice Smith",
		EncryptedAccountNumber: "enc-12345678",
		EncryptedSortCode:      "enc-112233",
		LastFourAccount:        "5678",
		LastTwoSortCode:        "33",
		UpdatedAt:              updatedAt,
	}

	saved, err := repo.Save(ctx, bd)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.ID == 0 {
		t.Fatal("expected non-zero ID after save")
	}

	found, err := repo.FindByCustomerID(ctx, "cust-bank-1")
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if found == nil {
		t.Fatal("expected bank details, got nil")
	}
	if found.ID != saved.ID {
		t.Errorf("ID: got %d, want %d", found.ID, saved.ID)
	}
	if found.CustomerID != "cust-bank-1" {
		t.Errorf("CustomerID: got %q, want %q", found.CustomerID, "cust-bank-1")
	}
	if found.AccountHolderName != "Alice Smith" {
		t.Errorf("AccountHolderName: got %q, want %q", found.AccountHolderName, "Alice Smith")
	}
	if found.EncryptedAccountNumber != "enc-12345678" {
		t.Errorf("EncryptedAccountNumber: got %q, want %q", found.EncryptedAccountNumber, "enc-12345678")
	}
	if found.LastFourAccount != "5678" {
		t.Errorf("LastFourAccount: got %q, want %q", found.LastFourAccount, "5678")
	}
	if !found.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", found.UpdatedAt, updatedAt)
	}
}

func TestBankRepo_FindByCustomerID_NotFound(t *testing.T) {
	client := testutil.EmulatorClient(t, "BankDetails")
	repo := NewRepository(client)
	ctx := context.Background()

	found, err := repo.FindByCustomerID(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if found != nil {
		t.Errorf("expected nil, got %+v", found)
	}
}

func TestBankRepo_Save_Update(t *testing.T) {
	client := testutil.EmulatorClient(t, "BankDetails")
	repo := NewRepository(client)
	ctx := context.Background()

	bd := &model.BankDetails{
		CustomerID:             "cust-bank-update",
		AccountHolderName:      "Bob Jones",
		EncryptedAccountNumber: "enc-old",
		EncryptedSortCode:      "enc-old-sc",
		LastFourAccount:        "1111",
		LastTwoSortCode:        "11",
		UpdatedAt:              time.Now().Truncate(time.Millisecond),
	}

	saved, err := repo.Save(ctx, bd)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	newTime := time.Now().Truncate(time.Millisecond)
	saved.EncryptedAccountNumber = "enc-new"
	saved.LastFourAccount = "9999"
	saved.UpdatedAt = newTime

	updated, err := repo.Save(ctx, saved)
	if err != nil {
		t.Fatalf("Save (update): %v", err)
	}
	if updated.ID != saved.ID {
		t.Errorf("ID changed: got %d, want %d", updated.ID, saved.ID)
	}

	found, err := repo.FindByCustomerID(ctx, "cust-bank-update")
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if found == nil {
		t.Fatal("expected bank details, got nil")
	}
	if found.EncryptedAccountNumber != "enc-new" {
		t.Errorf("EncryptedAccountNumber: got %q, want %q", found.EncryptedAccountNumber, "enc-new")
	}
	if found.LastFourAccount != "9999" {
		t.Errorf("LastFourAccount: got %q, want %q", found.LastFourAccount, "9999")
	}
	if !found.UpdatedAt.Equal(newTime) {
		t.Errorf("UpdatedAt: got %v, want %v", found.UpdatedAt, newTime)
	}
}
