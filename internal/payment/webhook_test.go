package payment

import (
	"context"
	"encoding/json"
	"testing"

	"errors"

	"azadi-go/internal/model"

	"github.com/stripe/stripe-go/v82"
)

type mockWebhookRepo struct {
	byEventID  *model.PaymentRecord
	byIntentID *model.PaymentRecord
	saved      *model.PaymentRecord
	saveErr    error
}

func (m *mockWebhookRepo) FindByWebhookEventID(_ context.Context, _ string) (*model.PaymentRecord, error) {
	return m.byEventID, nil
}

func (m *mockWebhookRepo) FindByStripePaymentIntentID(_ context.Context, _ string) (*model.PaymentRecord, error) {
	return m.byIntentID, nil
}

func (m *mockWebhookRepo) Save(_ context.Context, p *model.PaymentRecord) (*model.PaymentRecord, error) {
	m.saved = p
	if m.saveErr != nil {
		return nil, m.saveErr
	}
	return p, nil
}

type mockWebhookAudit struct {
	lastEvent string
}

func (m *mockWebhookAudit) LogEvent(_, eventType, _, _ string, _ map[string]string) {
	m.lastEvent = eventType
}

type mockWebhookStripe struct {
	event stripe.Event
	err   error
}

func (m *mockWebhookStripe) ConstructWebhookEvent(_ []byte, _ string) (stripe.Event, error) {
	return m.event, m.err
}

type mockEmailService struct {
	called bool
}

func (m *mockEmailService) SendPaymentConfirmation(_ string, _ int64) {
	m.called = true
}

func makeWebhookEvent(t *testing.T, eventType, piID, eventID string) stripe.Event {
	t.Helper()
	pi := stripe.PaymentIntent{ID: piID}
	raw, _ := json.Marshal(pi)
	return stripe.Event{
		ID:   eventID,
		Type: stripe.EventType(eventType),
		Data: &stripe.EventData{Raw: json.RawMessage(raw)},
	}
}

func TestHandleEvent_PaymentSucceeded(t *testing.T) {
	repo := &mockWebhookRepo{
		byIntentID: &model.PaymentRecord{
			Base: model.Base{ID: 1}, CustomerID: "CUST-001", AmountPence: 15000,
			Status: model.PaymentStatusPending,
		},
	}
	audit := &mockWebhookAudit{}
	email := &mockEmailService{}
	stripeSvc := &mockWebhookStripe{
		event: makeWebhookEvent(t, "payment_intent.succeeded", "pi_test", "evt_001"),
	}
	wh := NewWebhookHandler(repo, audit, stripeSvc, email)

	status, _ := wh.HandleEvent([]byte(`{}`), "sig", "127.0.0.1")
	if status != 200 {
		t.Errorf("status = %d, want 200", status)
	}
	if repo.saved == nil {
		t.Fatal("expected payment record to be saved")
	}
	if repo.saved.Status != model.PaymentStatusCompleted {
		t.Errorf("status = %q, want COMPLETED", repo.saved.Status)
	}
	if audit.lastEvent != "PAYMENT_COMPLETED" {
		t.Errorf("audit event = %q, want PAYMENT_COMPLETED", audit.lastEvent)
	}
	if !email.called {
		t.Error("expected email to be sent")
	}
}

func TestHandleEvent_PaymentFailed(t *testing.T) {
	repo := &mockWebhookRepo{
		byIntentID: &model.PaymentRecord{
			Base: model.Base{ID: 1}, CustomerID: "CUST-001", AmountPence: 15000,
			Status: model.PaymentStatusPending,
		},
	}
	audit := &mockWebhookAudit{}
	stripeSvc := &mockWebhookStripe{
		event: makeWebhookEvent(t, "payment_intent.payment_failed", "pi_test", "evt_002"),
	}
	wh := NewWebhookHandler(repo, audit, stripeSvc, &mockEmailService{})

	status, _ := wh.HandleEvent([]byte(`{}`), "sig", "127.0.0.1")
	if status != 200 {
		t.Errorf("status = %d, want 200", status)
	}
	if repo.saved == nil {
		t.Fatal("expected payment record to be saved")
	}
	if repo.saved.Status != model.PaymentStatusFailed {
		t.Errorf("status = %q, want FAILED", repo.saved.Status)
	}
	if audit.lastEvent != "PAYMENT_FAILED" {
		t.Errorf("audit event = %q, want PAYMENT_FAILED", audit.lastEvent)
	}
}

func TestHandleEvent_DuplicateWebhook(t *testing.T) {
	repo := &mockWebhookRepo{
		byEventID:  &model.PaymentRecord{Base: model.Base{ID: 1}, WebhookEventID: "evt_001"}, // Already processed
		byIntentID: &model.PaymentRecord{Base: model.Base{ID: 1}},
	}
	stripeSvc := &mockWebhookStripe{
		event: makeWebhookEvent(t, "payment_intent.succeeded", "pi_test", "evt_001"),
	}
	wh := NewWebhookHandler(repo, &mockWebhookAudit{}, stripeSvc, &mockEmailService{})

	status, _ := wh.HandleEvent([]byte(`{}`), "sig", "127.0.0.1")
	if status != 200 {
		t.Errorf("status = %d, want 200", status)
	}
	if repo.saved != nil {
		t.Error("should not save duplicate webhook")
	}
}

func TestHandleEvent_InvalidSignature(t *testing.T) {
	stripeSvc := &mockWebhookStripe{err: errors.New("invalid signature")}
	wh := NewWebhookHandler(&mockWebhookRepo{}, &mockWebhookAudit{}, stripeSvc, &mockEmailService{})

	status, msg := wh.HandleEvent([]byte(`{}`), "bad_sig", "127.0.0.1")
	if status != 400 {
		t.Errorf("status = %d, want 400", status)
	}
	if msg != "Invalid signature" {
		t.Errorf("msg = %q, want %q", msg, "Invalid signature")
	}
}

func TestHandleEvent_UnhandledType(t *testing.T) {
	stripeSvc := &mockWebhookStripe{
		event: stripe.Event{ID: "evt_999", Type: stripe.EventType("unknown.event"), Data: &stripe.EventData{Raw: json.RawMessage(`{}`)}},
	}
	wh := NewWebhookHandler(&mockWebhookRepo{}, &mockWebhookAudit{}, stripeSvc, &mockEmailService{})

	status, _ := wh.HandleEvent([]byte(`{}`), "sig", "127.0.0.1")
	if status != 200 {
		t.Errorf("status = %d, want 200", status)
	}
}

func TestHandleEvent_NoPaymentRecord(t *testing.T) {
	repo := &mockWebhookRepo{
		byIntentID: nil, // No matching record
	}
	stripeSvc := &mockWebhookStripe{
		event: makeWebhookEvent(t, "payment_intent.succeeded", "pi_unknown", "evt_003"),
	}
	wh := NewWebhookHandler(repo, &mockWebhookAudit{}, stripeSvc, &mockEmailService{})

	status, _ := wh.HandleEvent([]byte(`{}`), "sig", "127.0.0.1")
	if status != 200 {
		t.Errorf("status = %d, want 200", status)
	}
	if repo.saved != nil {
		t.Error("should not save when no payment record found")
	}
}
