package mailer

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"go.uber.org/zap"
)

// Config holds SMTP connection settings
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string // display sender e.g. "User Service <no-reply@example.com>"
	UseTLS   bool
}

// Mailer is the core SMTP client
type Mailer struct {
	cfg    *Config
	logger *zap.Logger
}

// Message represents a single outbound email
type Message struct {
	To      []string
	Subject string
	Body    string // HTML body
	CC      []string
	BCC     []string
}

// New constructs a Mailer
func New(cfg *Config, logger *zap.Logger) *Mailer {
	return &Mailer{cfg: cfg, logger: logger}
}

// Send delivers a Message via SMTP
func (m *Mailer) Send(msg *Message) error {
	if len(msg.To) == 0 {
		return fmt.Errorf("mailer: no recipients specified")
	}

	raw := m.buildRaw(msg)

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	var err error
	if m.cfg.UseTLS {
		err = m.sendTLS(addr, auth, msg.To, raw)
	} else {
		err = smtp.SendMail(addr, auth, m.cfg.Username, msg.To, []byte(raw))
	}

	if err != nil {
		m.logger.Error("mailer: failed to send email",
			zap.Strings("to", msg.To),
			zap.String("subject", msg.Subject),
			zap.Error(err),
		)
		return fmt.Errorf("mailer: send failed: %w", err)
	}

	m.logger.Info("mailer: email sent",
		zap.Strings("to", msg.To),
		zap.String("subject", msg.Subject),
	)
	return nil
}

// sendTLS dials with TLS (port 465) instead of STARTTLS (port 587)
func (m *Mailer) sendTLS(addr string, auth smtp.Auth, to []string, raw string) error {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         m.cfg.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("mailer: TLS dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		return fmt.Errorf("mailer: SMTP client error: %w", err)
	}
	defer client.Quit()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("mailer: auth failed: %w", err)
	}
	if err = client.Mail(m.cfg.Username); err != nil {
		return err
	}
	for _, r := range to {
		if err = client.Rcpt(r); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.Write([]byte(raw))
	return err
}

// buildRaw constructs the raw RFC 2822 email string
func (m *Mailer) buildRaw(msg *Message) string {
	var sb strings.Builder

	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	sb.WriteString(fmt.Sprintf("From: %s\r\n", m.cfg.From))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))

	if len(msg.CC) > 0 {
		sb.WriteString(fmt.Sprintf("CC: %s\r\n", strings.Join(msg.CC, ", ")))
	}

	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	sb.WriteString("\r\n")
	sb.WriteString(msg.Body)

	return sb.String()
}
