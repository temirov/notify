package config

import (
	"strings"
	"testing"
)

type envEntry struct {
	key   string
	value string
}

func TestLoadConfig(t *testing.T) {
	t.Helper()

	completeEnvironment := []envEntry{
		{key: "DATABASE_PATH", value: "test.db"},
		{key: "GRPC_AUTH_TOKEN", value: "unit-token"},
		{key: "LOG_LEVEL", value: "INFO"},
		{key: "MAX_RETRIES", value: "5"},
		{key: "RETRY_INTERVAL_SEC", value: "4"},
		{key: "SENDGRID_USERNAME", value: "apikey"},
		{key: "SENDGRID_PASSWORD", value: "secret"},
		{key: "SENDGRID_SMTP_SERVER", value: "smtp.test"},
		{key: "SENDGRID_SMTP_SERVER_PORT", value: "587"},
		{key: "FROM_EMAIL", value: "noreply@test"},
		{key: "TWILIO_ACCOUNT_SID", value: "sid"},
		{key: "TWILIO_AUTH_TOKEN", value: "auth"},
		{key: "TWILIO_FROM_NUMBER", value: "+10000000000"},
		{key: "CONNECTION_TIMEOUT_SEC", value: "3"},
		{key: "OPERATION_TIMEOUT_SEC", value: "7"},
	}

	testCases := []struct {
		name           string
		mutateEnv      func(t *testing.T)
		expectError    bool
		errorSubstring string
		expectedConfig Config
	}{
		{
			name: "AllVariablesPresent",
			mutateEnv: func(t *testing.T) {
				setEnvironment(t, completeEnvironment)
			},
			expectedConfig: Config{
				DatabasePath:         "test.db",
				GRPCAuthToken:        "unit-token",
				LogLevel:             "INFO",
				MaxRetries:           5,
				RetryIntervalSec:     4,
				SendGridUsername:     "apikey",
				SendGridPassword:     "secret",
				SendSmtpServer:       "smtp.test",
				SendSmtpServerPort:   587,
				FromEmail:            "noreply@test",
				TwilioAccountSID:     "sid",
				TwilioAuthToken:      "auth",
				TwilioFromNumber:     "+10000000000",
				ConnectionTimeoutSec: 3,
				OperationTimeoutSec:  7,
			},
		},
		{
			name: "MissingVariable",
			mutateEnv: func(t *testing.T) {
				truncated := completeEnvironment[:len(completeEnvironment)-1]
				setEnvironment(t, truncated)
			},
			expectError:    true,
			errorSubstring: "missing environment variable OPERATION_TIMEOUT_SEC",
		},
		{
			name: "InvalidInteger",
			mutateEnv: func(t *testing.T) {
				invalid := append([]envEntry{}, completeEnvironment...)
				invalid[3].value = "invalid"
				setEnvironment(t, invalid)
			},
			expectError:    true,
			errorSubstring: "invalid integer for MAX_RETRIES",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Helper()
			testCase.mutateEnv(t)

			loadedConfig, loadError := LoadConfig()
			if testCase.expectError {
				if loadError == nil {
					t.Fatalf("expected error")
				}
				if !strings.Contains(loadError.Error(), testCase.errorSubstring) {
					t.Fatalf("unexpected error %v", loadError)
				}
				return
			}

			if loadError != nil {
				t.Fatalf("load config error: %v", loadError)
			}

			if loadedConfig != testCase.expectedConfig {
				t.Fatalf("unexpected config %+v", loadedConfig)
			}
		})
	}
}

func setEnvironment(t *testing.T, entries []envEntry) {
	t.Helper()
	for _, entry := range entries {
		t.Setenv(entry.key, entry.value)
	}
}
