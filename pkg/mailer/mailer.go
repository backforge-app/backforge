// Package mailer provides a generic, reusable, and context-aware SMTP email sender.
//
// It handles the construction of MIME headers, HTML bodies, and secure SMTP
// authentication. It wraps standard net/smtp and adds context cancellation
// support to prevent hanging requests during network timeouts.
package mailer

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// Config holds the SMTP configuration required to send emails.
type Config struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
}

// Mailer is a generic SMTP email sender.
type Mailer struct {
	cfg Config
}

// New creates a new Mailer instance with the provided configuration.
func New(cfg Config) *Mailer {
	return &Mailer{
		cfg: cfg,
	}
}

// SendHTML sends an HTML-formatted email to the specified recipient.
//
// It respects the provided context, allowing for timeout or cancellation
// during the SMTP transmission.
func (m *Mailer) SendHTML(ctx context.Context, to, subject, htmlBody string) error {
	// 1. Construct standard MIME headers for HTML email.
	headers := make(map[string]string)
	headers["From"] = m.cfg.FromAddress
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	var msgBuilder strings.Builder
	for k, v := range headers {
		msgBuilder.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	// A blank line separates headers from the body in standard SMTP.
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(htmlBody)

	msg := []byte(msgBuilder.String())
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	// 2. Execute sending in a goroutine to support context cancellation.
	// net/smtp.SendMail does not natively support context, so we wrap it.
	ch := make(chan error, 1)
	go func() {
		ch <- smtp.SendMail(addr, auth, m.cfg.FromAddress, []string{to}, msg)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("email sending canceled or timed out: %w", ctx.Err())
	case err := <-ch:
		if err != nil {
			return fmt.Errorf("smtp send failed: %w", err)
		}
		return nil
	}
}
