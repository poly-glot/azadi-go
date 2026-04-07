package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"azadi-go/internal/model"

	"cloud.google.com/go/datastore"
)

type Encryptor interface {
	Encrypt(plaintext string) (string, error)
}

type Seeder struct {
	client    *datastore.Client
	encryptor Encryptor
}

func NewSeeder(client *datastore.Client, encryptor Encryptor) *Seeder {
	return &Seeder{client: client, encryptor: encryptor}
}

type seedCustomer struct {
	CustomerID   string  `json:"customerId"`
	FullName     string  `json:"fullName"`
	Email        string  `json:"email"`
	DOB          string  `json:"dob"`
	Postcode     string  `json:"postcode"`
	Phone        string  `json:"phone"`
	AddressLine1 string  `json:"addressLine1"`
	AddressLine2 *string `json:"addressLine2"`
	City         string  `json:"city"`
	Agreement    struct {
		AgreementNumber         string  `json:"agreementNumber"`
		Type                    string  `json:"type"`
		BalancePence            int64   `json:"balancePence"`
		APR                     float64 `json:"apr"`
		TermMonths              int     `json:"termMonths"`
		ContractMileage         int     `json:"contractMileage"`
		ExcessPricePerMilePence int64   `json:"excessPricePerMilePence"`
		VehicleModel            string  `json:"vehicleModel"`
		Registration            string  `json:"registration"`
		MonthlyPaymentPence     int64   `json:"monthlyPaymentPence"`
		PaymentsRemaining       int     `json:"paymentsRemaining"`
	} `json:"agreement"`
	BankDetails struct {
		AccountNumber string `json:"accountNumber"`
		SortCode      string `json:"sortCode"`
	} `json:"bankDetails"`
	Documents []struct {
		Title         string `json:"title"`
		FileName      string `json:"fileName"`
		ContentType   string `json:"contentType"`
		FileSizeBytes int64  `json:"fileSizeBytes"`
	} `json:"documents"`
}

func (s *Seeder) Seed(ctx context.Context, seedFile string) error {
	// Check for existing seed marker
	q := datastore.NewQuery("SeedMarker").Limit(1)
	var markers []*model.SeedMarker
	if keys, err := s.client.GetAll(ctx, q, &markers); err == nil && len(keys) > 0 {
		slog.Info("seed data already exists, skipping")
		return nil
	}

	slog.Info("reading seed data", "file", seedFile)
	data, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("reading seed file: %w", err)
	}

	var customers []seedCustomer
	if err := json.Unmarshal(data, &customers); err != nil {
		return fmt.Errorf("parsing seed data: %w", err)
	}

	for _, c := range customers {
		if err := s.seedCustomer(ctx, &c); err != nil {
			return fmt.Errorf("seeding customer %s: %w", c.CustomerID, err)
		}
	}

	// Save seed marker
	marker := &model.SeedMarker{SeededAt: time.Now()}
	if _, err := s.client.Put(ctx, datastore.IncompleteKey("SeedMarker", nil), marker); err != nil {
		return fmt.Errorf("saving seed marker: %w", err)
	}

	slog.Info("seeded customers", "count", len(customers))
	return nil
}

func (s *Seeder) seedCustomer(ctx context.Context, sc *seedCustomer) error {
	dob, err := time.Parse("2006-01-02", sc.DOB)
	if err != nil {
		return fmt.Errorf("parsing DOB: %w", err)
	}

	customer := &model.Customer{
		CustomerID:   sc.CustomerID,
		FullName:     sc.FullName,
		Email:        sc.Email,
		DOB:          dob,
		Postcode:     sc.Postcode,
		Phone:        sc.Phone,
		AddressLine1: sc.AddressLine1,
		City:         sc.City,
	}
	if sc.AddressLine2 != nil {
		customer.AddressLine2 = *sc.AddressLine2
	}
	if _, err := s.client.Put(ctx, datastore.IncompleteKey("Customer", nil), customer); err != nil {
		return fmt.Errorf("saving customer: %w", err)
	}

	agr := &model.Agreement{
		AgreementNumber:         sc.Agreement.AgreementNumber,
		CustomerID:              sc.CustomerID,
		Type:                    sc.Agreement.Type,
		BalancePence:            sc.Agreement.BalancePence,
		APR:                     fmt.Sprintf("%.1f", sc.Agreement.APR),
		OriginalTermMonths:      sc.Agreement.TermMonths,
		ContractMileage:         sc.Agreement.ContractMileage,
		ExcessPricePerMilePence: sc.Agreement.ExcessPricePerMilePence,
		VehicleModel:            sc.Agreement.VehicleModel,
		Registration:            sc.Agreement.Registration,
		LastPaymentPence:        sc.Agreement.MonthlyPaymentPence,
		LastPaymentDate:         time.Now().AddDate(0, -1, 0),
		NextPaymentPence:        sc.Agreement.MonthlyPaymentPence,
		NextPaymentDate:         time.Now().AddDate(0, 0, 14),
		PaymentsRemaining:       sc.Agreement.PaymentsRemaining,
		FinalPaymentDate:        time.Now().AddDate(0, sc.Agreement.PaymentsRemaining, 0),
	}
	agrKey, err := s.client.Put(ctx, datastore.IncompleteKey("Agreement", nil), agr)
	if err != nil {
		return fmt.Errorf("saving agreement: %w", err)
	}

	// Seed payment history
	for i := 6; i >= 1; i-- {
		payment := &model.PaymentRecord{
			AgreementID:           agrKey.ID,
			CustomerID:            sc.CustomerID,
			AmountPence:           50000 + int64(i)*1000,
			StripePaymentIntentID: fmt.Sprintf("pi_seed_%s_%d", sc.CustomerID, i),
			Status:                model.PaymentStatusCompleted,
			CreatedAt:             time.Now().Add(-time.Duration(i) * 30 * 24 * time.Hour),
			CompletedAt:           time.Now().Add(-time.Duration(i)*30*24*time.Hour + time.Minute),
		}
		if _, err := s.client.Put(ctx, datastore.IncompleteKey("PaymentRecord", nil), payment); err != nil {
			return fmt.Errorf("saving payment: %w", err)
		}
	}

	// Seed documents
	for _, doc := range sc.Documents {
		d := &model.Document{
			CustomerID:    sc.CustomerID,
			Title:         doc.Title,
			FileName:      doc.FileName,
			ContentType:   doc.ContentType,
			StoragePath:   fmt.Sprintf("documents/%s/%s", sc.CustomerID, doc.FileName),
			FileSizeBytes: doc.FileSizeBytes,
			CreatedAt:     time.Now().Add(-180 * 24 * time.Hour),
		}
		if _, err := s.client.Put(ctx, datastore.IncompleteKey("Document", nil), d); err != nil {
			return fmt.Errorf("saving document: %w", err)
		}
	}

	// Seed settlement
	settlement := &model.SettlementFigure{
		AgreementID:  agrKey.ID,
		CustomerID:   sc.CustomerID,
		AmountPence:  sc.Agreement.BalancePence + int64(float64(sc.Agreement.BalancePence)*0.02),
		CalculatedAt: time.Now(),
		ValidUntil:   time.Now().AddDate(0, 0, 28),
	}
	if _, err := s.client.Put(ctx, datastore.IncompleteKey("SettlementFigure", nil), settlement); err != nil {
		return fmt.Errorf("saving settlement: %w", err)
	}

	// Seed bank details
	encAccount, err := s.encryptor.Encrypt(sc.BankDetails.AccountNumber)
	if err != nil {
		return fmt.Errorf("encrypting account: %w", err)
	}
	encSort, err := s.encryptor.Encrypt(sc.BankDetails.SortCode)
	if err != nil {
		return fmt.Errorf("encrypting sort code: %w", err)
	}
	bank := &model.BankDetails{
		CustomerID:             sc.CustomerID,
		AccountHolderName:      sc.FullName,
		EncryptedAccountNumber: encAccount,
		EncryptedSortCode:      encSort,
		LastFourAccount:        sc.BankDetails.AccountNumber[len(sc.BankDetails.AccountNumber)-4:],
		LastTwoSortCode:        sc.BankDetails.SortCode[len(sc.BankDetails.SortCode)-2:],
		UpdatedAt:              time.Now(),
	}
	if _, err := s.client.Put(ctx, datastore.IncompleteKey("BankDetails", nil), bank); err != nil {
		return fmt.Errorf("saving bank details: %w", err)
	}

	return nil
}
