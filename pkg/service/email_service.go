package service

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"log/slog"
)

// SMTPConfig holds the SMTP settings.
type SMTPConfig struct {
	Host        string
	Port        string
	Username    string
	Password    string
	FromAddress string
}

// EmailSender defines the behavior of an email sending service.
type EmailSender interface {
	SendEmail(recipient string, subject string, message string) error
}

// SendGridEmailSender implements EmailSender using net/smtp.
type SendGridEmailSender struct {
	Config SMTPConfig
	Logger *slog.Logger
}

// NewSendGridEmailSender creates a new SendGridEmailSender with the provided configuration and logger.
func NewSendGridEmailSender(configuration SMTPConfig, logger *slog.Logger) *SendGridEmailSender {
	return &SendGridEmailSender{
		Config: configuration,
		Logger: logger,
	}
}

// SendEmail sends an email using the SMTP settings.
// If the port is "465", it uses an implicit TLS connection.
// All errors, including those from Quit, are returned to the caller.
func (senderInstance *SendGridEmailSender) SendEmail(recipient string, subject string, message string) error {
	emailMessage := buildEmailMessage(senderInstance.Config.FromAddress, recipient, subject, message)

	if senderInstance.Config.Port == "465" {
		serverAddr := net.JoinHostPort(senderInstance.Config.Host, senderInstance.Config.Port)
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // In production, perform proper certificate validation.
			ServerName:         senderInstance.Config.Host,
		}

		tlsConnection, dialError := tls.Dial("tcp", serverAddr, tlsConfig)
		if dialError != nil {
			return fmt.Errorf("failed to dial TLS: %w", dialError)
		}

		// Create SMTP client using the established TLS connection.
		smtpClient, clientError := smtp.NewClient(tlsConnection, senderInstance.Config.Host)
		if clientError != nil {
			_ = tlsConnection.Close() // Close if SMTP client creation fails.
			return fmt.Errorf("failed to create SMTP client: %w", clientError)
		}

		// Authenticate.
		smtpAuth := smtp.PlainAuth("", senderInstance.Config.Username, senderInstance.Config.Password, senderInstance.Config.Host)
		if authError := smtpClient.Auth(smtpAuth); authError != nil {
			_ = smtpClient.Quit()
			return fmt.Errorf("failed to authenticate: %w", authError)
		}

		// Set the sender and recipient.
		if mailError := smtpClient.Mail(senderInstance.Config.FromAddress); mailError != nil {
			_ = smtpClient.Quit()
			return fmt.Errorf("failed to set sender: %w", mailError)
		}
		if rcptError := smtpClient.Rcpt(recipient); rcptError != nil {
			_ = smtpClient.Quit()
			return fmt.Errorf("failed to set recipient: %w", rcptError)
		}

		// Write the email data.
		dataWriter, dataError := smtpClient.Data()
		if dataError != nil {
			_ = smtpClient.Quit()
			return fmt.Errorf("failed to get data writer: %w", dataError)
		}
		_, writeError := dataWriter.Write([]byte(emailMessage))
		if writeError != nil {
			_ = dataWriter.Close()
			_ = smtpClient.Quit()
			return fmt.Errorf("failed to write email message: %w", writeError)
		}
		if closeDataError := dataWriter.Close(); closeDataError != nil {
			_ = smtpClient.Quit()
			return fmt.Errorf("failed to close data writer: %w", closeDataError)
		}

		// Call Quit explicitly and return any error.
		if quitError := smtpClient.Quit(); quitError != nil {
			return fmt.Errorf("failed to quit SMTP client: %w", quitError)
		}
		return nil
	}

	// For other ports (e.g. 587), use smtp.SendMail (which handles STARTTLS).
	smtpAddress := net.JoinHostPort(senderInstance.Config.Host, senderInstance.Config.Port)
	smtpAuth := smtp.PlainAuth("", senderInstance.Config.Username, senderInstance.Config.Password, senderInstance.Config.Host)
	sendError := smtp.SendMail(smtpAddress, smtpAuth, senderInstance.Config.FromAddress, []string{recipient}, []byte(emailMessage))
	if sendError != nil {
		return fmt.Errorf("smtp send failed: %w", sendError)
	}
	return nil
}

func buildEmailMessage(fromAddress string, toAddress string, subject string, body string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("From: %s\r\n", fromAddress))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", toAddress))
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	builder.WriteString("MIME-Version: 1.0\r\n")
	builder.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	builder.WriteString("\r\n")
	builder.WriteString(body)
	return builder.String()
}
