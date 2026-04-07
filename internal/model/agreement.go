package model

import "time"

type Agreement struct {
	Base
	AgreementNumber         string    `datastore:"agreementNumber"`
	CustomerID              string    `datastore:"customerId"`
	Type                    string    `datastore:"type"`
	BalancePence            int64     `datastore:"balancePence"`
	APR                     string    `datastore:"apr"`
	OriginalTermMonths      int       `datastore:"originalTermMonths"`
	ContractMileage         int       `datastore:"contractMileage"`
	ExcessPricePerMilePence int64     `datastore:"excessPricePerMilePence"`
	VehicleModel            string    `datastore:"vehicleModel"`
	Registration            string    `datastore:"registration"`
	LastPaymentPence        int64     `datastore:"lastPaymentPence"`
	LastPaymentDate         time.Time `datastore:"lastPaymentDate"`
	NextPaymentPence        int64     `datastore:"nextPaymentPence"`
	NextPaymentDate         time.Time `datastore:"nextPaymentDate"`
	PaymentsRemaining       int       `datastore:"paymentsRemaining"`
	FinalPaymentDate        time.Time `datastore:"finalPaymentDate"`
	PaymentDateChanged      bool      `datastore:"paymentDateChanged"`
}

type AgreementResponse struct {
	ID                  int64
	AgreementNumber     string
	Type                string
	Balance             string
	APR                 string
	OriginalTermMonths  int
	ContractMileage     int
	ExcessPricePerMile  string
	VehicleModel        string
	Registration        string
	LastPayment         string
	LastPaymentDate     time.Time
	NextPayment         string
	NextPaymentDate     time.Time
	PaymentsRemaining   int
	FinalPaymentDate    time.Time
}

func NewAgreementResponse(a *Agreement) AgreementResponse {
	apr := "N/A"
	if a.APR != "" {
		apr = a.APR + "%"
	}
	return AgreementResponse{
		ID:                 a.ID,
		AgreementNumber:    a.AgreementNumber,
		Type:               a.Type,
		Balance:            FormatPence(a.BalancePence),
		APR:                apr,
		OriginalTermMonths: a.OriginalTermMonths,
		ContractMileage:    a.ContractMileage,
		ExcessPricePerMile: FormatPence(a.ExcessPricePerMilePence),
		VehicleModel:       a.VehicleModel,
		Registration:       a.Registration,
		LastPayment:        FormatPence(a.LastPaymentPence),
		LastPaymentDate:    a.LastPaymentDate,
		NextPayment:        FormatPence(a.NextPaymentPence),
		NextPaymentDate:    a.NextPaymentDate,
		PaymentsRemaining:  a.PaymentsRemaining,
		FinalPaymentDate:   a.FinalPaymentDate,
	}
}
