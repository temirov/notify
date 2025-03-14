package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/temirov/notify/pkg/model"
	"gorm.io/gorm"
)

func StartRetryWorker(
	ctx context.Context,
	database *gorm.DB,
	logger *slog.Logger,
	retryIntervalSec int,
	maxRetries int,
) {
	ticker := time.NewTicker(time.Duration(retryIntervalSec) * time.Second)
	defer ticker.Stop()

	logger.Info("Starting retry worker",
		"interval_sec", retryIntervalSec,
		"max_retries", maxRetries,
	)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Retry worker shutting down")
			return
		case <-ticker.C:
			processRetries(database, logger, maxRetries)
		}
	}
}

func processRetries(database *gorm.DB, logger *slog.Logger, maxRetries int) {
	var notifications []model.Notification
	err := database.Where(
		"(status = ? OR status = ?) AND retry_count < ?",
		model.StatusQueued, model.StatusFailed, maxRetries,
	).Find(&notifications).Error
	if err != nil {
		logger.Error("Fetching notifications for retry failed", "error", err)
		return
	}
	if len(notifications) == 0 {
		return // nothing to retry
	}

	for _, notif := range notifications {
		providerID, finalStatus := sendNotification(notif, logger)
		notif.ProviderMessageID = providerID
		notif.Status = finalStatus
		notif.RetryCount++
		notif.LastAttemptedAt = time.Now().UTC()

		saveErr := database.Save(&notif).Error
		if saveErr != nil {
			logger.Error("Failed to update notification after send attempt", "error", saveErr)
		}
	}
}

func sendNotification(notif model.Notification, logger *slog.Logger) (string, string) {
	logger.Info("Sending notification",
		"notification_id", notif.NotificationID,
		"type", notif.NotificationType,
		"retry_count", notif.RetryCount,
	)

	switch notif.NotificationType {
	case model.NotificationEmail:
		return sendWithSendGrid(notif, logger)
	case model.NotificationSMS:
		return sendWithTwilio(notif, logger)
	default:
		return "", model.StatusUnknown
	}
}

// ========== Email via SendGrid SMTP ========== //

func sendWithSendGrid(notif model.Notification, logger *slog.Logger) (string, string) {
	sender := NewEmailSender(SMTPConfig{
		Host:        "smtp.sendgrid.net",
		Port:        "587",
		Username:    getEnvOrDefault("SENDGRID_USERNAME", "apikey"),
		Password:    getEnvOrDefault("SENDGRID_PASSWORD", ""),
		FromAddress: getEnvOrDefault("FROM_EMAIL", "support@rsvp.mprlab.com"),
	})

	// Optional: random failure for demonstration
	if rand.Intn(3) == 0 {
		logger.Error("Simulated email send failure for demonstration")
		return "", model.StatusFailed
	}

	err := sender.SendEmail(notif.Recipient, notif.Subject, notif.Message)
	if err != nil {
		logger.Error("SendGrid SMTP send failed", "error", err)
		return "", model.StatusFailed
	}
	return "sendgrid-provider-id", model.StatusSent
}

// ========== SMS via Twilio REST API ========== //

func sendWithTwilio(notif model.Notification, logger *slog.Logger) (string, string) {
	accountSID := getEnvOrDefault("TWILIO_ACCOUNT_SID", "")
	authToken := getEnvOrDefault("TWILIO_AUTH_TOKEN", "")
	fromNumber := getEnvOrDefault("TWILIO_FROM_NUMBER", "")

	if accountSID == "" || authToken == "" || fromNumber == "" {
		logger.Error("Twilio env variables missing")
		return "", model.StatusFailed
	}

	// Optional: random failure for demonstration
	if rand.Intn(3) == 0 {
		logger.Error("Simulated SMS send failure for demonstration")
		return "", model.StatusFailed
	}

	// Twilio API endpoint
	twilioURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID)

	// Encode form data
	data := url.Values{}
	data.Set("To", notif.Recipient)
	data.Set("From", fromNumber)
	data.Set("Body", notif.Message)

	req, err := http.NewRequest(http.MethodPost, twilioURL, strings.NewReader(data.Encode()))
	if err != nil {
		logger.Error("Failed to create Twilio request", "error", err)
		return "", model.StatusFailed
	}

	req.SetBasicAuth(accountSID, authToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Twilio request failed", "error", err)
		return "", model.StatusFailed
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		logger.Error("Twilio API error",
			"status", resp.StatusCode,
			"body", string(bodyBytes),
		)
		return "", model.StatusFailed
	}

	// Typically Twilio responds with JSON containing "sid"
	// For demonstration, let's parse out the "sid" or just store entire body as "providerMessageID"
	return string(bodyBytes), model.StatusSent
}

// ========== Helpers ========== //

func getEnvOrDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}
