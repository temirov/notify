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

	HTTPListenAddr     string
	HTTPStaticRoot     string
	HTTPAllowedOrigins []string

	TAuthSigningKey string
	TAuthIssuer     string
	TAuthCookieName string

	SMTPUsername string
	SMTPPassword string
	SMTPHost     string
	SMTPPort     int
	FromEmail    string

	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string

	// Simplified timeout settings (in seconds)
	ConnectionTimeoutSec int
	OperationTimeoutSec  int
}

// LoadConfig retrieves all required environment variables concurrently.
func LoadConfig() (Config, error) {
	var configuration Config
	var waitGroup sync.WaitGroup

	taskFunctions := []func() error{
		loadEnvString("DATABASE_PATH", &configuration.DatabasePath),
		loadEnvString("GRPC_AUTH_TOKEN", &configuration.GRPCAuthToken),
		loadEnvString("LOG_LEVEL", &configuration.LogLevel),
		loadEnvInt("MAX_RETRIES", &configuration.MaxRetries),
		loadEnvInt("RETRY_INTERVAL_SEC", &configuration.RetryIntervalSec),
		loadEnvString("HTTP_LISTEN_ADDR", &configuration.HTTPListenAddr),
		loadEnvString("HTTP_STATIC_ROOT", &configuration.HTTPStaticRoot),
		loadEnvString("TAUTH_SIGNING_KEY", &configuration.TAuthSigningKey),
		loadEnvString("TAUTH_ISSUER", &configuration.TAuthIssuer),
		loadEnvString("SMTP_USERNAME", &configuration.SMTPUsername),
		loadEnvString("SMTP_PASSWORD", &configuration.SMTPPassword),
		loadEnvString("SMTP_HOST", &configuration.SMTPHost),
		loadEnvInt("SMTP_PORT", &configuration.SMTPPort),
		loadEnvString("FROM_EMAIL", &configuration.FromEmail),
		loadEnvInt("CONNECTION_TIMEOUT_SEC", &configuration.ConnectionTimeoutSec),
		loadEnvInt("OPERATION_TIMEOUT_SEC", &configuration.OperationTimeoutSec),
	}

	errorChannel := make(chan error, len(taskFunctions))
	for _, taskFunction := range taskFunctions {
		waitGroup.Add(1)
		go func(task func() error) {
			defer waitGroup.Done()
			if taskError := task(); taskError != nil {
				errorChannel <- taskError
			}
		}(taskFunction)
	}

	waitGroup.Wait()
	close(errorChannel)

	var errorMessages []string
	for errorValue := range errorChannel {
		errorMessages = append(errorMessages, errorValue.Error())
	}
	if len(errorMessages) > 0 {
		return Config{}, fmt.Errorf("configuration errors: %s", strings.Join(errorMessages, ", "))
	}

	configuration.TwilioAccountSID = strings.TrimSpace(os.Getenv("TWILIO_ACCOUNT_SID"))
	configuration.TwilioAuthToken = strings.TrimSpace(os.Getenv("TWILIO_AUTH_TOKEN"))
	configuration.TwilioFromNumber = strings.TrimSpace(os.Getenv("TWILIO_FROM_NUMBER"))
	configuration.TAuthCookieName = strings.TrimSpace(os.Getenv("TAUTH_COOKIE_NAME"))
	if configuration.TAuthCookieName == "" {
		configuration.TAuthCookieName = "app_session"
	}
	configuration.HTTPAllowedOrigins = parseCSV(os.Getenv("HTTP_ALLOWED_ORIGINS"))

	return configuration, nil
}

func loadEnvString(environmentKey string, destination *string) func() error {
	const missingEnvFormat = "missing environment variable %s"
	return func() error {
		environmentValue := strings.TrimSpace(os.Getenv(environmentKey))
		if environmentValue == "" {
			return fmt.Errorf(missingEnvFormat, environmentKey)
		}
		*destination = environmentValue
		return nil
	}
}

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

func (configuration Config) TwilioConfigured() bool {
	return configuration.TwilioAccountSID != "" && configuration.TwilioAuthToken != "" && configuration.TwilioFromNumber != ""
}

func parseCSV(value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	rawParts := strings.Split(trimmed, ",")
	var normalized []string
	for _, part := range rawParts {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}
		normalized = append(normalized, candidate)
	}
	return normalized
}
