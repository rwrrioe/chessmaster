package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations.sql
var initSQL string

// EnsureSchema runs the embedded initial schema if the players table is missing.
// Idempotent: a second call on a populated database is a cheap no-op.
func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	var exists bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = current_schema() AND table_name = 'players'
		)
	`).Scan(&exists); err != nil {
		return fmt.Errorf("schema check: %w", err)
	}
	if exists {
		return nil
	}
	if _, err := pool.Exec(ctx, initSQL); err != nil {
		return fmt.Errorf("schema apply: %w", err)
	}
	return nil
}
