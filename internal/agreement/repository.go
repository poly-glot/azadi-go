package agreement

import (
	"context"

	"azadi-go/internal/model"
	"azadi-go/internal/repo"

	"cloud.google.com/go/datastore"
)

type Repository struct {
	repo.Store[*model.Agreement]
}

func NewRepository(client *datastore.Client) *Repository {
	return &Repository{Store: repo.Store[*model.Agreement]{Client: client, Kind: "Agreement"}}
}

func (r *Repository) FindByCustomerID(ctx context.Context, customerID string) ([]*model.Agreement, error) {
	return r.FindAll(ctx, r.Query().FilterField("customerId", "=", customerID))
}

func (r *Repository) FindByAgreementNumber(ctx context.Context, agreementNumber string) (*model.Agreement, error) {
	return r.FindOne(ctx, r.Query().FilterField("agreementNumber", "=", agreementNumber))
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*model.Agreement, error) {
	return r.Store.FindByID(ctx, id, &model.Agreement{})
}
