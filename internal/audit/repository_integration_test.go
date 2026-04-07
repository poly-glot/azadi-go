package audit

import (
	"context"
	"testing"
	"time"

	"azadi-go/internal/model"
	"azadi-go/internal/testutil"
)

func TestAuditRepo_SaveAndFind(t *testing.T) {
	client := testutil.EmulatorClient(t, "AuditEvent")
	repo := NewRepository(client)
	ctx := context.Background()

	ts := time.Now().Truncate(time.Millisecond)
	evt := &model.AuditEvent{
		CustomerID:    "cust-audit-1",
		EventType:     "LOGIN",
		Timestamp:     ts,
		IPAddress:     "192.168.1.1",
		Details:       "successful login",
		SessionIDHash: "abc123hash",
	}

	saved, err := repo.Save(ctx, evt)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.ID == 0 {
		t.Fatal("expected non-zero ID after save")
	}

	events, err := repo.FindByCustomerID(ctx, "cust-audit-1")
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	got := events[0]
	if got.ID != saved.ID {
		t.Errorf("ID: got %d, want %d", got.ID, saved.ID)
	}
	if got.CustomerID != "cust-audit-1" {
		t.Errorf("CustomerID: got %q, want %q", got.CustomerID, "cust-audit-1")
	}
	if got.EventType != "LOGIN" {
		t.Errorf("EventType: got %q, want %q", got.EventType, "LOGIN")
	}
	if !got.Timestamp.Equal(ts) {
		t.Errorf("Timestamp: got %v, want %v", got.Timestamp, ts)
	}
	if got.IPAddress != "192.168.1.1" {
		t.Errorf("IPAddress: got %q, want %q", got.IPAddress, "192.168.1.1")
	}
	if got.Details != "successful login" {
		t.Errorf("Details: got %q, want %q", got.Details, "successful login")
	}
	if got.SessionIDHash != "abc123hash" {
		t.Errorf("SessionIDHash: got %q, want %q", got.SessionIDHash, "abc123hash")
	}
}

func TestAuditRepo_FindByCustomerID_Empty(t *testing.T) {
	client := testutil.EmulatorClient(t, "AuditEvent")
	repo := NewRepository(client)
	ctx := context.Background()

	events, err := repo.FindByCustomerID(ctx, "nonexistent-customer")
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestAuditRepo_Save_GeneratesID(t *testing.T) {
	client := testutil.EmulatorClient(t, "AuditEvent")
	repo := NewRepository(client)
	ctx := context.Background()

	evt1 := &model.AuditEvent{
		CustomerID: "cust-genid",
		EventType:  "ACTION_A",
		Timestamp:  time.Now().Truncate(time.Millisecond),
	}
	evt2 := &model.AuditEvent{
		CustomerID: "cust-genid",
		EventType:  "ACTION_B",
		Timestamp:  time.Now().Truncate(time.Millisecond),
	}

	saved1, err := repo.Save(ctx, evt1)
	if err != nil {
		t.Fatalf("Save evt1: %v", err)
	}
	saved2, err := repo.Save(ctx, evt2)
	if err != nil {
		t.Fatalf("Save evt2: %v", err)
	}

	if saved1.ID == 0 {
		t.Error("evt1 ID should be non-zero")
	}
	if saved2.ID == 0 {
		t.Error("evt2 ID should be non-zero")
	}
	if saved1.ID == saved2.ID {
		t.Errorf("expected different IDs, both got %d", saved1.ID)
	}
}
