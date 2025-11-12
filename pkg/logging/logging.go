package logging

import (
	"log/slog"
	"os"
	"strings"
)

// NewLogger creates a slog.Logger configured according to the provided log
// level string (DEBUG/INFO/WARN/ERROR), defaulting to INFO.
func NewLogger(levelString string) *slog.Logger {
	var level slog.Level
	switch strings.ToUpper(levelString) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}
