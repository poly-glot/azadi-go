package payment

import (
	"context"
	"errors"
	"testing"

	"azadi-go/internal/model"

	"github.com/stripe/stripe-go/v82"
)

type mockPayStripe struct {
	pi  *stripe.PaymentIntent
	err error
}

func (m *mockPayStripe) CreatePaymentIntent(_ int64, _, _ string) (*stripe.PaymentIntent, error) {
	return m.pi, m.err
}

type mockPayRepo struct {
	saved *model.PaymentRecord
	err   error
}

func (m *mockPayRepo) Save(_ context.Context, p *model.PaymentRecord) (*model.PaymentRecord, error) {
	m.saved = p
	if m.err != nil {
		return nil, m.err
	}
	p.ID = 1
	return p, nil
}

type mockPayAudit struct {
	lastEvent string
}

func (m *mockPayAudit) LogEvent(_, eventType, _, _ string, _ map[string]string) {
	m.lastEvent = eventType
}

type mockPayCustomerRepo struct {
	customer *model.Customer
	err      error
}

func (m *mockPayCustomerRepo) FindByCustomerID(_ context.Context, _ string) (*model.Customer, error) {
	return m.customer, m.err
}

type mockPayAgreementSvc struct {
	agreement *model.Agreement
	err       error
}

func (m *mockPayAgreementSvc) GetAgreement(_ context.Context, _ string, _ int64) (*model.Agreement, error) {
	return m.agreement, m.err
}

func TestInitiatePayment_Success(t *testing.T) {
	stripeSvc := &mockPayStripe{
		pi: &stripe.PaymentIntent{ID: "pi_test", ClientSecret: "pi_test_secret"},
	}
	repo := &mockPayRepo{}
	audit := &mockPayAudit{}
	custRepo := &mockPayCustomerRepo{
		customer: &model.Customer{Email: "test@test.com"},
	}
	agrSvc := &mockPayAgreementSvc{
		agreement: &model.Agreement{Base: model.Base{ID: 1}, AgreementNumber: "AGR-001"},
	}

	svc := NewService(stripeSvc, repo, audit, custRepo, agrSvc)
	pi, err := svc.InitiatePayment(context.Background(), "CUST-001", 1, 15000, "127.0.0.1", "sess-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pi.ID != "pi_test" {
		t.Errorf("PI ID = %q, want pi_test", pi.ID)
	}
	if repo.saved == nil {
		t.Fatal("expected payment record to be saved")
	}
	if repo.saved.Status != model.PaymentStatusPending {
		t.Errorf("status = %q, want PENDING", repo.saved.Status)
	}
	if repo.saved.AmountPence != 15000 {
		t.Errorf("amount = %d, want 15000", repo.saved.AmountPence)
	}
	if audit.lastEvent != "PAYMENT_INITIATED" {
		t.Errorf("audit event = %q, want PAYMENT_INITIATED", audit.lastEvent)
	}
}

func TestInitiatePayment_AgreementError(t *testing.T) {
	agrSvc := &mockPayAgreementSvc{err: errors.New("not found")}
	svc := NewService(&mockPayStripe{}, &mockPayRepo{}, &mockPayAudit{}, &mockPayCustomerRepo{}, agrSvc)

	_, err := svc.InitiatePayment(context.Background(), "CUST-001", 99, 15000, "127.0.0.1", "sess-123")
	if err == nil {
		t.Error("expected error")
	}
}

func TestInitiatePayment_StripeError(t *testing.T) {
	agrSvc := &mockPayAgreementSvc{
		agreement: &model.Agreement{Base: model.Base{ID: 1}, AgreementNumber: "AGR-001"},
	}
	stripeSvc := &mockPayStripe{err: errors.New("stripe error")}
	svc := NewService(stripeSvc, &mockPayRepo{}, &mockPayAudit{}, &mockPayCustomerRepo{}, agrSvc)

	_, err := svc.InitiatePayment(context.Background(), "CUST-001", 1, 15000, "127.0.0.1", "sess-123")
	if err == nil {
		t.Error("expected error")
	}
}

func TestInitiatePayment_NoCustomerEmail(t *testing.T) {
	stripeSvc := &mockPayStripe{
		pi: &stripe.PaymentIntent{ID: "pi_test", ClientSecret: "pi_test_secret"},
	}
	custRepo := &mockPayCustomerRepo{err: errors.New("not found")} // No email
	agrSvc := &mockPayAgreementSvc{
		agreement: &model.Agreement{Base: model.Base{ID: 1}, AgreementNumber: "AGR-001"},
	}

	svc := NewService(stripeSvc, &mockPayRepo{}, &mockPayAudit{}, custRepo, agrSvc)
	pi, err := svc.InitiatePayment(context.Background(), "CUST-001", 1, 15000, "127.0.0.1", "sess-123")
	if err != nil {
		t.Fatalf("should succeed even without email: %v", err)
	}
	if pi.ID != "pi_test" {
		t.Errorf("PI ID = %q, want pi_test", pi.ID)
	}
}

func TestInitiatePayment_SaveError(t *testing.T) {
	stripeSvc := &mockPayStripe{
		pi: &stripe.PaymentIntent{ID: "pi_test", ClientSecret: "pi_test_secret"},
	}
	repo := &mockPayRepo{err: errors.New("save failed")}
	agrSvc := &mockPayAgreementSvc{
		agreement: &model.Agreement{Base: model.Base{ID: 1}, AgreementNumber: "AGR-001"},
	}

	svc := NewService(stripeSvc, repo, &mockPayAudit{}, &mockPayCustomerRepo{}, agrSvc)
	_, err := svc.InitiatePayment(context.Background(), "CUST-001", 1, 15000, "127.0.0.1", "sess-123")
	if err == nil {
		t.Error("expected error when save fails")
	}
}
