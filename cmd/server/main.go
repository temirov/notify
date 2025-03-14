package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/temirov/notify/pkg/config"
	"github.com/temirov/notify/pkg/db"
	"github.com/temirov/notify/pkg/grpcapi"
	"github.com/temirov/notify/pkg/logging"
	"github.com/temirov/notify/pkg/model"
	"github.com/temirov/notify/pkg/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log/slog"
)

// notificationServiceServer implements grpcapi.NotificationServiceServer.
type notificationServiceServer struct {
	grpcapi.UnimplementedNotificationServiceServer
	notificationService service.NotificationService
	logger              *slog.Logger
}

func (server *notificationServiceServer) SendNotification(ctx context.Context, req *grpcapi.NotificationRequest) (*grpcapi.NotificationResponse, error) {
	var internalType model.NotificationType
	switch req.NotificationType {
	case grpcapi.NotificationType_EMAIL:
		internalType = model.NotificationEmail
	case grpcapi.NotificationType_SMS:
		internalType = model.NotificationSMS
	default:
		server.logger.Error("Unsupported notification type", "type", req.NotificationType)
		return nil, fmt.Errorf("unsupported notification type: %v", req.NotificationType)
	}

	modelRequest := model.NotificationRequest{
		NotificationType: internalType,
		Recipient:        req.Recipient,
		Subject:          req.Subject,
		Message:          req.Message,
	}

	modelResponse, err := server.notificationService.SendNotification(ctx, modelRequest)
	if err != nil {
		server.logger.Error("Service SendNotification error", "error", err)
		return nil, err
	}

	return mapModelToGrpcResponse(modelResponse), nil
}

func (server *notificationServiceServer) GetNotificationStatus(ctx context.Context, req *grpcapi.GetNotificationStatusRequest) (*grpcapi.NotificationResponse, error) {
	if req.NotificationId == "" {
		server.logger.Error("Missing notification ID")
		return nil, fmt.Errorf("missing notification ID")
	}

	modelResponse, err := server.notificationService.GetNotificationStatus(ctx, req.NotificationId)
	if err != nil {
		server.logger.Error("Service GetNotificationStatus error", "error", err)
		return nil, err
	}
	return mapModelToGrpcResponse(modelResponse), nil
}

// mapModelToGrpcResponse converts a model.NotificationResponse to a grpcapi.NotificationResponse.
func mapModelToGrpcResponse(modelResp model.NotificationResponse) *grpcapi.NotificationResponse {
	var grpcNotifType grpcapi.NotificationType
	switch modelResp.NotificationType {
	case model.NotificationEmail:
		grpcNotifType = grpcapi.NotificationType_EMAIL
	case model.NotificationSMS:
		grpcNotifType = grpcapi.NotificationType_SMS
	default:
		grpcNotifType = grpcapi.NotificationType_EMAIL
	}

	var grpcStatus grpcapi.Status
	switch modelResp.Status {
	case model.StatusQueued:
		grpcStatus = grpcapi.Status_QUEUED
	case model.StatusSent:
		grpcStatus = grpcapi.Status_SENT
	case model.StatusFailed:
		grpcStatus = grpcapi.Status_FAILED
	default:
		grpcStatus = grpcapi.Status_UNKNOWN
	}

	return &grpcapi.NotificationResponse{
		NotificationId:    modelResp.NotificationID,
		NotificationType:  grpcNotifType,
		Recipient:         modelResp.Recipient,
		Subject:           modelResp.Subject,
		Message:           modelResp.Message,
		Status:            grpcStatus,
		ProviderMessageId: modelResp.ProviderMessageID,
		RetryCount:        int32(modelResp.RetryCount),
		CreatedAt:         modelResp.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         modelResp.UpdatedAt.Format(time.RFC3339),
	}
}

func main() {
	configuration, configErr := config.LoadConfig()
	if configErr != nil {
		fallbackLogger := logging.NewLogger("INFO")
		for _, errMsg := range strings.Split(configErr.Error(), ", ") {
			fallbackLogger.Error("Configuration error", "detail", errMsg)
		}
		os.Exit(1)
	}

	mainLogger := logging.NewLogger(configuration.LogLevel)
	mainLogger.Info("Starting gRPC Notification Server on :50051")

	databaseInstance, dbErr := db.InitDB(configuration.DatabasePath, mainLogger)
	if dbErr != nil {
		mainLogger.Error("Failed to initialize DB", "error", dbErr)
		os.Exit(1)
	}

	notificationSvc := service.NewNotificationService(databaseInstance, mainLogger, configuration)

	// Start the background retry worker.
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()
	go notificationSvc.StartRetryWorker(workerCtx)

	// Set up gRPC server with an authentication interceptor.
	authInterceptor := func(logger *slog.Logger, requiredToken string) grpc.UnaryServerInterceptor {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				logger.Error("Missing metadata in gRPC request")
				return nil, fmt.Errorf("missing metadata")
			}
			authHeaders := md["authorization"]
			if len(authHeaders) == 0 {
				logger.Error("Missing authorization header")
				return nil, fmt.Errorf("missing authorization header")
			}
			if !strings.HasPrefix(authHeaders[0], "Bearer ") {
				logger.Error("Invalid authorization header format")
				return nil, fmt.Errorf("invalid authorization header")
			}
			token := strings.TrimPrefix(authHeaders[0], "Bearer ")
			if token != configuration.GRPCAuthToken {
				logger.Error("Invalid token provided", "got", token)
				return nil, fmt.Errorf("invalid token")
			}
			return handler(ctx, req)
		}
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor(mainLogger, configuration.GRPCAuthToken)))
	grpcapi.RegisterNotificationServiceServer(grpcServer, &notificationServiceServer{
		notificationService: notificationSvc,
		logger:              mainLogger,
	})

	listener, listenErr := net.Listen("tcp", ":50051")
	if listenErr != nil {
		mainLogger.Error("Failed to listen on :50051", "error", listenErr)
		os.Exit(1)
	}
	mainLogger.Info("gRPC server listening on :50051")

	if serveErr := grpcServer.Serve(listener); serveErr != nil {
		mainLogger.Error("gRPC server crashed", "error", serveErr)
		os.Exit(1)
	}
}
