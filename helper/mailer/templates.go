package mailer

import (
	"bytes"
	"fmt"
	"html/template"
)

// ─── Template data structs ────────────────────────────────────────────────────

// WelcomeData is the data injected into the welcome email template
type WelcomeData struct {
	Name     string
	AppName  string
	LoginURL string
}

// PasswordResetData is the data injected into the password reset template
type PasswordResetData struct {
	Name      string
	AppName   string
	ResetURL  string
	ExpiryMin int
}

// PasswordChangedData is the data for password-changed confirmation
type PasswordChangedData struct {
	Name         string
	AppName      string
	SupportEmail string
}

// StatusChangedData is the data for account status change notification
type StatusChangedData struct {
	Name         string
	AppName      string
	NewStatus    string
	SupportEmail string
}

// ─── Template strings ─────────────────────────────────────────────────────────

const welcomeHTML = `
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>Welcome</title></head>
<body style="font-family:Arial,sans-serif;background:#f4f4f4;margin:0;padding:0;">
  <table width="100%" cellpadding="0" cellspacing="0">
    <tr><td align="center" style="padding:40px 0;">
      <table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.1);">
        <tr><td style="background:#4F46E5;padding:32px;text-align:center;">
          <h1 style="color:#ffffff;margin:0;font-size:28px;">{{.AppName}}</h1>
        </td></tr>
        <tr><td style="padding:40px 32px;">
          <h2 style="color:#1a1a1a;margin-top:0;">Welcome, {{.Name}}! 👋</h2>
          <p style="color:#555;line-height:1.6;">Your account has been created successfully. We're thrilled to have you on board.</p>
          <div style="text-align:center;margin:32px 0;">
            <a href="{{.LoginURL}}" style="background:#4F46E5;color:#ffffff;padding:14px 32px;border-radius:6px;text-decoration:none;font-weight:bold;font-size:16px;">Get Started</a>
          </div>
          <p style="color:#999;font-size:13px;">If you did not create this account, please ignore this email.</p>
        </td></tr>
        <tr><td style="background:#f9f9f9;padding:20px 32px;text-align:center;">
          <p style="color:#aaa;font-size:12px;margin:0;">© {{.AppName}}. All rights reserved.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`

const passwordResetHTML = `
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>Password Reset</title></head>
<body style="font-family:Arial,sans-serif;background:#f4f4f4;margin:0;padding:0;">
  <table width="100%" cellpadding="0" cellspacing="0">
    <tr><td align="center" style="padding:40px 0;">
      <table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.1);">
        <tr><td style="background:#DC2626;padding:32px;text-align:center;">
          <h1 style="color:#ffffff;margin:0;font-size:28px;">{{.AppName}}</h1>
        </td></tr>
        <tr><td style="padding:40px 32px;">
          <h2 style="color:#1a1a1a;margin-top:0;">Password Reset Request</h2>
          <p style="color:#555;line-height:1.6;">Hi <strong>{{.Name}}</strong>, we received a request to reset your password.</p>
          <div style="text-align:center;margin:32px 0;">
            <a href="{{.ResetURL}}" style="background:#DC2626;color:#ffffff;padding:14px 32px;border-radius:6px;text-decoration:none;font-weight:bold;font-size:16px;">Reset Password</a>
          </div>
          <p style="color:#555;line-height:1.6;">This link expires in <strong>{{.ExpiryMin}} minutes</strong>.</p>
          <p style="color:#999;font-size:13px;">If you did not request a password reset, please ignore this email. Your password will remain unchanged.</p>
        </td></tr>
        <tr><td style="background:#f9f9f9;padding:20px 32px;text-align:center;">
          <p style="color:#aaa;font-size:12px;margin:0;">© {{.AppName}}. All rights reserved.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`

const passwordChangedHTML = `
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>Password Changed</title></head>
<body style="font-family:Arial,sans-serif;background:#f4f4f4;margin:0;padding:0;">
  <table width="100%" cellpadding="0" cellspacing="0">
    <tr><td align="center" style="padding:40px 0;">
      <table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.1);">
        <tr><td style="background:#059669;padding:32px;text-align:center;">
          <h1 style="color:#ffffff;margin:0;font-size:28px;">{{.AppName}}</h1>
        </td></tr>
        <tr><td style="padding:40px 32px;">
          <h2 style="color:#1a1a1a;margin-top:0;">Password Changed Successfully ✅</h2>
          <p style="color:#555;line-height:1.6;">Hi <strong>{{.Name}}</strong>, your password has been updated successfully.</p>
          <p style="color:#555;line-height:1.6;">If you did not make this change, please contact us immediately at <a href="mailto:{{.SupportEmail}}" style="color:#059669;">{{.SupportEmail}}</a>.</p>
        </td></tr>
        <tr><td style="background:#f9f9f9;padding:20px 32px;text-align:center;">
          <p style="color:#aaa;font-size:12px;margin:0;">© {{.AppName}}. All rights reserved.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`

const statusChangedHTML = `
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>Account Status Updated</title></head>
<body style="font-family:Arial,sans-serif;background:#f4f4f4;margin:0;padding:0;">
  <table width="100%" cellpadding="0" cellspacing="0">
    <tr><td align="center" style="padding:40px 0;">
      <table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.1);">
        <tr><td style="background:#D97706;padding:32px;text-align:center;">
          <h1 style="color:#ffffff;margin:0;font-size:28px;">{{.AppName}}</h1>
        </td></tr>
        <tr><td style="padding:40px 32px;">
          <h2 style="color:#1a1a1a;margin-top:0;">Account Status Updated</h2>
          <p style="color:#555;line-height:1.6;">Hi <strong>{{.Name}}</strong>, your account status has been changed to:</p>
          <div style="text-align:center;margin:24px 0;">
            <span style="background:#FEF3C7;color:#92400E;padding:10px 24px;border-radius:20px;font-weight:bold;font-size:18px;text-transform:uppercase;">{{.NewStatus}}</span>
          </div>
          <p style="color:#555;line-height:1.6;">If you have questions about this change, contact us at <a href="mailto:{{.SupportEmail}}" style="color:#D97706;">{{.SupportEmail}}</a>.</p>
        </td></tr>
        <tr><td style="background:#f9f9f9;padding:20px 32px;text-align:center;">
          <p style="color:#aaa;font-size:12px;margin:0;">© {{.AppName}}. All rights reserved.</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`

// ─── Template renderer ────────────────────────────────────────────────────────

func renderTemplate(tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("mailer: template parse error: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("mailer: template execute error: %w", err)
	}
	return buf.String(), nil
}
