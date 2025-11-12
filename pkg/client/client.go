package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/temirov/pinguin/pkg/grpcapi"
	"github.com/temirov/pinguin/pkg/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log/slog"
)

var ErrInvalidSettings = errors.New("invalid_client_settings")

type Settings struct {
	serverAddress     string
	authToken         string
	connectionTimeout time.Duration
	operationTimeout  time.Duration
}

func NewSettings(serverAddress string, authToken string, connectionTimeoutSeconds int, operationTimeoutSeconds int) (Settings, error) {
	address := strings.TrimSpace(serverAddress)
	if address == "" {
		return Settings{}, fmt.Errorf("%w: empty server address", ErrInvalidSettings)
	}
	token := strings.TrimSpace(authToken)
	if token == "" {
		return Settings{}, fmt.Errorf("%w: empty auth token", ErrInvalidSettings)
	}
	if connectionTimeoutSeconds <= 0 {
		return Settings{}, fmt.Errorf("%w: invalid connection timeout %d", ErrInvalidSettings, connectionTimeoutSeconds)
	}
	if operationTimeoutSeconds <= 0 {
		return Settings{}, fmt.Errorf("%w: invalid operation timeout %d", ErrInvalidSettings, operationTimeoutSeconds)
	}
	return Settings{
		serverAddress:     address,
		authToken:         token,
		connectionTimeout: time.Duration(connectionTimeoutSeconds) * time.Second,
		operationTimeout:  time.Duration(operationTimeoutSeconds) * time.Second,
	}, nil
}

func (s Settings) ServerAddress() string {
	return s.serverAddress
}

func (s Settings) AuthToken() string {
	return s.authToken
}

func (s Settings) ConnectionTimeout() time.Duration {
	return s.connectionTimeout
}

func (s Settings) OperationTimeout() time.Duration {
	return s.operationTimeout
}

type NotificationClient struct {
	conn       *grpc.ClientConn
	grpcClient grpcapi.NotificationServiceClient
	authToken  string
	logger     *slog.Logger
	settings   Settings
}

func NewNotificationClient(logger *slog.Logger, settings Settings) (*NotificationClient, error) {
	dialCtx, cancel := context.WithTimeout(context.Background(), settings.ConnectionTimeout())
	defer cancel()

	conn, err := grpc.NewClient(
		settings.ServerAddress(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			dialer := &net.Dialer{}
			return dialer.DialContext(ctx, "tcp", addr)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(grpcutil.MaxMessageSizeBytes),
			grpc.MaxCallSendMsgSize(grpcutil.MaxMessageSizeBytes),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	if dialCtx.Err() != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("dialing gRPC server timed out: %w", dialCtx.Err())
	}

	grpcClient := grpcapi.NewNotificationServiceClient(conn)
	return &NotificationClient{
		conn:       conn,
		grpcClient: grpcClient,
		authToken:  settings.AuthToken(),
		logger:     logger,
		settings:   settings,
	}, nil
}

func (clientInstance *NotificationClient) Close() error {
	return clientInstance.conn.Close()
}

func (clientInstance *NotificationClient) SendNotification(ctx context.Context, req *grpcapi.NotificationRequest) (*grpcapi.NotificationResponse, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+clientInstance.authToken)
	resp, err := clientInstance.grpcClient.SendNotification(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (clientInstance *NotificationClient) GetNotificationStatus(notificationID string) (*grpcapi.NotificationResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), clientInstance.settings.OperationTimeout())
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

func (clientInstance *NotificationClient) SendNotificationAndWait(req *grpcapi.NotificationRequest) (*grpcapi.NotificationResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), clientInstance.settings.OperationTimeout())
	defer cancel()

	resp, err := clientInstance.SendNotification(ctx, req)
	if err != nil {
		clientInstance.logger.Error("SendNotification failed", "error", err)
		return nil, err
	}
	const pollInterval = 2 * time.Second
	pollTimeout := clientInstance.settings.OperationTimeout()
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
		statusResp, statusErr := clientInstance.GetNotificationStatus(resp.NotificationId)
		if statusErr != nil {
			clientInstance.logger.Error("GetNotificationStatus failed", "notificationID", resp.NotificationId, "error", statusErr)
			return nil, statusErr
		}
		resp = statusResp
	}
}
