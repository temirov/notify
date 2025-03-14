package service

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// SMTPConfig for sending via SendGrid or other providers
type SMTPConfig struct {
	Host        string
	Port        string
	Username    string
	Password    string
	FromAddress string // e.g. "support@rsvp.mprlab.com"
}

// EmailSender uses standard net/smtp to send emails
type EmailSender struct {
	cfg SMTPConfig
}

func NewEmailSender(cfg SMTPConfig) *EmailSender {
	return &EmailSender{cfg: cfg}
}

// SendEmail sends a plain-text email
func (es *EmailSender) SendEmail(to, subject, body string) error {
	smtpAddr := net.JoinHostPort(es.cfg.Host, es.cfg.Port)
	msg := buildMessage(es.cfg.FromAddress, to, subject, body)

	auth := smtp.PlainAuth("", es.cfg.Username, es.cfg.Password, es.cfg.Host)
	err := smtp.SendMail(smtpAddr, auth, es.cfg.FromAddress, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("smtp send failed: %w", err)
	}
	return nil
}

func buildMessage(from, to, subject, body string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From: %s\r\n", from))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", to))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)
	return sb.String()
}
