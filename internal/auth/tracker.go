package auth

import (
	"context"
	"sync"
	"time"
)

const (
	maxAttempts  = 5
	lockDuration = 30 * time.Minute
)

type loginAttempt struct {
	count     int
	lockUntil time.Time
	createdAt time.Time
}

type LoginAttemptTracker struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
}

func NewLoginAttemptTracker(ctx context.Context) *LoginAttemptTracker {
	t := &LoginAttemptTracker{
		attempts: make(map[string]*loginAttempt),
	}
	go t.cleanupLoop(ctx)
	return t
}

func (t *LoginAttemptTracker) IsBlocked(agreementNumber string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	a, ok := t.attempts[agreementNumber]
	if !ok {
		return false
	}
	return a.count >= maxAttempts && time.Now().Before(a.lockUntil)
}

func (t *LoginAttemptTracker) RecordFailure(agreementNumber string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	a, ok := t.attempts[agreementNumber]
	if !ok {
		a = &loginAttempt{createdAt: time.Now()}
		t.attempts[agreementNumber] = a
	}
	a.count++
	if a.count >= maxAttempts {
		a.lockUntil = time.Now().Add(lockDuration)
	}
}

func (t *LoginAttemptTracker) RecordSuccess(agreementNumber string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.attempts, agreementNumber)
}

func (t *LoginAttemptTracker) ClearAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.attempts = make(map[string]*loginAttempt)
}

func (t *LoginAttemptTracker) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.mu.Lock()
			now := time.Now()
			for key, a := range t.attempts {
				if a.count >= maxAttempts {
					if now.After(a.lockUntil) {
						delete(t.attempts, key)
					}
				} else if now.Sub(a.createdAt) > lockDuration {
					delete(t.attempts, key)
				}
			}
			t.mu.Unlock()
		}
	}
}
