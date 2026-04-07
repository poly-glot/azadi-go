package model

import "time"

const (
	PaymentStatusPending   = "PENDING"
	PaymentStatusCompleted = "COMPLETED"
	PaymentStatusFailed    = "FAILED"
)

type PaymentRecord struct {
	Base
	AgreementID           int64     `datastore:"agreementId"`
	CustomerID            string    `datastore:"customerId"`
	AmountPence           int64     `datastore:"amountPence"`
	StripePaymentIntentID string    `datastore:"stripePaymentIntentId"`
	Status                string    `datastore:"status"`
	CreatedAt             time.Time `datastore:"createdAt"`
	CompletedAt           time.Time `datastore:"completedAt"`
	WebhookEventID        string    `datastore:"webhookEventId"`
}
