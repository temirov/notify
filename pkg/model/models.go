package model

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// NotificationType enumerations: "email" or "sms".
type NotificationType string

const (
	NotificationEmail NotificationType = "email"
	NotificationSMS   NotificationType = "sms"
)

// Status constants used for the Notification model.
const (
	StatusQueued  = "queued"
	StatusSent    = "sent"
	StatusFailed  = "failed"
	StatusUnknown = "unknown"
)

// Notification is our main model in the DB, with GORM & JSON tags.
// You can return this directly via JSON or create a separate struct if you like.
type Notification struct {
	ID                uint             `json:"-" gorm:"primaryKey"`
	NotificationID    string           `json:"notification_id" gorm:"uniqueIndex"`
	NotificationType  NotificationType `json:"notification_type"`
	Recipient         string           `json:"recipient"`
	Subject           string           `json:"subject,omitempty"`
	Message           string           `json:"message"`
	ProviderMessageID string           `json:"provider_message_id"`
	Status            string           `json:"status"`
	RetryCount        int              `json:"retry_count"`
	LastAttemptedAt   time.Time        `json:"last_attempted_at"`
	ScheduledFor      *time.Time       `json:"scheduled_for"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// NotificationRequest represents the incoming request payload (REST/gRPC).
type NotificationRequest struct {
	NotificationType NotificationType `json:"notification_type"`
	Recipient        string           `json:"recipient"`
	Subject          string           `json:"subject,omitempty"`
	Message          string           `json:"message"`
	ScheduledFor     *time.Time       `json:"scheduled_for,omitempty"`
}

// NotificationResponse is what you'll return to the client.
// You could also return the Notification itself, but some prefer a separate shape.
type NotificationResponse struct {
	NotificationID    string           `json:"notification_id"`
	NotificationType  NotificationType `json:"notification_type"`
	Recipient         string           `json:"recipient"`
	Subject           string           `json:"subject,omitempty"`
	Message           string           `json:"message"`
	Status            string           `json:"status"`
	ProviderMessageID string           `json:"provider_message_id"`
	RetryCount        int              `json:"retry_count"`
	ScheduledFor      *time.Time       `json:"scheduled_for,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// NewNotification constructs a ready-to-insert DB Notification from a request, defaulting status=queued.
func NewNotification(notificationID string, req NotificationRequest) Notification {
	now := time.Now().UTC()
	return Notification{
		NotificationID:   notificationID,
		NotificationType: req.NotificationType,
		Recipient:        req.Recipient,
		Subject:          req.Subject,
		Message:          req.Message,
		Status:           StatusQueued,
		ScheduledFor:     nil,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// NewNotificationResponse translates a DB Notification to a response shape.
func NewNotificationResponse(n Notification) NotificationResponse {
	var scheduledFor *time.Time
	if n.ScheduledFor != nil {
		normalizedScheduled := n.ScheduledFor.UTC()
		scheduledFor = &normalizedScheduled
	}
	return NotificationResponse{
		NotificationID:    n.NotificationID,
		NotificationType:  n.NotificationType,
		Recipient:         n.Recipient,
		Subject:           n.Subject,
		Message:           n.Message,
		Status:            n.Status,
		ProviderMessageID: n.ProviderMessageID,
		RetryCount:        n.RetryCount,
		ScheduledFor:      scheduledFor,
		CreatedAt:         n.CreatedAt,
		UpdatedAt:         n.UpdatedAt,
	}
}

// ====================== DB CRUD METHODS ====================== //

func CreateNotification(ctx context.Context, db *gorm.DB, n *Notification) error {
	return db.WithContext(ctx).Create(n).Error
}

func GetNotificationByID(ctx context.Context, db *gorm.DB, notificationID string) (*Notification, error) {
	var notif Notification
	err := db.WithContext(ctx).Where("notification_id = ?", notificationID).First(&notif).Error
	if err != nil {
		return nil, err
	}
	return &notif, nil
}

func SaveNotification(ctx context.Context, db *gorm.DB, n *Notification) error {
	return db.WithContext(ctx).Save(n).Error
}

func GetQueuedOrFailedNotifications(ctx context.Context, db *gorm.DB, maxRetries int, currentTime time.Time) ([]Notification, error) {
	var notifications []Notification
	err := db.WithContext(ctx).
		Where("(status = ? OR status = ?) AND retry_count < ?",
			StatusQueued, StatusFailed, maxRetries).
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func MustGetNotificationByID(ctx context.Context, db *gorm.DB, notificationID string) (*Notification, error) {
	n, err := GetNotificationByID(ctx, db, notificationID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification not found: %s", notificationID)
		}
		return nil, err
	}
	return n, nil
}
