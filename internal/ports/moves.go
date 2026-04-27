package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Move represents a single chess move persisted in the database.
type Move struct {
	ID       int64
	GameID   uuid.UUID
	Ply      int
	UCI      string
	SAN      string
	FENAfter string
	PlayedAt time.Time
}

// MoveRepo defines persistence operations for moves.
type MoveRepo interface {
	Append(ctx context.Context, gameID uuid.UUID, ply int, uci, san, fenAfter string) error
	ListByGame(ctx context.Context, gameID uuid.UUID) ([]Move, error)
}
