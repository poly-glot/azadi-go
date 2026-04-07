package model

import "time"

const (
	StatementStatusPending = "PENDING"
	StatementStatusSent    = "SENT"
)

type StatementRequest struct {
	Base
	CustomerID  string    `datastore:"customerId"`
	AgreementID int64     `datastore:"agreementId"`
	Status      string    `datastore:"status"`
	RequestedAt time.Time `datastore:"requestedAt"`
	FulfilledAt time.Time `datastore:"fulfilledAt"`
}
