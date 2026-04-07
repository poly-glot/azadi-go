package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"time"

	"azadi-go/internal/model"
)

type auditRepo interface {
	Save(ctx context.Context, e *model.AuditEvent) (*model.AuditEvent, error)
}

type Service struct {
	repo auditRepo
}

func NewService(repo auditRepo) *Service {
	return &Service{repo: repo}
}

func (s *Service) LogEvent(customerID, eventType, ipAddress, sessionID string, details map[string]string) {
	go func() {
		detailsJSON, err := json.Marshal(details)
		if err != nil {
			slog.Error("failed to marshal audit details", "error", err)
			return
		}
		event := &model.AuditEvent{
			CustomerID:    customerID,
			EventType:     eventType,
			Timestamp:     time.Now(),
			IPAddress:     ipAddress,
			Details:       string(detailsJSON),
			SessionIDHash: hashSessionID(sessionID),
		}
		ctx := context.Background()
		if _, err := s.repo.Save(ctx, event); err != nil {
			slog.Error("failed to write audit event", "eventType", eventType, "customerID", customerID, "error", err)
		}
	}()
}

func hashSessionID(sessionID string) string {
	if sessionID == "" {
		return "no-session"
	}
	h := sha256.Sum256([]byte(sessionID))
	return hex.EncodeToString(h[:])
}
