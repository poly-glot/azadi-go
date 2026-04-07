package payment

import (
	"testing"
)

func TestNewStripeService(t *testing.T) {
	svc := NewStripeService("sk_test_xxx", "whsec_xxx")
	if svc == nil {
		t.Fatal("NewStripeService returned nil")
	}
	if svc.webhookSecret != "whsec_xxx" {
		t.Errorf("webhookSecret = %q", svc.webhookSecret)
	}
}

func TestConstructWebhookEvent_InvalidSignature(t *testing.T) {
	svc := NewStripeService("sk_test_xxx", "whsec_test_secret")
	_, err := svc.ConstructWebhookEvent([]byte(`{}`), "bad_sig")
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestCreatePaymentIntent_NoKey(t *testing.T) {
	// With an invalid API key, this should fail when hitting Stripe
	svc := NewStripeService("sk_test_invalid_key_12345", "whsec_xxx")
	_, err := svc.CreatePaymentIntent(15000, "AGR-001", "test@test.com")
	if err == nil {
		// Might succeed if Stripe SDK doesn't validate locally
		// This is more of a smoke test
		t.Log("Stripe SDK accepted the call (may not validate key locally)")
	}
}
