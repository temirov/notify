package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds runtime configuration
type Config struct {
	ServerPort       int
	DatabasePath     string
	AuthToken        string
	LogLevel         string
	MaxRetries       int
	RetryIntervalSec int

	// SendGrid-specific
	SendGridUsername string // typically "apikey"
	SendGridPassword string // your SendGrid API key
	FromEmail        string // "support@rsvp.mprlab.com"
}

// LoadConfig reads from environment variables
func LoadConfig() Config {
	return Config{
		ServerPort:       getInt("SERVER_PORT", 8080),
		DatabasePath:     getStr("DATABASE_PATH", "app.db"),
		AuthToken:        getStr("NOTIFICATION_AUTH_TOKEN", ""), // optional
		LogLevel:         getStr("LOG_LEVEL", "INFO"),           // "DEBUG","INFO","WARN","ERROR"
		MaxRetries:       getInt("MAX_RETRIES", 3),
		RetryIntervalSec: getInt("RETRY_INTERVAL_SEC", 15),

		SendGridUsername: getStr("SENDGRID_USERNAME", "apikey"),
		SendGridPassword: getStr("SENDGRID_PASSWORD", ""), // must set your real API key
		FromEmail:        getStr("FROM_EMAIL", "support@rsvp.mprlab.com"),
	}
}

func (c Config) ServerAddress() string {
	return fmt.Sprintf(":%d", c.ServerPort)
}

func getStr(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	valInt, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return valInt
}
