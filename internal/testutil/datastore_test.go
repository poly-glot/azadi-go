package testutil

import (
	"context"
	"testing"

	"cloud.google.com/go/datastore"
)

func TestEmulatorClient_Smoke(t *testing.T) {
	client := EmulatorClient(t, "_smoke_test")

	type Ping struct{ Val string }
	key := datastore.IncompleteKey("_smoke_test", nil)
	k, err := client.Put(context.Background(), key, &Ping{Val: "ok"})
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if k.ID == 0 {
		t.Error("expected auto-generated ID")
	}

	var got Ping
	if err := client.Get(context.Background(), k, &got); err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Val != "ok" {
		t.Errorf("val = %q, want ok", got.Val)
	}
}
