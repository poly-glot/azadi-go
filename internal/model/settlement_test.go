package model

import (
	"testing"
	"time"
)

func TestNewSettlementResponse(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	validUntil := now.AddDate(0, 0, 30)

	tests := []struct {
		name       string
		figure     SettlementFigure
		wantAmount string
	}{
		{
			name: "typical settlement",
			figure: SettlementFigure{
				Base:         Base{ID: 1},
				AgreementID:  42,
				CustomerID:   "CUST-001",
				AmountPence:  1234567,
				CalculatedAt: now,
				ValidUntil:   validUntil,
			},
			wantAmount: "£12,345.67",
		},
		{
			name: "zero amount",
			figure: SettlementFigure{
				Base:         Base{ID: 2},
				AgreementID:  43,
				AmountPence:  0,
				CalculatedAt: now,
				ValidUntil:   validUntil,
			},
			wantAmount: "£0.00",
		},
		{
			name: "small amount",
			figure: SettlementFigure{
				Base:        Base{ID: 3},
				AmountPence: 99,
			},
			wantAmount: "£0.99",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSettlementResponse(&tt.figure)

			if got.ID != tt.figure.ID {
				t.Errorf("ID = %d, want %d", got.ID, tt.figure.ID)
			}
			if got.AgreementID != tt.figure.AgreementID {
				t.Errorf("AgreementID = %d, want %d", got.AgreementID, tt.figure.AgreementID)
			}
			if got.Amount != tt.wantAmount {
				t.Errorf("Amount = %q, want %q", got.Amount, tt.wantAmount)
			}
			if !got.CalculatedAt.Equal(tt.figure.CalculatedAt) {
				t.Errorf("CalculatedAt = %v, want %v", got.CalculatedAt, tt.figure.CalculatedAt)
			}
			if !got.ValidUntil.Equal(tt.figure.ValidUntil) {
				t.Errorf("ValidUntil = %v, want %v", got.ValidUntil, tt.figure.ValidUntil)
			}
		})
	}
}
