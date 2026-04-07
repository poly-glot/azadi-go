package bank

import (
	"context"

	"azadi-go/internal/model"
	"azadi-go/internal/repo"

	"cloud.google.com/go/datastore"
)

type Repository struct {
	repo.Store[*model.BankDetails]
}

func NewRepository(client *datastore.Client) *Repository {
	return &Repository{Store: repo.Store[*model.BankDetails]{Client: client, Kind: "BankDetails"}}
}

func (r *Repository) FindByCustomerID(ctx context.Context, customerID string) (*model.BankDetails, error) {
	return r.FindOne(ctx, r.Query().FilterField("customerId", "=", customerID))
}
