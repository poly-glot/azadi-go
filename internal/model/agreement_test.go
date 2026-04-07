package model

import (
	"testing"
	"time"
)

func TestNewAgreementResponse(t *testing.T) {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	nextMonth := now.AddDate(0, 1, 0)
	finalDate := now.AddDate(2, 0, 0)

	tests := []struct {
		name      string
		agreement Agreement
		wantAPR   string
		wantBal   string
		wantExc   string
		wantLast  string
		wantNext  string
	}{
		{
			name: "typical agreement with APR",
			agreement: Agreement{
				Base:                    Base{ID: 42},
				AgreementNumber:         "AGR-001",
				CustomerID:              "CUST-001",
				Type:                    "PCP",
				BalancePence:            1500000,
				APR:                     "6.9",
				OriginalTermMonths:      48,
				ContractMileage:         40000,
				ExcessPricePerMilePence: 850,
				VehicleModel:            "Ford Focus",
				Registration:            "AB12 CDE",
				LastPaymentPence:        25000,
				LastPaymentDate:         now,
				NextPaymentPence:        25000,
				NextPaymentDate:         nextMonth,
				PaymentsRemaining:       24,
				FinalPaymentDate:        finalDate,
			},
			wantAPR:  "6.9%",
			wantBal:  "£15,000.00",
			wantExc:  "£8.50",
			wantLast: "£250.00",
			wantNext: "£250.00",
		},
		{
			name: "empty APR shows N/A",
			agreement: Agreement{
				APR: "",
			},
			wantAPR:  "N/A",
			wantBal:  "£0.00",
			wantExc:  "£0.00",
			wantLast: "£0.00",
			wantNext: "£0.00",
		},
		{
			name: "zero balance",
			agreement: Agreement{
				APR:          "0.0",
				BalancePence: 0,
			},
			wantAPR: "0.0%",
			wantBal: "£0.00",
			wantExc: "£0.00",
			wantLast: "£0.00",
			wantNext: "£0.00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAgreementResponse(&tt.agreement)

			if got.ID != tt.agreement.ID {
				t.Errorf("ID = %d, want %d", got.ID, tt.agreement.ID)
			}
			if got.AgreementNumber != tt.agreement.AgreementNumber {
				t.Errorf("AgreementNumber = %q, want %q", got.AgreementNumber, tt.agreement.AgreementNumber)
			}
			if got.Type != tt.agreement.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.agreement.Type)
			}
			if got.APR != tt.wantAPR {
				t.Errorf("APR = %q, want %q", got.APR, tt.wantAPR)
			}
			if got.Balance != tt.wantBal {
				t.Errorf("Balance = %q, want %q", got.Balance, tt.wantBal)
			}
			if got.ExcessPricePerMile != tt.wantExc {
				t.Errorf("ExcessPricePerMile = %q, want %q", got.ExcessPricePerMile, tt.wantExc)
			}
			if got.LastPayment != tt.wantLast {
				t.Errorf("LastPayment = %q, want %q", got.LastPayment, tt.wantLast)
			}
			if got.NextPayment != tt.wantNext {
				t.Errorf("NextPayment = %q, want %q", got.NextPayment, tt.wantNext)
			}
			if got.OriginalTermMonths != tt.agreement.OriginalTermMonths {
				t.Errorf("OriginalTermMonths = %d, want %d", got.OriginalTermMonths, tt.agreement.OriginalTermMonths)
			}
			if got.ContractMileage != tt.agreement.ContractMileage {
				t.Errorf("ContractMileage = %d, want %d", got.ContractMileage, tt.agreement.ContractMileage)
			}
			if got.VehicleModel != tt.agreement.VehicleModel {
				t.Errorf("VehicleModel = %q, want %q", got.VehicleModel, tt.agreement.VehicleModel)
			}
			if got.Registration != tt.agreement.Registration {
				t.Errorf("Registration = %q, want %q", got.Registration, tt.agreement.Registration)
			}
			if got.PaymentsRemaining != tt.agreement.PaymentsRemaining {
				t.Errorf("PaymentsRemaining = %d, want %d", got.PaymentsRemaining, tt.agreement.PaymentsRemaining)
			}
			if !got.LastPaymentDate.Equal(tt.agreement.LastPaymentDate) {
				t.Errorf("LastPaymentDate = %v, want %v", got.LastPaymentDate, tt.agreement.LastPaymentDate)
			}
			if !got.NextPaymentDate.Equal(tt.agreement.NextPaymentDate) {
				t.Errorf("NextPaymentDate = %v, want %v", got.NextPaymentDate, tt.agreement.NextPaymentDate)
			}
			if !got.FinalPaymentDate.Equal(tt.agreement.FinalPaymentDate) {
				t.Errorf("FinalPaymentDate = %v, want %v", got.FinalPaymentDate, tt.agreement.FinalPaymentDate)
			}
		})
	}
}
