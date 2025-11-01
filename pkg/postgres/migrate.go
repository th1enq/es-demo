package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

//go:embed migrations/001_event_store.sql
var eventStoreMigration string

// RunMigrations executes SQL migration files for event store and demo accounts
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) error {
	logger.Info("Starting database migrations...")

	// Execute event store migration FIRST
	_, err := pool.Exec(ctx, eventStoreMigration)
	if err != nil {
		logger.Error("Failed to execute event store migration", zap.Error(err))
		return fmt.Errorf("failed to execute event store migration: %w", err)
	}
	logger.Info("Event store migration completed")

	logger.Info("Database migrations completed successfully")
	return nil
}
