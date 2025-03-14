package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/temirov/notify/pkg/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB opens (or creates) the SQLite file and auto-migrates schema
func InitDB(dbPath string, logger *slog.Logger) (*gorm.DB, error) {
	logger.Info("Initializing SQLite DB", "path", dbPath)

	gormLogger := &slogGormLogger{logger: logger}
	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite failed: %w", err)
	}

	if err := database.AutoMigrate(&model.Notification{}); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return database, nil
}

// Implement GORM’s logger.Interface for v1.25+ with context.Context

type slogGormLogger struct {
	logger *slog.Logger
}

// Ensure we implement the interface
var _ logger.Interface = (*slogGormLogger)(nil)

func (l *slogGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	// You can adjust log level dynamically if you wish. We'll just return the same logger
	return l
}

func (l *slogGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Info(msg, data...)
}

func (l *slogGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Warn(msg, data...)
}

func (l *slogGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Error(msg, data...)
}

func (l *slogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	// If you want to log details about every query (SQL, rows, elapsed time), implement here.
	sql, rows := fc()
	elapsed := time.Since(begin)

	if err != nil && err != gorm.ErrRecordNotFound {
		l.logger.Error("Trace",
			"error", err,
			"sql", sql,
			"rows", rows,
			"elapsed", elapsed,
		)
	} else {
		// Could do debug-level logging or skip entirely
	}
}
