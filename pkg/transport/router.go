package transport

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/temirov/notify/pkg/config"
	"github.com/temirov/notify/pkg/transport/handler"
	"gorm.io/gorm"
)

const (
	RouteNotifications      = "/notifications"
	RouteNotificationsExact = "/notifications/"
)

func CreateRouter(db *gorm.DB, logger *slog.Logger, authToken string, cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(RouteNotifications, authMiddleware(handler.CreateNotificationHandler(db, logger), authToken))
	mux.HandleFunc(RouteNotificationsExact, authMiddleware(handler.GetNotificationHandler(db, logger), authToken))

	return mux
}

func authMiddleware(next http.HandlerFunc, requiredToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if requiredToken == "" {
			// no auth if token not set
			next.ServeHTTP(w, r)
			return
		}
		// Expect "Authorization: Bearer <token>"
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenValue := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenValue != requiredToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}
