package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/temirov/pinguin/internal/model"
	"github.com/temirov/pinguin/internal/service"
	sessionvalidator "github.com/tyemirov/tauth/pkg/sessionvalidator"
	"log/slog"
)

func TestListNotificationsRequiresAuth(t *testing.T) {
	t.Helper()

	server := newTestHTTPServer(t, &stubNotificationService{}, &stubValidator{err: errors.New("unauthorized")})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/notifications", nil)

	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestListNotificationsReturnsData(t *testing.T) {
	t.Helper()

	stubSvc := &stubNotificationService{
		listResponse: []model.NotificationResponse{
			{NotificationID: "queued", Status: model.StatusQueued},
			{NotificationID: "errored", Status: model.StatusErrored},
		},
	}
	server := newTestHTTPServer(t, stubSvc, &stubValidator{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/notifications?status=queued&status=errored", nil)

	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var payload struct {
		Notifications []model.NotificationResponse `json:"notifications"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response decode error: %v", err)
	}
	if len(payload.Notifications) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(payload.Notifications))
	}
}

func TestRescheduleValidation(t *testing.T) {
	t.Helper()

	server := newTestHTTPServer(t, &stubNotificationService{}, &stubValidator{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/notifications/notif-1/schedule", bytes.NewBufferString(`{}`))
	request.Header.Set("Content-Type", "application/json")

	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestCancelNotificationErrorMapping(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name         string
		cancelError  error
		expectedCode int
	}{
		{
			name:         "Conflict",
			cancelError:  service.ErrNotificationNotEditable,
			expectedCode: http.StatusConflict,
		},
		{
			name:         "NotFound",
			cancelError:  model.ErrNotificationNotFound,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Internal",
			cancelError:  errors.New("boom"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Helper()

			stubSvc := &stubNotificationService{cancelErr: testCase.cancelError}
			server := newTestHTTPServer(t, stubSvc, &stubValidator{})

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/notifications/notif-1/cancel", nil)

			server.httpServer.Handler.ServeHTTP(recorder, request)
			if recorder.Code != testCase.expectedCode {
				t.Fatalf("expected %d, got %d", testCase.expectedCode, recorder.Code)
			}
		})
	}
}

func TestNewServerSupportsStaticRootAfterAPIRoutes(t *testing.T) {
	t.Helper()

	tempDir := t.TempDir()
	assetPath := filepath.Join(tempDir, "app.js")
	if writeErr := os.WriteFile(assetPath, []byte("console.log('ok');"), 0o644); writeErr != nil {
		t.Fatalf("failed to write static file: %v", writeErr)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	server, err := NewServer(Config{
		ListenAddr:          ":0",
		StaticRoot:          tempDir,
		NotificationService: &stubNotificationService{},
		SessionValidator:    &stubValidator{},
		Logger:              logger,
	})
	if err != nil {
		t.Fatalf("server init error: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/app.js", nil)

	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 when serving static content, got %d", recorder.Code)
	}
}

func TestBuildCORSDefaultDisablesCredentials(t *testing.T) {
	t.Helper()

	engine := gin.New()
	engine.Use(buildCORS(nil))
	engine.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	request.Header.Set("Origin", "https://evil.example")

	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("expected no credentials header, got %q", got)
	}
	if origin := recorder.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Fatalf("expected wildcard allow origin, got %q", origin)
	}
}

func TestBuildCORSEmitsCredentialsForExplicitAllowList(t *testing.T) {
	t.Helper()

	const allowedOrigin = "https://app.example"

	engine := gin.New()
	engine.Use(buildCORS([]string{allowedOrigin}))
	engine.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	request.Header.Set("Origin", allowedOrigin)

	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected credentials header, got %q", got)
	}
	if origin := recorder.Header().Get("Access-Control-Allow-Origin"); origin != allowedOrigin {
		t.Fatalf("expected allow origin %q, got %q", allowedOrigin, origin)
	}
}

func newTestHTTPServer(t *testing.T, svc service.NotificationService, validator SessionValidator) *Server {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	server, err := NewServer(Config{
		ListenAddr:          ":0",
		NotificationService: svc,
		SessionValidator:    validator,
		Logger:              logger,
	})
	if err != nil {
		t.Fatalf("server init error: %v", err)
	}
	return server
}

type stubValidator struct {
	err error
}

func (validator *stubValidator) ValidateRequest(_ *http.Request) (*sessionvalidator.Claims, error) {
	if validator.err != nil {
		return nil, validator.err
	}
	return &sessionvalidator.Claims{UserEmail: "user@example.com"}, nil
}

type stubNotificationService struct {
	listResponse       []model.NotificationResponse
	listErr            error
	rescheduleResponse model.NotificationResponse
	rescheduleErr      error
	cancelResponse     model.NotificationResponse
	cancelErr          error
}

func (stub *stubNotificationService) SendNotification(context.Context, model.NotificationRequest) (model.NotificationResponse, error) {
	return model.NotificationResponse{}, errors.New("not implemented")
}

func (stub *stubNotificationService) GetNotificationStatus(context.Context, string) (model.NotificationResponse, error) {
	return model.NotificationResponse{}, errors.New("not implemented")
}

func (stub *stubNotificationService) ListNotifications(context.Context, model.NotificationListFilters) ([]model.NotificationResponse, error) {
	return stub.listResponse, stub.listErr
}

func (stub *stubNotificationService) RescheduleNotification(context.Context, string, time.Time) (model.NotificationResponse, error) {
	return stub.rescheduleResponse, stub.rescheduleErr
}

func (stub *stubNotificationService) CancelNotification(context.Context, string) (model.NotificationResponse, error) {
	if stub.cancelErr != nil {
		return model.NotificationResponse{}, stub.cancelErr
	}
	return stub.cancelResponse, nil
}

func (stub *stubNotificationService) StartRetryWorker(context.Context) {}
