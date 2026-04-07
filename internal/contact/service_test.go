package contact

import (
	"context"
	"errors"
	"testing"

	"azadi-go/internal/model"
)

type mockCustRepo struct {
	customer *model.Customer
	saved    *model.Customer
	findErr  error
	saveErr  error
}

func (m *mockCustRepo) FindByCustomerID(_ context.Context, _ string) (*model.Customer, error) {
	return m.customer, m.findErr
}

func (m *mockCustRepo) Save(_ context.Context, c *model.Customer) (*model.Customer, error) {
	m.saved = c
	if m.saveErr != nil {
		return nil, m.saveErr
	}
	c.ID = 1
	return c, nil
}

type mockContactAudit struct {
	lastEvent string
}

func (m *mockContactAudit) LogEvent(_, eventType, _, _ string, _ map[string]string) {
	m.lastEvent = eventType
}

func TestGetCustomer_Success(t *testing.T) {
	repo := &mockCustRepo{
		customer: &model.Customer{CustomerID: "CUST-001", Email: "test@test.com"},
	}
	svc := NewService(repo, &mockContactAudit{})

	got, err := svc.GetCustomer(context.Background(), "CUST-001")
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != "test@test.com" {
		t.Errorf("email = %q", got.Email)
	}
}

func TestGetCustomer_NotFound(t *testing.T) {
	repo := &mockCustRepo{customer: nil}
	svc := NewService(repo, &mockContactAudit{})

	_, err := svc.GetCustomer(context.Background(), "CUST-001")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetCustomer_RepoError(t *testing.T) {
	repo := &mockCustRepo{findErr: errors.New("db error")}
	svc := NewService(repo, &mockContactAudit{})

	_, err := svc.GetCustomer(context.Background(), "CUST-001")
	if err == nil {
		t.Error("expected error")
	}
}

func TestUpdateContactDetails_Success(t *testing.T) {
	customer := &model.Customer{
		Base: model.Base{ID: 1}, CustomerID: "CUST-001",
		Email: "old@test.com", Phone: "01234",
	}
	repo := &mockCustRepo{customer: customer}
	audit := &mockContactAudit{}
	svc := NewService(repo, audit)

	err := svc.UpdateContactDetails(context.Background(), "CUST-001",
		"07111", "07222", "new@test.com",
		"123 Main St", "", "London", "SW1A 1AA",
		"127.0.0.1", "sess-123")
	if err != nil {
		t.Fatal(err)
	}
	if repo.saved == nil {
		t.Fatal("expected customer to be saved")
	}
	if repo.saved.Email != "new@test.com" {
		t.Errorf("email = %q", repo.saved.Email)
	}
	if repo.saved.Phone != "07111" {
		t.Errorf("phone = %q", repo.saved.Phone)
	}
	if audit.lastEvent != "CONTACT_DETAILS_UPDATED" {
		t.Errorf("audit = %q", audit.lastEvent)
	}
}

func TestUpdateContactDetails_CustomerNotFound(t *testing.T) {
	repo := &mockCustRepo{customer: nil}
	svc := NewService(repo, &mockContactAudit{})

	err := svc.UpdateContactDetails(context.Background(), "CUST-001",
		"", "", "", "", "", "", "",
		"127.0.0.1", "sess")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateContactDetails_SaveError(t *testing.T) {
	repo := &mockCustRepo{
		customer: &model.Customer{Base: model.Base{ID: 1}, CustomerID: "CUST-001"},
		saveErr:  errors.New("save failed"),
	}
	svc := NewService(repo, &mockContactAudit{})

	err := svc.UpdateContactDetails(context.Background(), "CUST-001",
		"", "", "", "", "", "", "",
		"127.0.0.1", "sess")
	if err == nil {
		t.Error("expected error")
	}
}
