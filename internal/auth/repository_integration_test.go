package auth

import (
	"context"
	"testing"
	"time"

	"azadi-go/internal/model"
	"azadi-go/internal/testutil"
)

func TestCustomerRepo_SaveAndFindByCustomerID(t *testing.T) {
	client := testutil.EmulatorClient(t, "Customer")
	repo := NewCustomerRepository(client)
	ctx := context.Background()

	dob := time.Now().Truncate(time.Millisecond)
	c := &model.Customer{
		CustomerID:   "cust-001",
		FullName:     "Alice Smith",
		Email:        "alice@example.com",
		DOB:          dob,
		Postcode:     "SW1A 1AA",
		Phone:        "020-1234",
		MobilePhone:  "07700-900000",
		AddressLine1: "10 Downing St",
		AddressLine2: "",
		City:         "London",
	}

	saved, err := repo.Save(ctx, c)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.ID == 0 {
		t.Fatal("expected non-zero ID after save")
	}

	found, err := repo.FindByCustomerID(ctx, "cust-001")
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if found == nil {
		t.Fatal("expected customer, got nil")
	}
	if found.ID != saved.ID {
		t.Errorf("ID: got %d, want %d", found.ID, saved.ID)
	}
	if found.CustomerID != "cust-001" {
		t.Errorf("CustomerID: got %q, want %q", found.CustomerID, "cust-001")
	}
	if found.FullName != "Alice Smith" {
		t.Errorf("FullName: got %q, want %q", found.FullName, "Alice Smith")
	}
	if found.Email != "alice@example.com" {
		t.Errorf("Email: got %q, want %q", found.Email, "alice@example.com")
	}
	if !found.DOB.Equal(dob) {
		t.Errorf("DOB: got %v, want %v", found.DOB, dob)
	}
	if found.City != "London" {
		t.Errorf("City: got %q, want %q", found.City, "London")
	}
}

func TestCustomerRepo_FindByCustomerID_NotFound(t *testing.T) {
	client := testutil.EmulatorClient(t, "Customer")
	repo := NewCustomerRepository(client)
	ctx := context.Background()

	found, err := repo.FindByCustomerID(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if found != nil {
		t.Errorf("expected nil, got %+v", found)
	}
}

func TestCustomerRepo_Save_Update(t *testing.T) {
	client := testutil.EmulatorClient(t, "Customer")
	repo := NewCustomerRepository(client)
	ctx := context.Background()

	c := &model.Customer{
		CustomerID: "cust-update",
		FullName:   "Bob Jones",
		Email:      "bob@example.com",
	}

	saved, err := repo.Save(ctx, c)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	saved.Email = "bob.updated@example.com"
	updated, err := repo.Save(ctx, saved)
	if err != nil {
		t.Fatalf("Save (update): %v", err)
	}
	if updated.ID != saved.ID {
		t.Errorf("ID changed: got %d, want %d", updated.ID, saved.ID)
	}

	found, err := repo.FindByCustomerID(ctx, "cust-update")
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if found == nil {
		t.Fatal("expected customer, got nil")
	}
	if found.Email != "bob.updated@example.com" {
		t.Errorf("Email: got %q, want %q", found.Email, "bob.updated@example.com")
	}
}

func TestCustomerRepo_FindAll(t *testing.T) {
	client := testutil.EmulatorClient(t, "Customer")
	repo := NewCustomerRepository(client)
	ctx := context.Background()

	customers := []*model.Customer{
		{CustomerID: "cust-all-1", FullName: "One", Email: "one@example.com"},
		{CustomerID: "cust-all-2", FullName: "Two", Email: "two@example.com"},
	}
	for _, c := range customers {
		if _, err := repo.Save(ctx, c); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(all) < 2 {
		t.Fatalf("FindAll: got %d customers, want at least 2", len(all))
	}

	ids := make(map[string]bool)
	for _, c := range all {
		ids[c.CustomerID] = true
	}
	if !ids["cust-all-1"] || !ids["cust-all-2"] {
		t.Errorf("FindAll missing expected customers: got IDs %v", ids)
	}
}
