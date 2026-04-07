package settlement

import (
	"context"

	"azadi-go/internal/model"
	"azadi-go/internal/repo"

	"cloud.google.com/go/datastore"
)

type Repository struct {
	repo.Store[*model.SettlementFigure]
}

func NewRepository(client *datastore.Client) *Repository {
	return &Repository{Store: repo.Store[*model.SettlementFigure]{Client: client, Kind: "SettlementFigure"}}
}

func (r *Repository) FindByCustomerID(ctx context.Context, customerID string) ([]*model.SettlementFigure, error) {
	return r.FindAll(ctx, r.Query().FilterField("customerId", "=", customerID))
}
