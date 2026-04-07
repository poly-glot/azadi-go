package payment

import (
	"context"
	"fmt"

	"azadi-go/internal/model"
	"azadi-go/internal/repo"

	"cloud.google.com/go/datastore"
)

type Repository struct {
	repo.Store[*model.PaymentRecord]
}

func NewRepository(client *datastore.Client) *Repository {
	return &Repository{Store: repo.Store[*model.PaymentRecord]{Client: client, Kind: "PaymentRecord"}}
}

func (r *Repository) FindByCustomerID(ctx context.Context, customerID string) ([]*model.PaymentRecord, error) {
	return r.FindAll(ctx, r.Query().FilterField("customerId", "=", customerID))
}

func (r *Repository) FindByStripePaymentIntentID(ctx context.Context, intentID string) (*model.PaymentRecord, error) {
	return r.FindOne(ctx, r.Query().FilterField("stripePaymentIntentId", "=", intentID))
}

func (r *Repository) FindByWebhookEventID(ctx context.Context, eventID string) (*model.PaymentRecord, error) {
	return r.FindOne(ctx, r.Query().FilterField("webhookEventId", "=", eventID))
}

func (r *Repository) SaveAll(ctx context.Context, records []*model.PaymentRecord) error {
	keys := make([]*datastore.Key, len(records))
	for i, rec := range records {
		if rec.ID != 0 {
			keys[i] = datastore.IDKey("PaymentRecord", rec.ID, nil)
		} else {
			keys[i] = datastore.IncompleteKey("PaymentRecord", nil)
		}
	}
	ks, err := r.Client.PutMulti(ctx, keys, records)
	if err != nil {
		return fmt.Errorf("saving payments: %w", err)
	}
	for i, k := range ks {
		records[i].ID = k.ID
	}
	return nil
}
