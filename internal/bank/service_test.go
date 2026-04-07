package bank

import (
	"context"
	"errors"
	"testing"

	"azadi-go/internal/model"
)

type mockBankRepo struct {
	details *model.BankDetails
	saved   *model.BankDetails
	findErr error
	saveErr error
}

func (m *mockBankRepo) FindByCustomerID(_ context.Context, _ string) (*model.BankDetails, error) {
	return m.details, m.findErr
}

func (m *mockBankRepo) Save(_ context.Context, b *model.BankDetails) (*model.BankDetails, error) {
	m.saved = b
	if m.saveErr != nil {
		return nil, m.saveErr
	}
	if b.ID == 0 {
		b.ID = 1
	}
	return b, nil
}

type mockBankEncryptor struct {
	result string
	err    error
}

func (m *mockBankEncryptor) Encrypt(_ string) (string, error) {
	return m.result, m.err
}

type mockBankAudit struct {
	lastEvent string
}

func (m *mockBankAudit) LogEvent(_, eventType, _, _ string, _ map[string]string) {
	m.lastEvent = eventType
}

type mockBankEmail struct {
	called bool
}

func (m *mockBankEmail) SendBankDetailsUpdated(_ string) {
	m.called = true
}

func TestGetBankDetails(t *testing.T) {
	repo := &mockBankRepo{
		details: &model.BankDetails{Base: model.Base{ID: 1}, LastFourAccount: "1234"},
	}
	svc := NewService(repo, &mockBankEncryptor{}, &mockBankAudit{}, &mockBankEmail{})

	got, err := svc.GetBankDetails(context.Background(), "CUST-001")
	if err != nil {
		t.Fatal(err)
	}
	if got.LastFourAccount != "1234" {
		t.Errorf("last4 = %q, want 1234", got.LastFourAccount)
	}
}

func TestService_UpdateBankDetails_Success(t *testing.T) {
	repo := &mockBankRepo{details: nil} // No existing details
	enc := &mockBankEncryptor{result: "encrypted"}
	audit := &mockBankAudit{}
	email := &mockBankEmail{}
	svc := NewService(repo, enc, audit, email)

	got, err := svc.UpdateBankDetails(context.Background(), "CUST-001",
		"John Smith", "12345678", "12-34-56", "127.0.0.1", "sess-123")
	if err != nil {
		t.Fatal(err)
	}
	if got.AccountHolderName != "John Smith" {
		t.Errorf("holder = %q", got.AccountHolderName)
	}
	if got.LastFourAccount != "5678" {
		t.Errorf("last4 = %q, want 5678", got.LastFourAccount)
	}
	if got.LastTwoSortCode != "56" {
		t.Errorf("last2 = %q, want 56", got.LastTwoSortCode)
	}
	if got.EncryptedAccountNumber != "encrypted" {
		t.Errorf("encrypted = %q", got.EncryptedAccountNumber)
	}
	if audit.lastEvent != "BANK_DETAILS_UPDATED" {
		t.Errorf("audit = %q", audit.lastEvent)
	}
	if !email.called {
		t.Error("expected email to be sent")
	}
}

func TestService_UpdateBankDetails_ExistingDetails(t *testing.T) {
	repo := &mockBankRepo{
		details: &model.BankDetails{Base: model.Base{ID: 42}, CustomerID: "CUST-001"},
	}
	svc := NewService(repo, &mockBankEncryptor{result: "enc"}, &mockBankAudit{}, &mockBankEmail{})

	got, err := svc.UpdateBankDetails(context.Background(), "CUST-001",
		"John Smith", "12345678", "12-34-56", "127.0.0.1", "sess-123")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 42 {
		t.Errorf("should update existing record, got ID = %d", got.ID)
	}
}

func TestService_UpdateBankDetails_FindError(t *testing.T) {
	repo := &mockBankRepo{findErr: errors.New("db error")}
	svc := NewService(repo, &mockBankEncryptor{}, &mockBankAudit{}, &mockBankEmail{})

	_, err := svc.UpdateBankDetails(context.Background(), "CUST-001",
		"John", "12345678", "12-34-56", "127.0.0.1", "sess")
	if err == nil {
		t.Error("expected error")
	}
}

func TestService_UpdateBankDetails_EncryptError(t *testing.T) {
	repo := &mockBankRepo{details: nil}
	enc := &mockBankEncryptor{err: errors.New("encrypt failed")}
	svc := NewService(repo, enc, &mockBankAudit{}, &mockBankEmail{})

	_, err := svc.UpdateBankDetails(context.Background(), "CUST-001",
		"John", "12345678", "12-34-56", "127.0.0.1", "sess")
	if err == nil {
		t.Error("expected error")
	}
}

func TestService_UpdateBankDetails_SaveError(t *testing.T) {
	repo := &mockBankRepo{details: nil, saveErr: errors.New("save failed")}
	svc := NewService(repo, &mockBankEncryptor{result: "enc"}, &mockBankAudit{}, &mockBankEmail{})

	_, err := svc.UpdateBankDetails(context.Background(), "CUST-001",
		"John", "12345678", "12-34-56", "127.0.0.1", "sess")
	if err == nil {
		t.Error("expected error")
	}
}
