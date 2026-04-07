package model

import (
	"strings"
	"time"
)

type Document struct {
	Base
	CustomerID    string    `datastore:"customerId"`
	Title         string    `datastore:"title"`
	FileName      string    `datastore:"fileName"`
	ContentType   string    `datastore:"contentType"`
	StoragePath   string    `datastore:"storagePath"`
	FileSizeBytes int64     `datastore:"fileSizeBytes"`
	CreatedAt     time.Time `datastore:"createdAt"`
}

func (d *Document) FileType() string {
	if i := strings.LastIndex(d.FileName, "."); i >= 0 {
		return strings.ToLower(d.FileName[i+1:])
	}
	return "pdf"
}
