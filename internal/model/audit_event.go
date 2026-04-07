package model

import "time"

type AuditEvent struct {
	Base
	CustomerID    string    `datastore:"customerId"`
	EventType     string    `datastore:"eventType"`
	Timestamp     time.Time `datastore:"timestamp"`
	IPAddress     string    `datastore:"ipAddress"`
	Details       string    `datastore:"details"`
	SessionIDHash string    `datastore:"sessionIdHash"`
}
