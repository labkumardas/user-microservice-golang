package kafka

import (
	"time"

	"github.com/google/uuid"
)

// EventEnvelope is the standard wrapper for every Kafka message.
// All services in the ecosystem must use this structure.
type EventEnvelope struct {
	EventID   string      `json:"event_id"`   // unique per message
	EventType string      `json:"event_type"` // matches topic constant
	Version   string      `json:"version"`    // schema version e.g. "v1"
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`  // originating service name
	Payload   interface{} `json:"payload"` // domain-specific data
}

// NewEvent constructs a ready-to-publish EventEnvelope
func NewEvent(eventType, source string, payload interface{}) EventEnvelope {
	return EventEnvelope{
		EventID:   uuid.New().String(),
		EventType: eventType,
		Version:   "v1",
		Timestamp: time.Now().UTC(),
		Source:    source,
		Payload:   payload,
	}
}

// ─── Payload types ────────────────────────────────────────────────────────────

// UserRegisteredPayload is published when a new user registers
type UserRegisteredPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UserUpdatedPayload is published when a user updates their profile
type UserUpdatedPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UserDeletedPayload is published when a user is soft-deleted
type UserDeletedPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// UserStatusChangedPayload is published when an admin changes user status
type UserStatusChangedPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	NewStatus string `json:"new_status"`
	OldStatus string `json:"old_status"`
}

// UserLoggedInPayload is published on successful login
type UserLoggedInPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// UserPasswordChangedPayload is published after a password change
type UserPasswordChangedPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// ─── Inbound command payloads (consumed) ─────────────────────────────────────

// WelcomeEmailPayload is consumed from notification-service
type WelcomeEmailPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// PasswordResetPayload is consumed from auth-service
type PasswordResetPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	ResetLink string `json:"reset_link"`
}

// StatusNotifyPayload is consumed from admin-service
type StatusNotifyPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	NewStatus string `json:"new_status"`
}
