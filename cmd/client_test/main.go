package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/temirov/notify/pkg/client"
	"os"
	"time"

	"github.com/temirov/notify/pkg/grpcapi"
	"log/slog"
)

func main() {
	// Parse command-line flags.
	recipient := flag.String("to", "", "Recipient email address")
	subject := flag.String("subject", "", "Email subject")
	message := flag.String("message", "", "Email message body")
	flag.Parse()

	if *recipient == "" || *subject == "" || *message == "" {
		fmt.Fprintln(os.Stderr, "Usage: client_test --to <recipient> --subject <subject> --message <message>")
		os.Exit(1)
	}

	// Create a structured logger.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create the NotificationClient. All business logic (such as authentication, dialing,
	// and request handling) is encapsulated in the client package.
	notificationClient, err := client.NewNotificationClient(logger)
	if err != nil {
		logger.Error("Failed to create notification client", "error", err)
		os.Exit(1)
	}
	defer notificationClient.Close()

	// Build the gRPC NotificationRequest for an email.
	notificationRequest := &grpcapi.NotificationRequest{
		NotificationType: grpcapi.NotificationType_EMAIL,
		Recipient:        *recipient,
		Subject:          *subject,
		Message:          *message,
	}

	// Use a context with timeout.
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Send the notification using the client package.
	response, err := notificationClient.SendNotification(notificationRequest)
	if err != nil {
		logger.Error("Failed to send notification", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Notification sent successfully. Notification ID: %s\n", response.NotificationId)
}
