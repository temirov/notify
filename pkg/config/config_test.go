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
		{key: "SMTP_USERNAME", value: "apikey"},
		{key: "SMTP_PASSWORD", value: "secret"},
		{key: "SMTP_HOST", value: "smtp.test"},
		{key: "SMTP_PORT", value: "587"},
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
		assert         func(t *testing.T, cfg Config)
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
				SMTPUsername:         "apikey",
				SMTPPassword:         "secret",
				SMTPHost:             "smtp.test",
				SMTPPort:             587,
				FromEmail:            "noreply@test",
				TwilioAccountSID:     "sid",
				TwilioAuthToken:      "auth",
				TwilioFromNumber:     "+10000000000",
				ConnectionTimeoutSec: 3,
				OperationTimeoutSec:  7,
			},
			assert: func(t *testing.T, cfg Config) {
				t.Helper()
				if !cfg.TwilioConfigured() {
					t.Fatalf("expected Twilio to be configured")
				}
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
		{
			name: "TwilioCredentialsOptional",
			mutateEnv: func(t *testing.T) {
				var trimmed []envEntry
				for _, entry := range completeEnvironment {
					if strings.HasPrefix(entry.key, "TWILIO_") {
						continue
					}
					trimmed = append(trimmed, entry)
				}
				setEnvironment(t, trimmed)
			},
			expectedConfig: Config{
				DatabasePath:         "test.db",
				GRPCAuthToken:        "unit-token",
				LogLevel:             "INFO",
				MaxRetries:           5,
				RetryIntervalSec:     4,
				SMTPUsername:         "apikey",
				SMTPPassword:         "secret",
				SMTPHost:             "smtp.test",
				SMTPPort:             587,
				FromEmail:            "noreply@test",
				ConnectionTimeoutSec: 3,
				OperationTimeoutSec:  7,
			},
			assert: func(t *testing.T, cfg Config) {
				t.Helper()
				if cfg.TwilioConfigured() {
					t.Fatalf("expected Twilio to be disabled")
				}
			},
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

			if testCase.assert != nil {
				testCase.assert(t, loadedConfig)
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
