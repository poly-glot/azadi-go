package document

import (
	"context"
	"testing"
	"time"

	"azadi-go/internal/model"
	"azadi-go/internal/testutil"
)

func TestRepo_SaveAndFindByID(t *testing.T) {
	client := testutil.EmulatorClient(t, "Document")
	repo := NewRepository(client)
	ctx := context.Background()

	now := time.Now().Truncate(time.Millisecond)
	d := &model.Document{
		CustomerID:    "CUST-100",
		Title:         "Annual Statement",
		FileName:      "statement-2025.pdf",
		ContentType:   "application/pdf",
		StoragePath:   "docs/CUST-100/statement-2025.pdf",
		FileSizeBytes: 102400,
		CreatedAt:     now,
	}

	saved, err := repo.Save(ctx, d)
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
	if got.CustomerID != "CUST-100" {
		t.Errorf("CustomerID = %q, want %q", got.CustomerID, "CUST-100")
	}
	if got.Title != "Annual Statement" {
		t.Errorf("Title = %q, want %q", got.Title, "Annual Statement")
	}
	if got.FileName != "statement-2025.pdf" {
		t.Errorf("FileName = %q, want %q", got.FileName, "statement-2025.pdf")
	}
	if got.ContentType != "application/pdf" {
		t.Errorf("ContentType = %q, want %q", got.ContentType, "application/pdf")
	}
	if got.StoragePath != "docs/CUST-100/statement-2025.pdf" {
		t.Errorf("StoragePath = %q, want %q", got.StoragePath, "docs/CUST-100/statement-2025.pdf")
	}
	if got.FileSizeBytes != 102400 {
		t.Errorf("FileSizeBytes = %d, want %d", got.FileSizeBytes, 102400)
	}
	if !got.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, now)
	}
}

func TestRepo_FindByCustomerID(t *testing.T) {
	client := testutil.EmulatorClient(t, "Document")
	repo := NewRepository(client)
	ctx := context.Background()

	now := time.Now().Truncate(time.Millisecond)
	custID := "CUST-200"

	d1 := &model.Document{
		CustomerID:    custID,
		Title:         "Doc One",
		FileName:      "doc1.pdf",
		ContentType:   "application/pdf",
		StoragePath:   "docs/CUST-200/doc1.pdf",
		FileSizeBytes: 1024,
		CreatedAt:     now,
	}
	d2 := &model.Document{
		CustomerID:    custID,
		Title:         "Doc Two",
		FileName:      "doc2.png",
		ContentType:   "image/png",
		StoragePath:   "docs/CUST-200/doc2.png",
		FileSizeBytes: 2048,
		CreatedAt:     now,
	}

	if _, err := repo.Save(ctx, d1); err != nil {
		t.Fatalf("Save d1: %v", err)
	}
	if _, err := repo.Save(ctx, d2); err != nil {
		t.Fatalf("Save d2: %v", err)
	}

	results, err := repo.FindByCustomerID(ctx, custID)
	if err != nil {
		t.Fatalf("FindByCustomerID: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}

	titles := map[string]bool{}
	for _, r := range results {
		if r.ID == 0 {
			t.Error("expected non-zero ID")
		}
		titles[r.Title] = true
	}
	if !titles["Doc One"] || !titles["Doc Two"] {
		t.Errorf("expected Doc One and Doc Two in results, got %v", titles)
	}
}

func TestRepo_FindByID_NotFound(t *testing.T) {
	client := testutil.EmulatorClient(t, "Document")
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
