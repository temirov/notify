package model

import "time"

type NotificationType string

const (
	NotificationEmail NotificationType = "email"
	NotificationSMS   NotificationType = "sms"
)

// Status constants
const (
	StatusQueued  = "queued"
	StatusSent    = "sent"
	StatusFailed  = "failed"
	StatusUnknown = "unknown"
)

// Notification is our GORM model
type Notification struct {
	ID                uint   `gorm:"primaryKey"`
	NotificationID    string `gorm:"uniqueIndex"`
	NotificationType  NotificationType
	Recipient         string
	Subject           string
	Message           string
	ProviderMessageID string
	Status            string
	RetryCount        int
	LastAttemptedAt   time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
