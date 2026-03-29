package topics

// Topic constants — single source of truth for all Kafka topic names.
// Both producer and consumer import from here; no magic strings anywhere.

const (
	// User lifecycle events (published by user-service)
	UserRegistered    = "user.registered"
	UserUpdated       = "user.updated"
	UserDeleted       = "user.deleted"
	UserStatusChanged = "user.status.changed"
	UserLoggedIn      = "user.logged_in"

	// Password events
	UserPasswordChanged = "user.password.changed"

	// Inbound commands (consumed by user-service from other services)
	UserSendWelcomeEmail  = "user.send_welcome_email"
	UserSendPasswordReset = "user.send_password_reset"
	UserSendStatusNotify  = "user.send_status_notify"
)
