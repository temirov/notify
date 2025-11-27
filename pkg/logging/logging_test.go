package logging

import (
	"context"
	"log/slog"
	"testing"
)

func TestNewLoggerLevels(t *testing.T) {
	t.Helper()
	testCases := []struct {
		level       string
		expectDebug bool
		expectWarn  bool
		expectError bool
	}{
		{"debug", true, true, true},
		{"info", false, true, true},
		{"warn", false, true, true},
		{"error", false, false, true},
		{"", false, true, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.level, func(t *testing.T) {
			t.Helper()
			logger := NewLogger(tc.level)
			if logger == nil {
				t.Fatalf("expected logger instance")
			}
			ctx := context.Background()
			if got := logger.Enabled(ctx, slog.LevelDebug); got != tc.expectDebug {
				t.Fatalf("debug enabled mismatch: got %v want %v", got, tc.expectDebug)
			}
			if got := logger.Enabled(ctx, slog.LevelWarn); got != tc.expectWarn {
				t.Fatalf("warn enabled mismatch: got %v want %v", got, tc.expectWarn)
			}
			if got := logger.Enabled(ctx, slog.LevelError); got != tc.expectError {
				t.Fatalf("error enabled mismatch: got %v want %v", got, tc.expectError)
			}
		})
	}
}
