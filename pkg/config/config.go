package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Config struct {
	DatabasePath     string
	GRPCAuthToken    string
	LogLevel         string
	MaxRetries       int
	RetryIntervalSec int

	SendGridUsername   string
	SendGridPassword   string
	SendSmtpServer     string
	SendSmtpServerPort int
	FromEmail          string

	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string
}

// LoadConfig retrieves all required environment variables concurrently.
// It returns a fully populated Config instance or an aggregated error if any variable is missing or invalid.
func LoadConfig() (Config, error) {
	var configuration Config
	var waitGroup sync.WaitGroup

	// Define tasks using helper functions.
	taskFunctions := []func() error{
		loadEnvString("DATABASE_PATH", &configuration.DatabasePath),
		loadEnvString("GRPC_AUTH_TOKEN", &configuration.GRPCAuthToken),
		loadEnvString("LOG_LEVEL", &configuration.LogLevel),
		loadEnvInt("MAX_RETRIES", &configuration.MaxRetries),
		loadEnvInt("RETRY_INTERVAL_SEC", &configuration.RetryIntervalSec),
		loadEnvString("SENDGRID_USERNAME", &configuration.SendGridUsername),
		loadEnvString("SENDGRID_PASSWORD", &configuration.SendGridPassword),
		loadEnvString("SENDGRID_SMTP_SERVER", &configuration.SendSmtpServer),
		loadEnvInt("SENDGRID_SMTP_SERVER_PORT", &configuration.SendSmtpServerPort),
		loadEnvString("FROM_EMAIL", &configuration.FromEmail),
		loadEnvString("TWILIO_ACCOUNT_SID", &configuration.TwilioAccountSID),
		loadEnvString("TWILIO_AUTH_TOKEN", &configuration.TwilioAuthToken),
		loadEnvString("TWILIO_FROM_NUMBER", &configuration.TwilioFromNumber),
	}

	// Buffered channel to collect errors from tasks.
	errorChannel := make(chan error, len(taskFunctions))

	// Launch each task concurrently.
	for _, taskFunction := range taskFunctions {
		waitGroup.Add(1)
		go func(task func() error) {
			defer waitGroup.Done()
			if taskError := task(); taskError != nil {
				errorChannel <- taskError
			}
		}(taskFunction)
	}

	// Wait for all tasks to finish.
	waitGroup.Wait()
	close(errorChannel)

	var errorMessages []string
	for errorValue := range errorChannel {
		errorMessages = append(errorMessages, errorValue.Error())
	}
	if len(errorMessages) > 0 {
		// Using a constant format string fixes the vet error.
		return Config{}, fmt.Errorf("configuration errors: %s", strings.Join(errorMessages, ", "))
	}
	return configuration, nil
}

// loadEnvString returns a function that retrieves a non-empty string environment variable.
func loadEnvString(environmentKey string, destination *string) func() error {
	const missingEnvFormat = "missing environment variable %s"
	return func() error {
		environmentValue := os.Getenv(environmentKey)
		if environmentValue == "" {
			return fmt.Errorf(missingEnvFormat, environmentKey)
		}
		*destination = environmentValue
		return nil
	}
}

// loadEnvInt returns a function that retrieves an integer environment variable.
func loadEnvInt(environmentKey string, destination *int) func() error {
	const missingEnvFormat = "missing environment variable %s"
	const invalidIntFormat = "invalid integer for %s: %v"
	return func() error {
		environmentValue := os.Getenv(environmentKey)
		if environmentValue == "" {
			return fmt.Errorf(missingEnvFormat, environmentKey)
		}
		parsedInteger, conversionError := strconv.Atoi(environmentValue)
		if conversionError != nil {
			return fmt.Errorf(invalidIntFormat, environmentKey, conversionError)
		}
		*destination = parsedInteger
		return nil
	}
}
