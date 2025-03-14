package transport

import (
	"log/slog"
	"net/http"

	"github.com/temirov/notify/pkg/transport/handler"
	"gorm.io/gorm"
)

func CreateRouter(db *gorm.DB, logger *slog.Logger, authToken string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/notifications", authMiddleware(
		handler.CreateNotificationHandler(db, logger),
		authToken,
	))
	mux.HandleFunc("/notifications/", authMiddleware(
		handler.GetNotificationHandler(db, logger),
		authToken,
	))

	return mux
}

func authMiddleware(next http.HandlerFunc, requiredToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if requiredToken == "" {
			next.ServeHTTP(w, r)
			return
		}
		authHeader := r.Header.Get("Authorization")
		bearerPrefix := "Bearer "
		if len(authHeader) < len(bearerPrefix) ||
			authHeader[:len(bearerPrefix)] != bearerPrefix {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenValue := authHeader[len(bearerPrefix):]
		if tokenValue != requiredToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}
