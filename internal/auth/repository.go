package auth

import (
	"context"

	"azadi-go/internal/model"
	"azadi-go/internal/repo"

	"cloud.google.com/go/datastore"
)

type CustomerRepository struct {
	repo.Store[*model.Customer]
}

func NewCustomerRepository(client *datastore.Client) *CustomerRepository {
	return &CustomerRepository{Store: repo.Store[*model.Customer]{Client: client, Kind: "Customer"}}
}

func (r *CustomerRepository) FindByCustomerID(ctx context.Context, customerID string) (*model.Customer, error) {
	return r.FindOne(ctx, r.Query().FilterField("customerId", "=", customerID))
}

func (r *CustomerRepository) FindAll(ctx context.Context) ([]*model.Customer, error) {
	return r.Store.FindAll(ctx, r.Query())
}
