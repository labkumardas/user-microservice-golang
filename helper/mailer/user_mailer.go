package mailer

import "fmt"

// UserMailer provides high-level, domain-specific email methods for user events.
// It wraps Mailer and pre-fills templates so callers never touch raw HTML.
type UserMailer struct {
	mailer  *Mailer
	appName string
	support string // support email address
}

// NewUserMailer constructs a UserMailer
func NewUserMailer(m *Mailer, appName, supportEmail string) *UserMailer {
	return &UserMailer{
		mailer:  m,
		appName: appName,
		support: supportEmail,
	}
}

// SendWelcome sends a welcome email after successful registration
func (u *UserMailer) SendWelcome(toEmail, name, loginURL string) error {
	body, err := renderTemplate(welcomeHTML, WelcomeData{
		Name:     name,
		AppName:  u.appName,
		LoginURL: loginURL,
	})
	if err != nil {
		return fmt.Errorf("SendWelcome: %w", err)
	}

	return u.mailer.Send(&Message{
		To:      []string{toEmail},
		Subject: fmt.Sprintf("Welcome to %s! 🎉", u.appName),
		Body:    body,
	})
}

// SendPasswordReset sends a password reset link email
func (u *UserMailer) SendPasswordReset(toEmail, name, resetURL string, expiryMin int) error {
	body, err := renderTemplate(passwordResetHTML, PasswordResetData{
		Name:      name,
		AppName:   u.appName,
		ResetURL:  resetURL,
		ExpiryMin: expiryMin,
	})
	if err != nil {
		return fmt.Errorf("SendPasswordReset: %w", err)
	}

	return u.mailer.Send(&Message{
		To:      []string{toEmail},
		Subject: fmt.Sprintf("[%s] Password Reset Request", u.appName),
		Body:    body,
	})
}

// SendPasswordChanged notifies user that their password was changed
func (u *UserMailer) SendPasswordChanged(toEmail, name string) error {
	body, err := renderTemplate(passwordChangedHTML, PasswordChangedData{
		Name:         name,
		AppName:      u.appName,
		SupportEmail: u.support,
	})
	if err != nil {
		return fmt.Errorf("SendPasswordChanged: %w", err)
	}

	return u.mailer.Send(&Message{
		To:      []string{toEmail},
		Subject: fmt.Sprintf("[%s] Your Password Was Changed", u.appName),
		Body:    body,
	})
}

// SendStatusChanged notifies user that their account status was updated
func (u *UserMailer) SendStatusChanged(toEmail, name, newStatus string) error {
	body, err := renderTemplate(statusChangedHTML, StatusChangedData{
		Name:         name,
		AppName:      u.appName,
		NewStatus:    newStatus,
		SupportEmail: u.support,
	})
	if err != nil {
		return fmt.Errorf("SendStatusChanged: %w", err)
	}

	return u.mailer.Send(&Message{
		To:      []string{toEmail},
		Subject: fmt.Sprintf("[%s] Your Account Status Has Been Updated", u.appName),
		Body:    body,
	})
}
