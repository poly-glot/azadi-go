package payment

import (
	"fmt"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/webhook"
)

type StripeService struct {
	webhookSecret string
}

func NewStripeService(apiKey, webhookSecret string) *StripeService {
	stripe.Key = apiKey
	return &StripeService{webhookSecret: webhookSecret}
}

func (s *StripeService) CreatePaymentIntent(amountPence int64, agreementNumber, customerEmail string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amountPence),
		Currency:      stripe.String("gbp"),
		CaptureMethod: stripe.String("manual"),
	}
	params.AddMetadata("agreementNumber", agreementNumber)
	params.AddMetadata("customerEmail", customerEmail)
	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("creating payment intent: %w", err)
	}
	return pi, nil
}

func (s *StripeService) CapturePaymentIntent(paymentIntentID string) (*stripe.PaymentIntent, error) {
	pi, err := paymentintent.Capture(paymentIntentID, nil)
	if err != nil {
		return nil, fmt.Errorf("capturing payment intent: %w", err)
	}
	return pi, nil
}

func (s *StripeService) CancelPaymentIntent(paymentIntentID string) (*stripe.PaymentIntent, error) {
	pi, err := paymentintent.Cancel(paymentIntentID, nil)
	if err != nil {
		return nil, fmt.Errorf("canceling payment intent: %w", err)
	}
	return pi, nil
}

func (s *StripeService) ConstructWebhookEvent(payload []byte, sigHeader string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, sigHeader, s.webhookSecret)
}
