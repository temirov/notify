package service

import (
	"context"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/temirov/notify/pkg/model"
	"gorm.io/gorm"
)

// StartRetryWorker periodically attempts to send queued/failed notifications
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

// sendNotification calls "sendWithSendGrid" if it's an email
func sendNotification(notif model.Notification, logger *slog.Logger) (string, string) {
	logger.Info("Sending notification",
		"notification_id", notif.NotificationID,
		"type", notif.NotificationType,
		"retry_count", notif.RetryCount,
	)

	switch notif.NotificationType {
	case model.NotificationEmail:
		return sendWithSendGrid(notif, logger)
	default:
		return "", model.StatusUnknown
	}
}

// For demonstration, we randomly fail 1 in 3 times to test retries.
// Remove or adjust the random failure if you want guaranteed success.
func sendWithSendGrid(notif model.Notification, logger *slog.Logger) (string, string) {
	sender := NewEmailSender(SMTPConfig{
		Host:        "smtp.sendgrid.net",
		Port:        "587",
		Username:    getEnvOrDefault("SENDGRID_USERNAME", "apikey"), // typical for SendGrid
		Password:    getEnvOrDefault("SENDGRID_PASSWORD", ""),       // your API key
		FromAddress: getEnvOrDefault("FROM_EMAIL", "support@rsvp.mprlab.com"),
	})

	// Simulate random failure for demonstration
	if rand.Intn(3) == 0 {
		logger.Error("Simulated send failure for demonstration")
		return "", model.StatusFailed
	}

	err := sender.SendEmail(notif.Recipient, notif.Subject, notif.Message)
	if err != nil {
		logger.Error("SendGrid SMTP send failed", "error", err)
		return "", model.StatusFailed
	}
	return "sendgrid-provider-id", model.StatusSent
}

func getEnvOrDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}
