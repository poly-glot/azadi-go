// Package repo provides a generic Datastore repository that eliminates
// boilerplate CRUD operations across domain packages.
package repo

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/datastore"
)

// Entity is the constraint for any model stored in Datastore.
// Every model embeds model.Base which satisfies this via pointer receiver.
type Entity interface {
	GetID() int64
	SetID(int64)
}

// Store provides generic Datastore CRUD for any Entity type.
// T must be a pointer-to-struct that implements Entity (e.g. *model.Agreement).
type Store[T Entity] struct {
	Client *datastore.Client
	Kind   string
}

// Save persists an entity. New entities (ID == 0) get auto-generated IDs;
// existing entities are updated in place.
func (s *Store[T]) Save(ctx context.Context, e T) (T, error) {
	var key *datastore.Key
	if e.GetID() != 0 {
		key = datastore.IDKey(s.Kind, e.GetID(), nil)
	} else {
		key = datastore.IncompleteKey(s.Kind, nil)
	}
	k, err := s.Client.Put(ctx, key, e)
	if err != nil {
		return e, fmt.Errorf("saving %s: %w", s.Kind, err)
	}
	e.SetID(k.ID)
	return e, nil
}

// FindByID retrieves a single entity by Datastore key ID.
// Returns (zero, nil) when not found.
func (s *Store[T]) FindByID(ctx context.Context, id int64, dest T) (T, error) {
	key := datastore.IDKey(s.Kind, id, nil)
	if err := s.Client.Get(ctx, key, dest); err != nil {
		if errors.Is(err, datastore.ErrNoSuchEntity) {
			var zero T
			return zero, nil
		}
		return dest, fmt.Errorf("finding %s by ID: %w", s.Kind, err)
	}
	dest.SetID(key.ID)
	return dest, nil
}

// FindOne runs a query and returns the first match, or (zero, nil) if none.
func (s *Store[T]) FindOne(ctx context.Context, q *datastore.Query) (T, error) {
	q = q.Limit(1)
	var results []T
	keys, err := s.Client.GetAll(ctx, q, &results)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("querying %s: %w", s.Kind, err)
	}
	if len(results) == 0 {
		var zero T
		return zero, nil
	}
	results[0].SetID(keys[0].ID)
	return results[0], nil
}

// FindAll runs a query and returns all matching entities with IDs populated.
func (s *Store[T]) FindAll(ctx context.Context, q *datastore.Query) ([]T, error) {
	var results []T
	keys, err := s.Client.GetAll(ctx, q, &results)
	if err != nil {
		return nil, fmt.Errorf("querying %s: %w", s.Kind, err)
	}
	for i, k := range keys {
		results[i].SetID(k.ID)
	}
	return results, nil
}

// Query returns a new Datastore query for this Store's kind.
func (s *Store[T]) Query() *datastore.Query {
	return datastore.NewQuery(s.Kind)
}
