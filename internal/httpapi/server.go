package httpapi

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/temirov/pinguin/internal/model"
	"github.com/temirov/pinguin/internal/service"
	sessionvalidator "github.com/tyemirov/tauth/pkg/sessionvalidator"
	"gorm.io/gorm"
	"log/slog"
)

const (
	contextKeyClaims = "auth_claims"
	defaultTimeout   = 5 * time.Second
)

// SessionValidator exposes the subset of validator behaviour we depend on.
type SessionValidator interface {
	ValidateRequest(request *http.Request) (*sessionvalidator.Claims, error)
}

// Config captures all inputs required to construct the HTTP server.
type Config struct {
	ListenAddr           string
	StaticRoot           string
	AllowedOrigins       []string
	SessionValidator     SessionValidator
	NotificationService  service.NotificationService
	Logger               *slog.Logger
	ReadHeaderTimeout    time.Duration
	ShutdownGraceTimeout time.Duration
}

// Server hosts authenticated HTTP endpoints and static assets for the UI.
type Server struct {
	config     Config
	httpServer *http.Server
	logger     *slog.Logger
}

// NewServer wires Gin, middleware, and handlers for the HTTP API.
func NewServer(cfg Config) (*Server, error) {
	if strings.TrimSpace(cfg.ListenAddr) == "" {
		return nil, errors.New("httpapi: listen address is required")
	}
	if cfg.SessionValidator == nil {
		return nil, errors.New("httpapi: session validator is required")
	}
	if cfg.NotificationService == nil {
		return nil, errors.New("httpapi: notification service is required")
	}
	if cfg.Logger == nil {
		return nil, errors.New("httpapi: logger is required")
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(requestLogger(cfg.Logger))

	engine.GET("/healthz", func(contextGin *gin.Context) {
		contextGin.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	if cfg.StaticRoot != "" {
		engine.StaticFS("/", gin.Dir(filepath.Clean(cfg.StaticRoot), false))
	}

	engine.Use(buildCORS(cfg.AllowedOrigins))

	protected := engine.Group("/api")
	protected.Use(sessionMiddleware(cfg.SessionValidator))

	handler := newNotificationHandler(cfg.NotificationService, cfg.Logger)
	protected.GET("/notifications", handler.listNotifications)
	protected.PATCH("/notifications/:id/schedule", handler.rescheduleNotification)
	protected.POST("/notifications/:id/cancel", handler.cancelNotification)

	httpServer := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           engine,
		ReadHeaderTimeout: pickDuration(cfg.ReadHeaderTimeout, defaultTimeout),
	}

	return &Server{
		config:     cfg,
		httpServer: httpServer,
		logger:     cfg.Logger,
	}, nil
}

// Start begins serving HTTP traffic.
func (server *Server) Start() error {
	err := server.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully terminates the HTTP server.
func (server *Server) Shutdown(ctx context.Context) error {
	timeout := pickDuration(server.config.ShutdownGraceTimeout, defaultTimeout)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return server.httpServer.Shutdown(ctx)
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(contextGin *gin.Context) {
		started := time.Now()
		contextGin.Next()
		logger.Info(
			"http_request_completed",
			"method", contextGin.Request.Method,
			"path", contextGin.Request.URL.Path,
			"status", contextGin.Writer.Status(),
			"duration_ms", time.Since(started).Milliseconds(),
		)
	}
}

func buildCORS(allowedOrigins []string) gin.HandlerFunc {
	if len(allowedOrigins) == 0 {
		cfg := cors.Config{
			AllowAllOrigins:  true,
			AllowHeaders:     []string{"Content-Type", "X-Requested-With"},
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodOptions},
			AllowCredentials: true,
		}
		return cors.New(cfg)
	}
	cfg := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowHeaders:     []string{"Content-Type", "X-Requested-With"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodOptions},
		AllowCredentials: true,
	}
	return cors.New(cfg)
}

func sessionMiddleware(validator SessionValidator) gin.HandlerFunc {
	return func(contextGin *gin.Context) {
		claims, err := validator.ValidateRequest(contextGin.Request)
		if err != nil {
			contextGin.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		contextGin.Set(contextKeyClaims, claims)
		contextGin.Next()
	}
}

type notificationHandler struct {
	service service.NotificationService
	logger  *slog.Logger
}

func newNotificationHandler(svc service.NotificationService, logger *slog.Logger) *notificationHandler {
	return &notificationHandler{service: svc, logger: logger}
}

func (handler *notificationHandler) listNotifications(contextGin *gin.Context) {
	statusFilters := contextGin.QueryArray("status")
	filter := model.NotificationListFilters{
		Statuses: parseStatusFilters(statusFilters),
	}
	responses, err := handler.service.ListNotifications(contextGin.Request.Context(), filter)
	if err != nil {
		handler.writeError(contextGin, err)
		return
	}
	contextGin.JSON(http.StatusOK, gin.H{"notifications": responses})
}

func (handler *notificationHandler) rescheduleNotification(contextGin *gin.Context) {
	var payload struct {
		ScheduledTime string `json:"scheduled_time"`
	}
	if err := contextGin.ShouldBindJSON(&payload); err != nil {
		contextGin.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if strings.TrimSpace(payload.ScheduledTime) == "" {
		contextGin.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_time is required"})
		return
	}
	parsedTime, err := time.Parse(time.RFC3339, payload.ScheduledTime)
	if err != nil {
		contextGin.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_time must be RFC3339"})
		return
	}
	response, svcErr := handler.service.RescheduleNotification(contextGin.Request.Context(), contextGin.Param("id"), parsedTime)
	if svcErr != nil {
		handler.writeError(contextGin, svcErr)
		return
	}
	contextGin.JSON(http.StatusOK, response)
}

func (handler *notificationHandler) cancelNotification(contextGin *gin.Context) {
	response, err := handler.service.CancelNotification(contextGin.Request.Context(), contextGin.Param("id"))
	if err != nil {
		handler.writeError(contextGin, err)
		return
	}
	contextGin.JSON(http.StatusOK, response)
}

func (handler *notificationHandler) writeError(contextGin *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotificationNotEditable):
		contextGin.JSON(http.StatusConflict, gin.H{"error": "notification can only be edited while queued"})
	case errors.Is(err, service.ErrScheduleInPast):
		contextGin.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_time must be in the future"})
	case errors.Is(err, model.ErrNotificationNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		contextGin.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
	default:
		handler.logger.Error("http_handler_error", "error", err)
		contextGin.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func parseStatusFilters(values []string) []model.NotificationStatus {
	if len(values) == 0 {
		return nil
	}
	unique := make(map[model.NotificationStatus]struct{}, len(values))
	var statuses []model.NotificationStatus
	for _, raw := range values {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		status := model.NotificationStatus(strings.ToLower(trimmed))
		if _, exists := unique[status]; exists {
			continue
		}
		unique[status] = struct{}{}
		statuses = append(statuses, status)
	}
	return statuses
}

func pickDuration(candidate time.Duration, fallback time.Duration) time.Duration {
	if candidate <= 0 {
		return fallback
	}
	return candidate
}
