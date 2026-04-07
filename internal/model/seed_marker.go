package model

import "time"

type SeedMarker struct {
	Base
	SeededAt time.Time `datastore:"seededAt"`
}
