package seed

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"azadi-go/internal/testutil"

	"cloud.google.com/go/datastore"
)

type mockEncryptor struct{}

func (m *mockEncryptor) Encrypt(plaintext string) (string, error) {
	return "enc:" + plaintext, nil
}

func TestSeeder_Seed(t *testing.T) {
	client := testutil.EmulatorClient(t,
		"Customer", "Agreement", "PaymentRecord", "Document",
		"SettlementFigure", "BankDetails", "SeedMarker",
	)

	seedFile := findSeedFile(t)
	seeder := NewSeeder(client, &mockEncryptor{})
	ctx := context.Background()

	// First seed should succeed
	if err := seeder.Seed(ctx, seedFile); err != nil {
		t.Fatalf("Seed: %v", err)
	}

	// Verify customers created
	q := datastore.NewQuery("Customer").KeysOnly()
	keys, err := client.GetAll(ctx, q, nil)
	if err != nil {
		t.Fatalf("query customers: %v", err)
	}
	if len(keys) < 1 {
		t.Error("expected at least 1 customer seeded")
	}

	// Verify agreements created
	q = datastore.NewQuery("Agreement").KeysOnly()
	keys, err = client.GetAll(ctx, q, nil)
	if err != nil {
		t.Fatalf("query agreements: %v", err)
	}
	if len(keys) < 1 {
		t.Error("expected at least 1 agreement seeded")
	}

	// Verify seed marker created
	q = datastore.NewQuery("SeedMarker").KeysOnly().Limit(1)
	markerKeys, err := client.GetAll(ctx, q, nil)
	if err != nil {
		t.Fatalf("query seed marker: %v", err)
	}
	if len(markerKeys) == 0 {
		t.Error("expected seed marker to be created")
	}
}

func TestSeeder_Seed_Idempotent(t *testing.T) {
	client := testutil.EmulatorClient(t,
		"Customer", "Agreement", "PaymentRecord", "Document",
		"SettlementFigure", "BankDetails", "SeedMarker",
	)

	seedFile := findSeedFile(t)
	seeder := NewSeeder(client, &mockEncryptor{})
	ctx := context.Background()

	// Seed twice — second call should be a no-op
	if err := seeder.Seed(ctx, seedFile); err != nil {
		t.Fatalf("first Seed: %v", err)
	}

	// Count customers after first seed
	q := datastore.NewQuery("Customer").KeysOnly()
	keys1, _ := client.GetAll(ctx, q, nil)

	if err := seeder.Seed(ctx, seedFile); err != nil {
		t.Fatalf("second Seed: %v", err)
	}

	keys2, _ := client.GetAll(ctx, q, nil)
	if len(keys2) != len(keys1) {
		t.Errorf("second seed changed customer count: %d -> %d", len(keys1), len(keys2))
	}
}

func TestSeeder_Seed_BadFile(t *testing.T) {
	client := testutil.EmulatorClient(t, "SeedMarker")
	seeder := NewSeeder(client, &mockEncryptor{})
	err := seeder.Seed(context.Background(), "/nonexistent/file.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestSeeder_Seed_BadJSON(t *testing.T) {
	client := testutil.EmulatorClient(t, "SeedMarker")
	seeder := NewSeeder(client, &mockEncryptor{})

	tmp := t.TempDir()
	f := filepath.Join(tmp, "bad.json")
	os.WriteFile(f, []byte("not json"), 0644)

	err := seeder.Seed(context.Background(), f)
	if err == nil {
		t.Error("expected error for bad JSON")
	}
}

func TestNewSeeder(t *testing.T) {
	s := NewSeeder(nil, &mockEncryptor{})
	if s == nil {
		t.Fatal("NewSeeder returned nil")
	}
}

func findSeedFile(t *testing.T) string {
	t.Helper()
	// Walk up from test location to find seed/customers.json
	candidates := []string{
		"../../seed/customers.json",
		"../../../seed/customers.json",
		filepath.Join(os.Getenv("PWD"), "seed/customers.json"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	t.Skip("seed/customers.json not found")
	return ""
}
