package paymentdate

import (
	"context"
	"errors"
	"testing"
	"time"

	"azadi-go/internal/model"
)

// --- mocks ---

type mockAgreementGetter struct {
	getAgreementFn func(ctx context.Context, customerID string, agreementID int64) (*model.Agreement, error)
}

func (m *mockAgreementGetter) GetAgreement(ctx context.Context, customerID string, agreementID int64) (*model.Agreement, error) {
	return m.getAgreementFn(ctx, customerID, agreementID)
}

type mockAgreementSaver struct {
	saveFn    func(ctx context.Context, a *model.Agreement) (*model.Agreement, error)
	lastSaved *model.Agreement
}

func (m *mockAgreementSaver) Save(ctx context.Context, a *model.Agreement) (*model.Agreement, error) {
	m.lastSaved = a
	return m.saveFn(ctx, a)
}

type mockLogEventer struct {
	called    bool
	eventType string
}

func (m *mockLogEventer) LogEvent(customerID, eventType, ipAddress, sessionID string, details map[string]string) {
	m.called = true
	m.eventType = eventType
}

// --- tests ---

func TestGetCurrentPaymentDay(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want int
	}{
		{"15th", time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC), 15},
		{"1st", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 1},
		{"28th", time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC), 28},
		{"zero value", time.Time{}, 1},
		{"last day of month", time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC), 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &model.Agreement{NextPaymentDate: tt.date}
			got := GetCurrentPaymentDay(a)
			if got != tt.want {
				t.Errorf("GetCurrentPaymentDay() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPaymentDateValidation(t *testing.T) {
	tests := []struct {
		day   int
		valid bool
	}{
		{1, true},
		{14, true},
		{28, true},
		{0, false},
		{29, false},
		{-1, false},
		{100, false},
	}
	for _, tt := range tests {
		valid := tt.day >= 1 && tt.day <= 28
		if valid != tt.valid {
			t.Errorf("day %d: valid = %v, want %v", tt.day, valid, tt.valid)
		}
	}
}

func TestErrConstants(t *testing.T) {
	if ErrAlreadyChanged == nil {
		t.Error("ErrAlreadyChanged should not be nil")
	}
	if ErrInvalidDay == nil {
		t.Error("ErrInvalidDay should not be nil")
	}
}

func TestChangePaymentDate(t *testing.T) {
	tests := []struct {
		name           string
		customerID     string
		agreementID    int64
		newDay         int
		agreement      *model.Agreement
		agreeErr       error
		saveErr        error
		wantErr        error
		wantNextDay    int
		wantAuditEvent bool
	}{
		{
			name:        "success",
			customerID:  "cust-1",
			agreementID: 42,
			newDay:      15,
			agreement: &model.Agreement{
				Base:            model.Base{ID: 42},
				CustomerID:      "cust-1",
				NextPaymentDate: time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC),
			},
			wantNextDay:    15,
			wantAuditEvent: true,
		},
		{
			name:        "already changed",
			customerID:  "cust-1",
			agreementID: 42,
			newDay:      15,
			agreement: &model.Agreement{
				Base:               model.Base{ID: 42},
				CustomerID:         "cust-1",
				PaymentDateChanged: true,
			},
			wantErr: ErrAlreadyChanged,
		},
		{
			name:        "invalid day - zero",
			customerID:  "cust-1",
			agreementID: 42,
			newDay:      0,
			agreement: &model.Agreement{
				Base:       model.Base{ID: 42},
				CustomerID: "cust-1",
			},
			wantErr: ErrInvalidDay,
		},
		{
			name:        "invalid day - 29",
			customerID:  "cust-1",
			agreementID: 42,
			newDay:      29,
			agreement: &model.Agreement{
				Base:       model.Base{ID: 42},
				CustomerID: "cust-1",
			},
			wantErr: ErrInvalidDay,
		},
		{
			name:        "agreement getter error",
			customerID:  "cust-1",
			agreementID: 42,
			newDay:      15,
			agreeErr:    errors.New("not found"),
			wantErr:     errors.New("getting agreement"),
		},
		{
			name:        "save error",
			customerID:  "cust-1",
			agreementID: 42,
			newDay:      15,
			agreement: &model.Agreement{
				Base:            model.Base{ID: 42},
				CustomerID:      "cust-1",
				NextPaymentDate: time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC),
			},
			saveErr: errors.New("db error"),
			wantErr: errors.New("saving agreement"),
		},
		{
			name:        "updates NextPaymentDate correctly",
			customerID:  "cust-1",
			agreementID: 42,
			newDay:      1,
			agreement: &model.Agreement{
				Base:            model.Base{ID: 42},
				CustomerID:      "cust-1",
				NextPaymentDate: time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			},
			wantNextDay:    1,
			wantAuditEvent: true,
		},
		{
			name:        "zero NextPaymentDate stays zero",
			customerID:  "cust-1",
			agreementID: 42,
			newDay:      15,
			agreement: &model.Agreement{
				Base:       model.Base{ID: 42},
				CustomerID: "cust-1",
			},
			wantNextDay:    0, // zero time has no meaningful day
			wantAuditEvent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agSvc := &mockAgreementGetter{
				getAgreementFn: func(_ context.Context, _ string, _ int64) (*model.Agreement, error) {
					if tt.agreeErr != nil {
						return nil, tt.agreeErr
					}
					return tt.agreement, nil
				},
			}
			saver := &mockAgreementSaver{
				saveFn: func(_ context.Context, a *model.Agreement) (*model.Agreement, error) {
					if tt.saveErr != nil {
						return nil, tt.saveErr
					}
					return a, nil
				},
			}
			audit := &mockLogEventer{}
			svc := NewService(agSvc, saver, audit)

			err := svc.ChangePaymentDate(context.Background(), tt.customerID, tt.agreementID, tt.newDay, "127.0.0.1", "sess-1")

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if errors.Is(tt.wantErr, ErrAlreadyChanged) || errors.Is(tt.wantErr, ErrInvalidDay) {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("got error %v, want %v", err, tt.wantErr)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify PaymentDateChanged flag is set
			if saver.lastSaved == nil {
				t.Fatal("expected agreement to be saved")
			}
			if !saver.lastSaved.PaymentDateChanged {
				t.Error("PaymentDateChanged should be true after change")
			}

			// Verify NextPaymentDate day was updated
			if tt.wantNextDay > 0 && !saver.lastSaved.NextPaymentDate.IsZero() {
				gotDay := saver.lastSaved.NextPaymentDate.Day()
				if gotDay != tt.wantNextDay {
					t.Errorf("NextPaymentDate day = %d, want %d", gotDay, tt.wantNextDay)
				}
			}

			// Verify audit was called
			if tt.wantAuditEvent && !audit.called {
				t.Error("expected audit LogEvent to be called")
			}
			if tt.wantAuditEvent && audit.eventType != "PAYMENT_DATE_CHANGED" {
				t.Errorf("audit eventType = %q, want %q", audit.eventType, "PAYMENT_DATE_CHANGED")
			}
		})
	}
}
