package model

import "azadi-go/internal/repo"

// Base provides the common Datastore ID field for all entities.
// Embed it as the first field of every model struct.
// The pointer receiver on GetID/SetID means T must be a pointer type
// when used with repo.Store[T].
type Base struct {
	ID int64 `datastore:"-"`
}

func (b *Base) GetID() int64  { return b.ID }
func (b *Base) SetID(id int64) { b.ID = id }

// Compile-time assertions that all models satisfy repo.Entity.
var (
	_ repo.Entity = (*Agreement)(nil)
	_ repo.Entity = (*Customer)(nil)
	_ repo.Entity = (*PaymentRecord)(nil)
	_ repo.Entity = (*BankDetails)(nil)
	_ repo.Entity = (*SettlementFigure)(nil)
	_ repo.Entity = (*StatementRequest)(nil)
	_ repo.Entity = (*Document)(nil)
	_ repo.Entity = (*AuditEvent)(nil)
	_ repo.Entity = (*SeedMarker)(nil)
)
