package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/temirov/pinguin/pkg/client"
	"github.com/temirov/pinguin/pkg/config"
	"github.com/temirov/pinguin/pkg/grpcapi"
	"log/slog"
)

func main() {
	recipient := flag.String("to", "", "Recipient email address")
	subject := flag.String("subject", "", "Email subject")
	message := flag.String("message", "", "Email message body")
	flag.Parse()

	if *recipient == "" || *subject == "" || *message == "" {
		fmt.Fprintln(os.Stderr, "Usage: client_test --to <recipient> --subject <subject> --message <message>")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	notificationClient, err := client.NewNotificationClient(logger, cfg)
	if err != nil {
		logger.Error("Failed to create notification client", "error", err)
		os.Exit(1)
	}
	defer notificationClient.Close()

	notificationRequest := &grpcapi.NotificationRequest{
		NotificationType: grpcapi.NotificationType_EMAIL,
		Recipient:        *recipient,
		Subject:          *subject,
		Message:          *message,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.OperationTimeoutSec)*time.Second)
	defer cancel()

	response, err := notificationClient.SendNotification(ctx, notificationRequest)
	if err != nil {
		logger.Error("Failed to send notification", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Notification sent successfully. Notification ID: %s\n", response.NotificationId)
}
