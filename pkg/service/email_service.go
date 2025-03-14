package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/temirov/notify/pkg/config"
	"log/slog"
)

type SMTPConfig struct {
	Host        string
	Port        string
	Username    string
	Password    string
	FromAddress string
	Timeouts    config.Config
}

type EmailSender interface {
	SendEmail(ctx context.Context, recipient string, subject string, message string) error
}

type SendGridEmailSender struct {
	Config SMTPConfig
	Logger *slog.Logger
}

func NewSendGridEmailSender(configuration SMTPConfig, logger *slog.Logger) *SendGridEmailSender {
	return &SendGridEmailSender{
		Config: configuration,
		Logger: logger,
	}
}

func (senderInstance *SendGridEmailSender) SendEmail(ctx context.Context, recipient string, subject string, message string) error {
	emailMessage := buildEmailMessage(senderInstance.Config.FromAddress, recipient, subject, message)

	if senderInstance.Config.Port == "465" {
		serverAddr := net.JoinHostPort(senderInstance.Config.Host, senderInstance.Config.Port)
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // In production, perform proper certificate validation.
			ServerName:         senderInstance.Config.Host,
		}

		dialer := &net.Dialer{
			Timeout: time.Duration(senderInstance.Config.Timeouts.ConnectionTimeoutSec) * time.Second,
		}

		tlsConnection, dialError := tls.DialWithDialer(dialer, "tcp", serverAddr, tlsConfig)
		if dialError != nil {
			return fmt.Errorf("failed to dial TLS: %w", dialError)
		}
		defer tlsConnection.Close()

		if ctx.Err() != nil {
			return ctx.Err()
		}

		smtpClient, clientError := smtp.NewClient(tlsConnection, senderInstance.Config.Host)
		if clientError != nil {
			return fmt.Errorf("failed to create SMTP client: %w", clientError)
		}
		defer smtpClient.Quit()

		smtpAuth := smtp.PlainAuth("", senderInstance.Config.Username, senderInstance.Config.Password, senderInstance.Config.Host)
		if authError := smtpClient.Auth(smtpAuth); authError != nil {
			return fmt.Errorf("failed to authenticate: %w", authError)
		}

		if mailError := smtpClient.Mail(senderInstance.Config.FromAddress); mailError != nil {
			return fmt.Errorf("failed to set sender: %w", mailError)
		}
		if rcptError := smtpClient.Rcpt(recipient); rcptError != nil {
			return fmt.Errorf("failed to set recipient: %w", rcptError)
		}

		dataWriter, dataError := smtpClient.Data()
		if dataError != nil {
			return fmt.Errorf("failed to get data writer: %w", dataError)
		}
		_, writeError := dataWriter.Write([]byte(emailMessage))
		if writeError != nil {
			dataWriter.Close()
			return fmt.Errorf("failed to write email message: %w", writeError)
		}
		if closeDataError := dataWriter.Close(); closeDataError != nil {
			return fmt.Errorf("failed to close data writer: %w", closeDataError)
		}

		return nil
	}

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
