package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/temirov/notify/pkg/model"
	"gorm.io/gorm"
)

type NotificationRequest struct {
	NotificationType model.NotificationType `json:"notification_type"`
	Recipient        string                 `json:"recipient"`
	Subject          string                 `json:"subject"`
	Message          string                 `json:"message"`
}

type NotificationResponse struct {
	NotificationID    string                 `json:"notification_id"`
	NotificationType  model.NotificationType `json:"notification_type"`
	Recipient         string                 `json:"recipient"`
	Subject           string                 `json:"subject,omitempty"`
	Message           string                 `json:"message"`
	Status            string                 `json:"status"`
	ProviderMessageID string                 `json:"provider_message_id"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	RetryCount        int                    `json:"retry_count"`
}

// CreateNotificationHandler for POST /notifications
func CreateNotificationHandler(db *gorm.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req NotificationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if req.NotificationType == "" || req.Recipient == "" || req.Message == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		userFacingID := uuid.New().String()
		notif := model.Notification{
			NotificationID:   userFacingID,
			NotificationType: req.NotificationType,
			Recipient:        req.Recipient,
			Subject:          req.Subject,
			Message:          req.Message,
			Status:           model.StatusQueued,
		}

		if err := db.Create(&notif).Error; err != nil {
			logger.Error("Failed to create notification", "error", err)
			http.Error(w, "Failed to create notification", http.StatusInternalServerError)
			return
		}

		resp := NotificationResponse{
			NotificationID:    notif.NotificationID,
			NotificationType:  notif.NotificationType,
			Recipient:         notif.Recipient,
			Subject:           notif.Subject,
			Message:           notif.Message,
			Status:            notif.Status,
			ProviderMessageID: notif.ProviderMessageID,
			CreatedAt:         notif.CreatedAt,
			UpdatedAt:         notif.UpdatedAt,
			RetryCount:        notif.RetryCount,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// GetNotificationHandler for GET /notifications/<id>
func GetNotificationHandler(db *gorm.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// The route is "/notifications/" so the next part is the ID
		pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/notifications/"), "/")
		if len(pathParts) < 1 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		notificationID := pathParts[0]
		if notificationID == "" {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		var notif model.Notification
		err := db.Where("notification_id = ?", notificationID).First(&notif).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Notification not found", http.StatusNotFound)
			return
		} else if err != nil {
			logger.Error("Failed to retrieve notification", "error", err)
			http.Error(w, "Failed to get notification", http.StatusInternalServerError)
			return
		}

		resp := NotificationResponse{
			NotificationID:    notif.NotificationID,
			NotificationType:  notif.NotificationType,
			Recipient:         notif.Recipient,
			Subject:           notif.Subject,
			Message:           notif.Message,
			Status:            notif.Status,
			ProviderMessageID: notif.ProviderMessageID,
			CreatedAt:         notif.CreatedAt,
			UpdatedAt:         notif.UpdatedAt,
			RetryCount:        notif.RetryCount,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
