package client

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/temirov/pinguin/pkg/config"
	"github.com/temirov/pinguin/pkg/grpcapi"
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
	config     config.Config
}

// NewNotificationClient creates a new NotificationClient.
// It reads GRPC_SERVER_ADDR (defaulting to "localhost:50051") and uses the provided config for timeouts.
func NewNotificationClient(logger *slog.Logger, cfg config.Config) (*NotificationClient, error) {
	serverAddress := os.Getenv("GRPC_SERVER_ADDR")
	if serverAddress == "" {
		serverAddress = "localhost:50051"
	}

	// Create a context with timeout for dialing
	dialCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnectionTimeoutSec)*time.Second)
	defer cancel()

	// Use grpc.NewClient with dial options (replaces deprecated DialContext)
	conn, err := grpc.NewClient(
		serverAddress,
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			// Use the standard dialer with the provided context
			dialer := &net.Dialer{}
			return dialer.DialContext(ctx, "tcp", addr)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// No need for WithBlock; grpc.NewClient blocks until the connection is ready or fails
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	// Check if the context was canceled before proceeding
	if dialCtx.Err() != nil {
		conn.Close()
		return nil, fmt.Errorf("dialing gRPC server timed out: %w", dialCtx.Err())
	}

	grpcClient := grpcapi.NewNotificationServiceClient(conn)
	return &NotificationClient{
		conn:       conn,
		grpcClient: grpcClient,
		authToken:  cfg.GRPCAuthToken,
		logger:     logger,
		config:     cfg,
	}, nil
}

// Close closes the underlying gRPC connection.
func (clientInstance *NotificationClient) Close() error {
	return clientInstance.conn.Close()
}

// SendNotification sends the provided NotificationRequest using the provided context
// and appends the required authorization header.
// Note: Errors are simply returned without logging here.
func (clientInstance *NotificationClient) SendNotification(ctx context.Context, req *grpcapi.NotificationRequest) (*grpcapi.NotificationResponse, error) {
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(clientInstance.config.OperationTimeoutSec)*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(clientInstance.config.OperationTimeoutSec)*time.Second)
	defer cancel()

	resp, err := clientInstance.SendNotification(ctx, req)
	if err != nil {
		clientInstance.logger.Error("SendNotification failed", "error", err)
		return nil, err
	}

	const pollInterval = 2 * time.Second
	pollTimeout := time.Duration(clientInstance.config.OperationTimeoutSec) * time.Second
	startTime := time.Now()

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
