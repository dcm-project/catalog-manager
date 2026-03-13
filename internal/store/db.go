package store

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dcm-project/catalog-manager/internal/config"
	"github.com/dcm-project/catalog-manager/internal/store/model"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the database connection and performs auto-migration
func InitDB(cfg *config.Config, slogger *slog.Logger) (*gorm.DB, error) {
	var dialector gorm.Dialector

	// Select database dialect based on configuration
	if cfg.Database.Type == "pgsql" {
		dsn := fmt.Sprintf("host=%s user=%s password=%s port=%s dbname=%s",
			cfg.Database.Hostname,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Port,
			cfg.Database.Name,
		)
		dialector = postgres.Open(dsn)
	} else {
		dialector = sqlite.Open(cfg.Database.Name)
	}

	// Configure GORM logger to respect the application's configured log level.
	// Map slog levels to GORM log levels so DB logging follows LOG_LEVEL.
	var gormLogLevel logger.LogLevel
	var slogBridgeLevel slog.Level
	switch {
	case slogger.Handler().Enabled(context.Background(), slog.LevelDebug):
		gormLogLevel = logger.Info // GORM Info = log all queries
		slogBridgeLevel = slog.LevelDebug
	case slogger.Handler().Enabled(context.Background(), slog.LevelInfo):
		gormLogLevel = logger.Warn // slow queries + errors
		slogBridgeLevel = slog.LevelWarn
	default:
		gormLogLevel = logger.Error // errors only
		slogBridgeLevel = slog.LevelError
	}

	gormLogger := logger.New(
		slog.NewLogLogger(slogger.Handler(), slogBridgeLevel),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Open database connection
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable foreign key constraints for SQLite
	if cfg.Database.Type != "pgsql" {
		if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
			return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
		}
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	slogger.Info("Database connection established", "type", cfg.Database.Type)

	// Auto-migrate all models
	if err := db.AutoMigrate(
		&model.ServiceType{},
		&model.CatalogItem{},
		&model.CatalogItemInstance{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database schema: %w", err)
	}

	slogger.Info("Database schema migrated")

	return db, nil
}
