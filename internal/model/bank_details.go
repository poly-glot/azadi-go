package model

import "time"

type BankDetails struct {
	Base
	CustomerID             string    `datastore:"customerId"`
	AccountHolderName      string    `datastore:"accountHolderName"`
	EncryptedAccountNumber string    `datastore:"encryptedAccountNumber"`
	EncryptedSortCode      string    `datastore:"encryptedSortCode"`
	LastFourAccount        string    `datastore:"lastFourAccount"`
	LastTwoSortCode        string    `datastore:"lastTwoSortCode"`
	UpdatedAt              time.Time `datastore:"updatedAt"`
}
