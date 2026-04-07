package model

import "time"

type SettlementFigure struct {
	Base
	AgreementID  int64     `datastore:"agreementId"`
	CustomerID   string    `datastore:"customerId"`
	AmountPence  int64     `datastore:"amountPence"`
	CalculatedAt time.Time `datastore:"calculatedAt"`
	ValidUntil   time.Time `datastore:"validUntil"`
}

type SettlementResponse struct {
	ID           int64
	AgreementID  int64
	Amount       string
	CalculatedAt time.Time
	ValidUntil   time.Time
}

func NewSettlementResponse(s *SettlementFigure) SettlementResponse {
	return SettlementResponse{
		ID:           s.ID,
		AgreementID:  s.AgreementID,
		Amount:       FormatPence(s.AmountPence),
		CalculatedAt: s.CalculatedAt,
		ValidUntil:   s.ValidUntil,
	}
}
