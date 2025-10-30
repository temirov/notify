package service

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/temirov/pinguin/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log/slog"
)

func TestSendNotificationRespectsSchedule(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name                    string
		scheduledOffset         *time.Duration
		expectedStatus          string
		expectImmediateDispatch bool
	}{
		{
			name:                    "ImmediateSendWithoutSchedule",
			scheduledOffset:         nil,
			expectedStatus:          model.StatusSent,
			expectImmediateDispatch: true,
		},
		{
			name:                    "ImmediateSendForPastSchedule",
			scheduledOffset:         durationPointer(-1 * time.Minute),
			expectedStatus:          model.StatusSent,
			expectImmediateDispatch: true,
		},
		{
			name:                    "QueuedForFutureSchedule",
			scheduledOffset:         durationPointer(2 * time.Minute),
			expectedStatus:          model.StatusQueued,
			expectImmediateDispatch: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Helper()

			database := openIsolatedDatabase(t)
			emailSender := &stubEmailSender{}
			smsSender := &stubSmsSender{}

			serviceInstance := &notificationServiceImpl{
				database:         database,
				logger:           slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
				emailSender:      emailSender,
				smsSender:        smsSender,
				maxRetries:       5,
				retryIntervalSec: 1,
				smsEnabled:       true,
			}

			var scheduledFor *time.Time
			if testCase.scheduledOffset != nil {
				scheduledTime := time.Now().UTC().Add(*testCase.scheduledOffset)
				scheduledFor = &scheduledTime
			}

			request := model.NotificationRequest{
				NotificationType: model.NotificationEmail,
				Recipient:        "user@example.com",
				Subject:          "Subject",
				Message:          "Body",
				ScheduledFor:     scheduledFor,
			}

			response, responseError := serviceInstance.SendNotification(context.Background(), request)
			if responseError != nil {
				t.Fatalf("SendNotification error: %v", responseError)
			}

			if response.Status != testCase.expectedStatus {
				t.Fatalf("unexpected status %s", response.Status)
			}

			if testCase.expectImmediateDispatch && emailSender.callCount != 1 {
				t.Fatalf("expected immediate dispatch")
			}
			if !testCase.expectImmediateDispatch && emailSender.callCount != 0 {
				t.Fatalf("unexpected immediate dispatch")
			}

			if testCase.scheduledOffset == nil && response.ScheduledFor != nil {
				t.Fatalf("expected nil scheduledFor in response")
			}
			if testCase.scheduledOffset != nil {
				if response.ScheduledFor == nil {
					t.Fatalf("expected scheduledFor value in response")
				}
			}
		})
	}
}

func TestSendNotificationRejectsUnsupportedTypes(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name            string
		scheduledOffset *time.Duration
	}{
		{
			name:            "ImmediateUnsupportedType",
			scheduledOffset: nil,
		},
		{
			name:            "ScheduledUnsupportedType",
			scheduledOffset: durationPointer(2 * time.Minute),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Helper()

			database := openIsolatedDatabase(t)
			emailSender := &stubEmailSender{}
			smsSender := &stubSmsSender{}

			serviceInstance := &notificationServiceImpl{
				database:         database,
				logger:           slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
				emailSender:      emailSender,
				smsSender:        smsSender,
				maxRetries:       5,
				retryIntervalSec: 1,
				smsEnabled:       true,
			}

			var scheduledFor *time.Time
			if testCase.scheduledOffset != nil {
				scheduledTime := time.Now().UTC().Add(*testCase.scheduledOffset)
				scheduledFor = &scheduledTime
			}

			_, responseError := serviceInstance.SendNotification(context.Background(), model.NotificationRequest{
				NotificationType: model.NotificationType("push"),
				Recipient:        "user@example.com",
				Subject:          "Subject",
				Message:          "Body",
				ScheduledFor:     scheduledFor,
			})
			if responseError == nil {
				t.Fatalf("expected unsupported type error")
			}

			if emailSender.callCount != 0 {
				t.Fatalf("unexpected email dispatch attempts")
			}
			if smsSender.callCount != 0 {
				t.Fatalf("unexpected sms dispatch attempts")
			}

			pendingNotifications, pendingError := model.GetQueuedOrFailedNotifications(context.Background(), database, 5, time.Now().UTC())
			if pendingError != nil {
				t.Fatalf("pending notifications error: %v", pendingError)
			}
			if len(pendingNotifications) != 0 {
				t.Fatalf("unexpected stored notifications")
			}
		})
	}
}

func TestProcessRetriesRespectsSchedule(t *testing.T) {
	t.Helper()

	database := openIsolatedDatabase(t)
	emailSender := &stubEmailSender{}
	smsSender := &stubSmsSender{}

	serviceInstance := &notificationServiceImpl{
		database:         database,
		logger:           slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
		emailSender:      emailSender,
		smsSender:        smsSender,
		maxRetries:       5,
		retryIntervalSec: 1,
		smsEnabled:       true,
	}

	notificationIdentifier := "notif-scheduled"
	now := time.Now().UTC()
	futureScheduled := now.Add(5 * time.Minute)

	scheduledNotification := model.Notification{
		NotificationID:   notificationIdentifier,
		NotificationType: model.NotificationEmail,
		Recipient:        "user@example.com",
		Message:          "Body",
		Status:           model.StatusQueued,
		ScheduledFor:     &futureScheduled,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if createError := model.CreateNotification(context.Background(), database, &scheduledNotification); createError != nil {
		t.Fatalf("create notification error: %v", createError)
	}

	serviceInstance.processRetries(context.Background())
	if emailSender.callCount != 0 {
		t.Fatalf("expected zero dispatches before schedule")
	}

	pastScheduled := now.Add(-1 * time.Minute)
	scheduledNotification.ScheduledFor = &pastScheduled
	scheduledNotification.Status = model.StatusQueued
	if saveError := model.SaveNotification(context.Background(), database, &scheduledNotification); saveError != nil {
		t.Fatalf("save notification error: %v", saveError)
	}

	serviceInstance.processRetries(context.Background())
	if emailSender.callCount != 1 {
		t.Fatalf("expected one dispatch after schedule")
	}

	fetchedNotification, fetchError := model.GetNotificationByID(context.Background(), database, notificationIdentifier)
	if fetchError != nil {
		t.Fatalf("fetch notification error: %v", fetchError)
	}

	if fetchedNotification.Status != model.StatusSent {
		t.Fatalf("unexpected status %s", fetchedNotification.Status)
	}
	if fetchedNotification.RetryCount != 1 {
		t.Fatalf("unexpected retry count %d", fetchedNotification.RetryCount)
	}
	if fetchedNotification.LastAttemptedAt.IsZero() {
		t.Fatalf("expected last attempted timestamp")
	}
}

func TestSendNotificationValidatesRequiredFields(t *testing.T) {
	t.Helper()

	database := openIsolatedDatabase(t)
	emailSender := &stubEmailSender{}
	smsSender := &stubSmsSender{}

	serviceInstance := &notificationServiceImpl{
		database:         database,
		logger:           slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
		emailSender:      emailSender,
		smsSender:        smsSender,
		maxRetries:       3,
		retryIntervalSec: 1,
		smsEnabled:       true,
	}

	_, sendError := serviceInstance.SendNotification(context.Background(), model.NotificationRequest{
		NotificationType: model.NotificationSMS,
		Recipient:        "",
		Message:          "Body",
	})
	if sendError == nil {
		t.Fatalf("expected validation error")
	}

	_, fetchError := model.GetQueuedOrFailedNotifications(context.Background(), database, 3, time.Now().UTC())
	if fetchError != nil {
		t.Fatalf("fetch pending error: %v", fetchError)
	}

	if emailSender.callCount != 0 || smsSender.callCount != 0 {
		t.Fatalf("unexpected dispatch attempts")
	}
}

func TestSendNotificationRejectsUnsupportedTypeForScheduledRequests(t *testing.T) {
	t.Helper()

	database := openIsolatedDatabase(t)
	emailSender := &stubEmailSender{}
	smsSender := &stubSmsSender{}

	serviceInstance := &notificationServiceImpl{
		database:         database,
		logger:           slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
		emailSender:      emailSender,
		smsSender:        smsSender,
		maxRetries:       3,
		retryIntervalSec: 1,
		smsEnabled:       true,
	}

	scheduledFor := time.Now().UTC().Add(5 * time.Minute)
	_, sendError := serviceInstance.SendNotification(context.Background(), model.NotificationRequest{
		NotificationType: model.NotificationType("push"),
		Recipient:        "user@example.com",
		Message:          "Body",
		ScheduledFor:     &scheduledFor,
	})

	if sendError == nil {
		t.Fatalf("expected unsupported type error")
	}

	if emailSender.callCount != 0 || smsSender.callCount != 0 {
		t.Fatalf("unexpected dispatch attempts")
	}

	var notificationCount int64
	if countError := database.WithContext(context.Background()).Model(&model.Notification{}).Count(&notificationCount).Error; countError != nil {
		t.Fatalf("count notifications error: %v", countError)
	}
	if notificationCount != 0 {
		t.Fatalf("expected zero stored notifications, got %d", notificationCount)
	}
}

func TestSendNotificationRejectsSmsWhenSenderDisabled(t *testing.T) {
	t.Helper()

	database := openIsolatedDatabase(t)
	serviceInstance := &notificationServiceImpl{
		database:         database,
		logger:           slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
		emailSender:      &stubEmailSender{},
		smsSender:        nil,
		maxRetries:       3,
		retryIntervalSec: 1,
		smsEnabled:       false,
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("unexpected panic when sms sender disabled: %v", recovered)
		}
	}()

	_, sendError := serviceInstance.SendNotification(context.Background(), model.NotificationRequest{
		NotificationType: model.NotificationSMS,
		Recipient:        "+15555555555",
		Message:          "Body",
	})
	if sendError == nil {
		t.Fatalf("expected error when sms sender is disabled")
	}

	var notificationCount int64
	if countError := database.WithContext(context.Background()).Model(&model.Notification{}).Count(&notificationCount).Error; countError != nil {
		t.Fatalf("count notifications error: %v", countError)
	}
	if notificationCount != 0 {
		t.Fatalf("expected zero notifications stored, got %d", notificationCount)
	}
}

func TestProcessRetriesFailsSmsWhenSenderDisabled(t *testing.T) {
	t.Helper()

	database := openIsolatedDatabase(t)
	serviceInstance := &notificationServiceImpl{
		database:         database,
		logger:           slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
		emailSender:      &stubEmailSender{},
		smsSender:        nil,
		maxRetries:       3,
		retryIntervalSec: 1,
		smsEnabled:       false,
	}

	now := time.Now().UTC()
	smsNotification := model.Notification{
		NotificationID:   "notif-sms-disabled",
		NotificationType: model.NotificationSMS,
		Recipient:        "+15555555555",
		Message:          "Body",
		Status:           model.StatusQueued,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if createError := model.CreateNotification(context.Background(), database, &smsNotification); createError != nil {
		t.Fatalf("create notification error: %v", createError)
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("unexpected panic during retry processing: %v", recovered)
		}
	}()

	serviceInstance.processRetries(context.Background())

	updatedNotification, fetchError := model.GetNotificationByID(context.Background(), database, "notif-sms-disabled")
	if fetchError != nil {
		t.Fatalf("fetch notification error: %v", fetchError)
	}

	if updatedNotification.Status != model.StatusFailed {
		t.Fatalf("expected status failed, got %s", updatedNotification.Status)
	}
}

type stubEmailSender struct {
	callCount int
}

func (sender *stubEmailSender) SendEmail(_ context.Context, _ string, _ string, _ string) error {
	sender.callCount++
	return nil
}

type stubSmsSender struct {
	callCount int
}

func (sender *stubSmsSender) SendSms(_ context.Context, _ string, _ string) (string, error) {
	sender.callCount++
	return "queued", nil
}

func openIsolatedDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	databaseName := time.Now().UTC().Format("20060102150405.000000000")
	database, openError := gorm.Open(sqlite.Open("file:"+databaseName+"?mode=memory&cache=shared"), &gorm.Config{})
	if openError != nil {
		t.Fatalf("sqlite open error: %v", openError)
	}
	if migrateError := database.AutoMigrate(&model.Notification{}); migrateError != nil {
		t.Fatalf("migration error: %v", migrateError)
	}
	return database
}

func durationPointer(value time.Duration) *time.Duration {
	return &value
}
