package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/temirov/pinguin/pkg/grpcapi"
	"github.com/temirov/pinguin/pkg/secret"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NotificationSender interface {
	SendNotification(context.Context, *grpcapi.NotificationRequest) (*grpcapi.NotificationResponse, error)
}

type SecretGenerator interface {
	GenerateSecret(context.Context, secret.ByteLength) (string, error)
}

type Dependencies struct {
	Sender           NotificationSender
	OperationTimeout time.Duration
	Output           io.Writer
	SecretGenerator  SecretGenerator
}

func NewRootCommand(dependencies Dependencies) *cobra.Command {
	root := &cobra.Command{
		Use:           "pinguin-cli",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(buildSendCommand(dependencies))
	root.AddCommand(buildGenerateSecretCommand(dependencies))
	return root
}

func buildSendCommand(dependencies Dependencies) *cobra.Command {
	var (
		typeInput      string
		recipientInput string
		subjectInput   string
		messageInput   string
		scheduledInput string
	)

	command := &cobra.Command{
		Use:   "send",
		Short: "Submit a notification to the Pinguin service",
		RunE: func(cmd *cobra.Command, args []string) error {
			notificationType, err := parseNotificationType(typeInput)
			if err != nil {
				return err
			}

			request := &grpcapi.NotificationRequest{
				NotificationType: notificationType,
				Recipient:        recipientInput,
				Subject:          subjectInput,
				Message:          messageInput,
			}

			if scheduledInput != "" {
				scheduledTime, parseErr := time.Parse(time.RFC3339, scheduledInput)
				if parseErr != nil {
					return fmt.Errorf("invalid scheduled time %q: %w", scheduledInput, parseErr)
				}
				request.ScheduledTime = timestamppb.New(scheduledTime.UTC())
			}

			timeout := dependencies.OperationTimeout
			if timeout <= 0 {
				timeout = 30 * time.Second
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			response, sendErr := dependencies.Sender.SendNotification(ctx, request)
			if sendErr != nil {
				return sendErr
			}

			output := dependencies.Output
			if output == nil {
				output = io.Discard
			}

			_, writeErr := fmt.Fprintf(
				output,
				"Notification %s sent with status %s\n",
				response.NotificationId,
				response.Status.String(),
			)
			if writeErr != nil {
				return writeErr
			}

			return nil
		},
	}

	command.Flags().StringVar(&typeInput, "type", "", "Notification type (email or sms)")
	command.Flags().StringVar(&recipientInput, "recipient", "", "Notification recipient")
	command.Flags().StringVar(&subjectInput, "subject", "", "Email subject (ignored for sms)")
	command.Flags().StringVar(&messageInput, "message", "", "Notification message")
	command.Flags().StringVar(&scheduledInput, "scheduled-time", "", "RFC3339 timestamp for scheduled delivery")

	markRequired(command, "type")
	markRequired(command, "recipient")
	markRequired(command, "message")

	return command
}

func parseNotificationType(input string) (grpcapi.NotificationType, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "email":
		return grpcapi.NotificationType_EMAIL, nil
	case "sms":
		return grpcapi.NotificationType_SMS, nil
	default:
		return grpcapi.NotificationType_EMAIL, fmt.Errorf("invalid notification type %q", input)
	}
}

func markRequired(cmd *cobra.Command, name string) {
	_ = cmd.MarkFlagRequired(name)
}

func buildGenerateSecretCommand(dependencies Dependencies) *cobra.Command {
	var bytesLength int

	command := &cobra.Command{
		Use:   "generate-secret",
		Short: "Generate a NOTIFICATION_AUTH_TOKEN value",
		RunE: func(cmd *cobra.Command, args []string) error {
			generator := dependencies.SecretGenerator
			if generator == nil {
				return errors.New("secret generator is not configured")
			}

			length := secret.DefaultByteLength()
			if bytesLength > 0 {
				parsedLength, err := secret.NewByteLength(bytesLength)
				if err != nil {
					return fmt.Errorf("invalid secret length: %w", err)
				}
				length = parsedLength
			}

			secretValue, err := generator.GenerateSecret(cmd.Context(), length)
			if err != nil {
				return err
			}

			output := dependencies.Output
			if output == nil {
				output = io.Discard
			}
			_, writeErr := fmt.Fprintln(output, secretValue)
			if writeErr != nil {
				return writeErr
			}

			return nil
		},
	}

	command.Flags().IntVar(&bytesLength, "bytes", 0, "Number of random bytes for the secret (minimum 32)")

	return command
}
