package document

import (
	"context"

	"azadi-go/internal/model"
	"azadi-go/internal/repo"

	"cloud.google.com/go/datastore"
)

type Repository struct {
	repo.Store[*model.Document]
}

func NewRepository(client *datastore.Client) *Repository {
	return &Repository{Store: repo.Store[*model.Document]{Client: client, Kind: "Document"}}
}

func (r *Repository) FindByCustomerID(ctx context.Context, customerID string) ([]*model.Document, error) {
	return r.FindAll(ctx, r.Query().FilterField("customerId", "=", customerID))
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*model.Document, error) {
	return r.Store.FindByID(ctx, id, &model.Document{})
}
