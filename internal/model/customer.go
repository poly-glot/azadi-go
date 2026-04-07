package model

import "time"

type Customer struct {
	Base
	CustomerID   string    `datastore:"customerId"`
	FullName     string    `datastore:"fullName"`
	Email        string    `datastore:"email"`
	DOB          time.Time `datastore:"dob"`
	Postcode     string    `datastore:"postcode"`
	Phone        string    `datastore:"phone"`
	MobilePhone  string    `datastore:"mobilePhone"`
	AddressLine1 string    `datastore:"addressLine1"`
	AddressLine2 string    `datastore:"addressLine2"`
	City         string    `datastore:"city"`
}
