package audit

import (
	"context"
	"sync"
	"testing"
	"time"

	"azadi-go/internal/model"
)

type mockAuditRepo struct {
	mu    sync.Mutex
	saved *model.AuditEvent
	err   error
}

func (m *mockAuditRepo) Save(_ context.Context, e *model.AuditEvent) (*model.AuditEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saved = e
	if m.err != nil {
		return nil, m.err
	}
	e.ID = 1
	return e, nil
}

func TestLogEvent(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := NewService(repo)

	svc.LogEvent("CUST-001", "PAYMENT_INITIATED", "127.0.0.1", "sess-123", map[string]string{
		"amount": "15000",
	})

	// LogEvent runs in goroutine, wait briefly
	time.Sleep(100 * time.Millisecond)

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.saved == nil {
		t.Fatal("expected audit event to be saved")
	}
	if repo.saved.CustomerID != "CUST-001" {
		t.Errorf("customerID = %q", repo.saved.CustomerID)
	}
	if repo.saved.EventType != "PAYMENT_INITIATED" {
		t.Errorf("eventType = %q", repo.saved.EventType)
	}
	if repo.saved.IPAddress != "127.0.0.1" {
		t.Errorf("ip = %q", repo.saved.IPAddress)
	}
	if repo.saved.SessionIDHash == "" || repo.saved.SessionIDHash == "no-session" {
		t.Errorf("sessionHash = %q, expected hash", repo.saved.SessionIDHash)
	}
}

func TestLogEvent_EmptySession(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := NewService(repo)

	svc.LogEvent("CUST-001", "LOGIN", "127.0.0.1", "", nil)
	time.Sleep(100 * time.Millisecond)

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.saved == nil {
		t.Fatal("expected event saved")
	}
	if repo.saved.SessionIDHash != "no-session" {
		t.Errorf("sessionHash = %q, want no-session", repo.saved.SessionIDHash)
	}
}

func TestHashSessionID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantHash string
	}{
		{"empty returns no-session", "", 0, "no-session"},
		{"non-empty returns hex hash", "sess-123", 64, ""},
		{"consistent hashing", "sess-123", 64, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hashSessionID(tt.input)
			if tt.input == "" {
				if got != "no-session" {
					t.Errorf("hashSessionID(%q) = %q, want %q", tt.input, got, "no-session")
				}
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("hashSessionID(%q) length = %d, want %d", tt.input, len(got), tt.wantLen)
			}
		})
	}

	// Consistency check
	a := hashSessionID("test-session")
	b := hashSessionID("test-session")
	if a != b {
		t.Error("hash should be deterministic")
	}

	// Different inputs produce different hashes
	c := hashSessionID("other-session")
	if a == c {
		t.Error("different inputs should produce different hashes")
	}
}

func TestNewService(t *testing.T) {
	svc := NewService(nil)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
}
