package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/temirov/notify/pkg/grpcapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log/slog"
)

// NotificationClient is a thin wrapper over the gRPC NotificationService client.
type NotificationClient struct {
	conn       *grpc.ClientConn
	grpcClient grpcapi.NotificationServiceClient
	authToken  string
	logger     *slog.Logger
}

// NewNotificationClient creates a new NotificationClient.
// It reads GRPC_SERVER_ADDR (defaulting to "localhost:50051") and GRPC_AUTH_TOKEN from environment variables.
func NewNotificationClient(logger *slog.Logger) (*NotificationClient, error) {
	serverAddress := os.Getenv("GRPC_SERVER_ADDR")
	if serverAddress == "" {
		serverAddress = "localhost:50051"
	}

	authToken := os.Getenv("GRPC_AUTH_TOKEN")
	if authToken == "" {
		return nil, errors.New("GRPC_AUTH_TOKEN environment variable is not set")
	}

	dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	grpcClient := grpcapi.NewNotificationServiceClient(conn)
	return &NotificationClient{
		conn:       conn,
		grpcClient: grpcClient,
		authToken:  authToken,
		logger:     logger,
	}, nil
}

// Close closes the underlying gRPC connection.
func (clientInstance *NotificationClient) Close() error {
	return clientInstance.conn.Close()
}

// SendNotification sends the provided NotificationRequest using a context with timeout
// and appends the required authorization header.
// Note: Errors are simply returned without logging here.
func (clientInstance *NotificationClient) SendNotification(req *grpcapi.NotificationRequest) (*grpcapi.NotificationResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+clientInstance.authToken)
	resp, err := clientInstance.grpcClient.SendNotification(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetNotificationStatus retrieves the status of a notification by its ID.
// Note: Errors are simply returned without logging here.
func (clientInstance *NotificationClient) GetNotificationStatus(notificationID string) (*grpcapi.NotificationResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+clientInstance.authToken)
	req := &grpcapi.GetNotificationStatusRequest{
		NotificationId: notificationID,
	}
	resp, err := clientInstance.grpcClient.GetNotificationStatus(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// SendNotificationAndWait sends a notification and then polls for its final status until it is SENT or FAILED.
// It logs errors only here, so duplicate logging is avoided.
func (clientInstance *NotificationClient) SendNotificationAndWait(req *grpcapi.NotificationRequest) (*grpcapi.NotificationResponse, error) {
	resp, err := clientInstance.SendNotification(req)
	if err != nil {
		clientInstance.logger.Error("SendNotification failed", "error", err)
		return nil, err
	}

	// Define polling parameters.
	const pollInterval = 2 * time.Second
	const pollTimeout = 30 * time.Second
	startTime := time.Now()

	// Poll until the status is SENT or FAILED.
	for {
		switch resp.Status {
		case grpcapi.Status_SENT:
			return resp, nil
		case grpcapi.Status_FAILED:
			return resp, fmt.Errorf("notification failed")
		}

		if time.Since(startTime) > pollTimeout {
			return resp, fmt.Errorf("timeout waiting for notification to be sent")
		}

		time.Sleep(pollInterval)
		statusResp, err := clientInstance.GetNotificationStatus(resp.NotificationId)
		if err != nil {
			clientInstance.logger.Error("GetNotificationStatus failed", "notificationID", resp.NotificationId, "error", err)
			return nil, err
		}
		resp = statusResp
	}
}
