package service

import (
	"context"
	"fmt"
	"time"

	"github.com/temirov/notify/pkg/config"
	"github.com/temirov/notify/pkg/model"
	"gorm.io/gorm"
	"log/slog"
)

// NotificationService defines the external interface for processing notifications.
type NotificationService interface {
	// SendNotification immediately dispatches the notification and stores it.
	SendNotification(ctx context.Context, request model.NotificationRequest) (model.NotificationResponse, error)
	// GetNotificationStatus retrieves the stored notification status.
	GetNotificationStatus(ctx context.Context, notificationID string) (model.NotificationResponse, error)
	// StartRetryWorker begins a background worker that processes retries with exponential backoff.
	StartRetryWorker(ctx context.Context)
}

type notificationServiceImpl struct {
	database         *gorm.DB
	logger           *slog.Logger
	emailSender      EmailSender
	smsSender        SmsSender
	maxRetries       int
	retryIntervalSec int
}

// NewNotificationService creates a new NotificationService instance using the provided
// database, logger, and service-level configuration. It instantiates its own protocol-specific senders.
func NewNotificationService(db *gorm.DB, logger *slog.Logger, cfg config.Config) NotificationService {
	emailSenderInstance := NewSendGridEmailSender(SMTPConfig{
		Host:        cfg.SendSmtpServer,
		Port:        fmt.Sprintf("%d", cfg.SendSmtpServerPort),
		Username:    cfg.SendGridUsername,
		Password:    cfg.SendGridPassword,
		FromAddress: cfg.FromEmail,
		Timeouts:    cfg, // Pass the full config for timeouts
	}, logger)
	smsSenderInstance := NewTwilioSmsSender(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioFromNumber, logger, cfg)
	return &notificationServiceImpl{
		database:         db,
		logger:           logger,
		emailSender:      emailSenderInstance,
		smsSender:        smsSenderInstance,
		maxRetries:       cfg.MaxRetries,
		retryIntervalSec: cfg.RetryIntervalSec,
	}
}

func (serviceInstance *notificationServiceImpl) SendNotification(ctx context.Context, request model.NotificationRequest) (model.NotificationResponse, error) {
	if request.Recipient == "" || request.Message == "" {
		serviceInstance.logger.Error("Missing required fields", "recipient", request.Recipient, "message", request.Message)
		return model.NotificationResponse{}, fmt.Errorf("missing required fields: recipient or message")
	}
	notificationID := fmt.Sprintf("notif-%d", time.Now().UnixNano())
	newNotification := model.NewNotification(notificationID, request)

	currentTime := time.Now().UTC()
	shouldAttemptImmediateSend := true
	if request.ScheduledFor != nil && request.ScheduledFor.After(currentTime) {
		shouldAttemptImmediateSend = false
	}

	var dispatchError error
	if shouldAttemptImmediateSend {
		switch newNotification.NotificationType {
		case model.NotificationEmail:
			dispatchError = serviceInstance.emailSender.SendEmail(ctx, newNotification.Recipient, newNotification.Subject, newNotification.Message)
			if dispatchError == nil {
				newNotification.Status = model.StatusSent
				newNotification.LastAttemptedAt = currentTime
				// When using SMTP no provider message ID is returned.
			}
		case model.NotificationSMS:
			var providerMessageID string
			providerMessageID, dispatchError = serviceInstance.smsSender.SendSms(ctx, newNotification.Recipient, newNotification.Message)
			if dispatchError == nil {
				newNotification.Status = model.StatusSent
				newNotification.ProviderMessageID = providerMessageID
				newNotification.LastAttemptedAt = currentTime
			}
		default:
			serviceInstance.logger.Error("Unsupported notification type", "type", newNotification.NotificationType)
			return model.NotificationResponse{}, fmt.Errorf("unsupported notification type: %s", newNotification.NotificationType)
		}
		if dispatchError != nil {
			serviceInstance.logger.Error("Immediate dispatch failed", "error", dispatchError)
			newNotification.Status = model.StatusFailed
			newNotification.LastAttemptedAt = currentTime
		}
	}

	if err := model.CreateNotification(ctx, serviceInstance.database, &newNotification); err != nil {
		serviceInstance.logger.Error("Failed to store notification", "error", err)
		return model.NotificationResponse{}, err
	}
	return model.NewNotificationResponse(newNotification), nil
}

func (serviceInstance *notificationServiceImpl) GetNotificationStatus(ctx context.Context, notificationID string) (model.NotificationResponse, error) {
	if notificationID == "" {
		serviceInstance.logger.Error("Missing notification_id")
		return model.NotificationResponse{}, fmt.Errorf("missing notification_id")
	}
	notificationRecord, retrievalError := model.MustGetNotificationByID(ctx, serviceInstance.database, notificationID)
	if retrievalError != nil {
		serviceInstance.logger.Error("Failed to retrieve notification", "error", retrievalError)
		return model.NotificationResponse{}, retrievalError
	}
	return model.NewNotificationResponse(*notificationRecord), nil
}

func (serviceInstance *notificationServiceImpl) StartRetryWorker(ctx context.Context) {
	retryTicker := time.NewTicker(time.Duration(serviceInstance.retryIntervalSec) * time.Second)
	defer retryTicker.Stop()

	serviceInstance.logger.Info("Starting retry worker", "base_interval_sec", serviceInstance.retryIntervalSec, "max_retries", serviceInstance.maxRetries)
	for {
		select {
		case <-ctx.Done():
			serviceInstance.logger.Info("Retry worker shutting down")
			return
		case <-retryTicker.C:
			serviceInstance.processRetries(ctx)
		}
	}
}

func (serviceInstance *notificationServiceImpl) processRetries(ctx context.Context) {
	currentTime := time.Now().UTC()
	notificationRecords, fetchError := model.GetQueuedOrFailedNotifications(ctx, serviceInstance.database, serviceInstance.maxRetries, currentTime)
	if fetchError != nil {
		serviceInstance.logger.Error("Failed to fetch notifications for retry", "error", fetchError)
		return
	}
	for _, notificationRecord := range notificationRecords {
		if notificationRecord.ScheduledFor != nil && currentTime.Before(notificationRecord.ScheduledFor.UTC()) {
			continue
		}
		// Apply exponential backoff: baseInterval * 2^(retry_count)
		if notificationRecord.RetryCount > 0 {
			backoffDuration := time.Duration(serviceInstance.retryIntervalSec) * time.Second * (1 << uint(notificationRecord.RetryCount))
			nextAttemptTime := notificationRecord.LastAttemptedAt.Add(backoffDuration)
			if currentTime.Before(nextAttemptTime) {
				continue
			}
		}
		var dispatchError error
		switch notificationRecord.NotificationType {
		case model.NotificationEmail:
			dispatchError = serviceInstance.emailSender.SendEmail(ctx, notificationRecord.Recipient, notificationRecord.Subject, notificationRecord.Message)
			if dispatchError == nil {
				notificationRecord.Status = model.StatusSent
				notificationRecord.ProviderMessageID = ""
			}
		case model.NotificationSMS:
			var providerMessageID string
			providerMessageID, dispatchError = serviceInstance.smsSender.SendSms(ctx, notificationRecord.Recipient, notificationRecord.Message)
			if dispatchError == nil {
				notificationRecord.Status = model.StatusSent
				notificationRecord.ProviderMessageID = providerMessageID
			}
		default:
			serviceInstance.logger.Error("Unsupported notification type during retry", "notification_id", notificationRecord.NotificationID)
			continue
		}
		if dispatchError != nil {
			serviceInstance.logger.Error("Dispatch failed during retry", "notification_id", notificationRecord.NotificationID, "error", dispatchError)
			notificationRecord.Status = model.StatusFailed
		}
		notificationRecord.RetryCount++
		notificationRecord.LastAttemptedAt = currentTime
		if saveError := model.SaveNotification(ctx, serviceInstance.database, &notificationRecord); saveError != nil {
			serviceInstance.logger.Error("Failed to update notification after retry", "notification_id", notificationRecord.NotificationID, "error", saveError)
		}
	}
}
