package audit

import (
	"context"

	"azadi-go/internal/model"
	"azadi-go/internal/repo"

	"cloud.google.com/go/datastore"
)

type Repository struct {
	repo.Store[*model.AuditEvent]
}

func NewRepository(client *datastore.Client) *Repository {
	return &Repository{Store: repo.Store[*model.AuditEvent]{Client: client, Kind: "AuditEvent"}}
}

func (r *Repository) FindByCustomerID(ctx context.Context, customerID string) ([]*model.AuditEvent, error) {
	return r.FindAll(ctx, r.Query().FilterField("customerId", "=", customerID))
}
