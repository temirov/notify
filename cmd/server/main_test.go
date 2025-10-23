package main

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/temirov/pinguin/pkg/client"
	"github.com/temirov/pinguin/pkg/config"
	"github.com/temirov/pinguin/pkg/grpcapi"
	"github.com/temirov/pinguin/pkg/model"
	"github.com/temirov/pinguin/pkg/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log/slog"
)

func TestNotificationServerHandlesClientRequests(t *testing.T) {
	t.Helper()

	authToken := "unit-token"
	notificationService := &stubNotificationService{
		sendResponse: model.NotificationResponse{
			NotificationID:   "notif-123",
			NotificationType: model.NotificationEmail,
			Recipient:        "user@example.com",
			Message:          "Hello",
			Status:           model.StatusSent,
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		},
		statusResponses: []model.NotificationResponse{
			{
				NotificationID:   "notif-123",
				NotificationType: model.NotificationEmail,
				Recipient:        "user@example.com",
				Message:          "Hello",
				Status:           model.StatusSent,
				CreatedAt:        time.Now().UTC(),
				UpdatedAt:        time.Now().UTC(),
			},
		},
	}

	serverAddress, shutdown := startTestNotificationServer(t, notificationService, authToken)
	defer shutdown()

	t.Setenv("GRPC_SERVER_ADDR", serverAddress)

	clientConfig := config.Config{
		ConnectionTimeoutSec: 5,
		OperationTimeoutSec:  5,
		GRPCAuthToken:        authToken,
	}
	clientLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	notificationClient, clientError := client.NewNotificationClient(clientLogger, clientConfig)
	if clientError != nil {
		t.Fatalf("create client error: %v", clientError)
	}
	defer notificationClient.Close()

	grpcRequest := &grpcapi.NotificationRequest{
		NotificationType: grpcapi.NotificationType_EMAIL,
		Recipient:        "user@example.com",
		Subject:          "Unit",
		Message:          "Hello",
	}

	sendResponse, sendError := notificationClient.SendNotification(context.Background(), grpcRequest)
	if sendError != nil {
		t.Fatalf("send notification error: %v", sendError)
	}
	if sendResponse.NotificationId != "notif-123" {
		t.Fatalf("unexpected notification id %s", sendResponse.NotificationId)
	}

	statusResponse, statusError := notificationClient.GetNotificationStatus("notif-123")
	if statusError != nil {
		t.Fatalf("status retrieval error: %v", statusError)
	}
	if statusResponse.Status != grpcapi.Status_SENT {
		t.Fatalf("unexpected status %v", statusResponse.Status)
	}

	waitResponse, waitError := notificationClient.SendNotificationAndWait(grpcRequest)
	if waitError != nil {
		t.Fatalf("send and wait error: %v", waitError)
	}
	if waitResponse.Status != grpcapi.Status_SENT {
		t.Fatalf("unexpected wait status %v", waitResponse.Status)
	}

	if len(notificationService.sendCalls) != 2 {
		t.Fatalf("expected two send calls, got %d", len(notificationService.sendCalls))
	}
	if len(notificationService.statusCalls) == 0 {
		t.Fatalf("expected status calls")
	}
}

func startTestNotificationServer(t *testing.T, svc service.NotificationService, token string) (string, func()) {
	t.Helper()

	listener, listenError := net.Listen("tcp", "127.0.0.1:0")
	if listenError != nil {
		t.Fatalf("listen error: %v", listenError)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(buildAuthInterceptor(logger, token)))
	grpcapi.RegisterNotificationServiceServer(grpcServer, &notificationServiceServer{
		notificationService: svc,
		logger:              logger,
	})

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	shutdown := func() {
		grpcServer.Stop()
		listener.Close()
	}

	return listener.Addr().String(), shutdown
}

func buildAuthInterceptor(logger *slog.Logger, expectedToken string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		metadataValues, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Error("missing metadata")
			return nil, context.Canceled
		}
		headers := metadataValues.Get("authorization")
		if len(headers) == 0 {
			logger.Error("missing authorization header")
			return nil, context.Canceled
		}
		if headers[0] != "Bearer "+expectedToken {
			logger.Error("invalid token", "value", headers[0])
			return nil, context.Canceled
		}
		return handler(ctx, req)
	}
}

type stubNotificationService struct {
	mutex           sync.Mutex
	sendCalls       []model.NotificationRequest
	statusCalls     []string
	sendResponse    model.NotificationResponse
	statusResponses []model.NotificationResponse
}

func (stub *stubNotificationService) SendNotification(ctx context.Context, request model.NotificationRequest) (model.NotificationResponse, error) {
	stub.mutex.Lock()
	defer stub.mutex.Unlock()
	stub.sendCalls = append(stub.sendCalls, request)
	return stub.sendResponse, nil
}

func (stub *stubNotificationService) GetNotificationStatus(ctx context.Context, notificationID string) (model.NotificationResponse, error) {
	stub.mutex.Lock()
	defer stub.mutex.Unlock()
	stub.statusCalls = append(stub.statusCalls, notificationID)
	if len(stub.statusResponses) == 0 {
		return stub.sendResponse, nil
	}
	response := stub.statusResponses[0]
	stub.statusResponses = stub.statusResponses[1:]
	return response, nil
}

func (stub *stubNotificationService) StartRetryWorker(ctx context.Context) {}
