package service

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/temirov/notify/pkg/model"
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
