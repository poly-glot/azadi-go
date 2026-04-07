package email

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"azadi-go/internal/model"
)

type emailCustomerRepo interface {
	FindByCustomerID(ctx context.Context, customerID string) (*model.Customer, error)
}

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Service struct {
	apiKey       string
	fromEmail    string
	baseURL      string
	customerRepo emailCustomerRepo
	httpClient   httpDoer
}

func NewService(apiKey, fromEmail string, customerRepo emailCustomerRepo) *Service {
	return &Service{
		apiKey:       apiKey,
		fromEmail:    fromEmail,
		baseURL:      "https://api.resend.com",
		customerRepo: customerRepo,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// SendEmail sends an email asynchronously (fire-and-forget).
func (s *Service) SendEmail(to, subject, htmlBody string) {
	go s.sendEmail(to, subject, htmlBody)
}

// sendEmail sends an email synchronously. Call from a goroutine.
func (s *Service) sendEmail(to, subject, htmlBody string) {
	payload := map[string]any{
		"from":    s.fromEmail,
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal email payload", "error", err)
		return
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", s.baseURL+"/emails", bytes.NewReader(body))
	if err != nil {
		slog.Error("failed to create email request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		slog.Error("failed to send email", "error", err, "to", to)
		return
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		slog.Error("email API error", "status", resp.StatusCode, "to", to)
	} else {
		slog.Info("email sent", "to", to, "subject", subject)
	}
}

// resolveAndSend looks up the customer email and sends synchronously.
// Always call from a goroutine — never block a request handler.
func (s *Service) resolveAndSend(customerID, subject, htmlBody string) {
	customer, err := s.customerRepo.FindByCustomerID(context.Background(), customerID)
	if err != nil || customer == nil || customer.Email == "" {
		slog.Warn("cannot resolve email for customer", "customerID", customerID)
		return
	}
	s.sendEmail(customer.Email, subject, htmlBody)
}

func (s *Service) SendPaymentConfirmation(customerID string, amountPence int64) {
	go s.resolveAndSend(customerID, "Payment Confirmed - Azadi Finance", PaymentConfirmationHTML(amountPence))
}

func (s *Service) SendSettlementFigure(customerID string, amountPence int64, validUntil time.Time) {
	go s.resolveAndSend(customerID, "Settlement Figure - Azadi Finance", SettlementFigureHTML(amountPence, validUntil))
}

func (s *Service) SendBankDetailsUpdated(customerID string) {
	go s.resolveAndSend(customerID, "Bank Details Updated - Azadi Finance", BankDetailsUpdatedHTML())
}

func (s *Service) SendPaymentDateChanged(customerID, newDate string) {
	go s.resolveAndSend(customerID, "Payment Date Changed - Azadi Finance", PaymentDateChangedHTML(newDate))
}

func (s *Service) SendLoginAlert(customerID, ipAddress string) {
	go s.resolveAndSend(customerID, "Login Alert - Azadi Finance", LoginAlertHTML(ipAddress))
}
