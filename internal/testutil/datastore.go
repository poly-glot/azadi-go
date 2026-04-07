// Package testutil provides shared helpers for integration tests
// that need a Firestore Datastore emulator.
package testutil

import (
	"context"
	"fmt"
	"os"
	"testing"

	"cloud.google.com/go/datastore"
)

const (
	emulatorHostEnv = "DATASTORE_EMULATOR_HOST"
	projectID       = "demo-azadi"
)

// EmulatorClient returns a Datastore client connected to the emulator.
// It skips the test if the emulator is not running.
// It cleans up entities of the given kinds after the test.
func EmulatorClient(t *testing.T, kinds ...string) *datastore.Client {
	t.Helper()

	host := os.Getenv(emulatorHostEnv)
	if host == "" {
		t.Skipf("skipping: %s not set (run docker compose up firebase-emulator)", emulatorHostEnv)
	}

	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		t.Fatalf("creating datastore client: %v", err)
	}

	t.Cleanup(func() {
		cleanupKinds(client, kinds)
		client.Close()
	})

	return client
}

// cleanupKinds deletes all entities for the given kinds.
// Best-effort; logs but does not fail on errors.
func cleanupKinds(client *datastore.Client, kinds []string) {
	ctx := context.Background()
	for _, kind := range kinds {
		q := datastore.NewQuery(kind).KeysOnly()
		keys, err := client.GetAll(ctx, q, nil)
		if err != nil {
			fmt.Printf("testutil: cleanup query %s: %v\n", kind, err)
			continue
		}
		if len(keys) > 0 {
			if err := client.DeleteMulti(ctx, keys); err != nil {
				fmt.Printf("testutil: cleanup delete %s: %v\n", kind, err)
			}
		}
	}
}
