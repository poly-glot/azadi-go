package auth

import (
	"context"
	"testing"
)

func TestLoginAttemptTracker_NotBlockedInitially(t *testing.T) {
	tracker := NewLoginAttemptTracker(context.Background())
	if tracker.IsBlocked("AGR-001") {
		t.Error("should not be blocked initially")
	}
}

func TestLoginAttemptTracker_BlockAfter5Failures(t *testing.T) {
	tracker := NewLoginAttemptTracker(context.Background())
	for i := 0; i < 5; i++ {
		tracker.RecordFailure("AGR-001")
	}
	if !tracker.IsBlocked("AGR-001") {
		t.Error("should be blocked after 5 failures")
	}
}

func TestLoginAttemptTracker_NotBlockedBefore5(t *testing.T) {
	tracker := NewLoginAttemptTracker(context.Background())
	for i := 0; i < 4; i++ {
		tracker.RecordFailure("AGR-001")
	}
	if tracker.IsBlocked("AGR-001") {
		t.Error("should not be blocked after only 4 failures")
	}
}

func TestLoginAttemptTracker_ClearOnSuccess(t *testing.T) {
	tracker := NewLoginAttemptTracker(context.Background())
	for i := 0; i < 3; i++ {
		tracker.RecordFailure("AGR-001")
	}
	tracker.RecordSuccess("AGR-001")
	for i := 0; i < 5; i++ {
		tracker.RecordFailure("AGR-001")
	}
	if !tracker.IsBlocked("AGR-001") {
		t.Error("should be blocked after reset + 5 failures")
	}
}

func TestLoginAttemptTracker_DifferentAgreements(t *testing.T) {
	tracker := NewLoginAttemptTracker(context.Background())
	for i := 0; i < 5; i++ {
		tracker.RecordFailure("AGR-001")
	}
	if tracker.IsBlocked("AGR-002") {
		t.Error("different agreement should not be blocked")
	}
}

func TestLoginAttemptTracker_ClearAll(t *testing.T) {
	tracker := NewLoginAttemptTracker(context.Background())
	for i := 0; i < 5; i++ {
		tracker.RecordFailure("AGR-001")
	}
	tracker.ClearAll()
	if tracker.IsBlocked("AGR-001") {
		t.Error("should not be blocked after ClearAll")
	}
}
