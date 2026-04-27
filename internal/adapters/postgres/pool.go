// Package postgres provides PostgreSQL-backed implementations of the port interfaces.
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// New creates and validates a pgxpool connection pool using the provided connection URL.
func New(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
