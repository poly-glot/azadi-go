// Package domain defines shared interfaces used across multiple packages.
// Interfaces live at the consumer boundary; these exist because 3+ consumers
// share the exact same contract.
package domain

import (
	"context"

	"azadi-go/internal/model"
)

// AuditLogger logs audit events asynchronously. Implemented by audit.Service.
type AuditLogger interface {
	LogEvent(customerID, eventType, ipAddress, sessionID string, details map[string]string)
}

// AgreementLister retrieves agreements for a customer. Implemented by agreement.Service.
type AgreementLister interface {
	GetAgreementsForCustomer(ctx context.Context, customerID string) ([]*model.Agreement, error)
}
