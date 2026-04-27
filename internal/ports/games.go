package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Game represents a chess game row.
type Game struct {
	ID         uuid.UUID
	WhiteID    *uuid.UUID
	BlackID    *uuid.UUID
	Mode       string
	Status     string
	InviteCode *string
	PGN        string
	Result     *string
	StartedAt  *time.Time
	EndedAt    *time.Time
	CreatedAt  time.Time
}

// GameRepo defines persistence operations for games.
type GameRepo interface {
	Create(ctx context.Context, g Game) (Game, error)
	ByID(ctx context.Context, id uuid.UUID) (Game, error)
	ByInviteCode(ctx context.Context, code string) (Game, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status, result, pgn string) error
	JoinAsBlack(ctx context.Context, id uuid.UUID, blackID uuid.UUID) error
	ListByPlayer(ctx context.Context, playerID uuid.UUID, limit int) ([]Game, error)
}
