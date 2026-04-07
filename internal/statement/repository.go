package statement

import (
	"context"

	"azadi-go/internal/model"
	"azadi-go/internal/repo"

	"cloud.google.com/go/datastore"
)

type Repository struct {
	repo.Store[*model.StatementRequest]
}

func NewRepository(client *datastore.Client) *Repository {
	return &Repository{Store: repo.Store[*model.StatementRequest]{Client: client, Kind: "StatementRequest"}}
}

func (r *Repository) FindByCustomerID(ctx context.Context, customerID string) ([]*model.StatementRequest, error) {
	return r.FindAll(ctx, r.Query().FilterField("customerId", "=", customerID))
}
