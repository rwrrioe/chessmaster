package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Player represents a registered chess player.
type Player struct {
	ID           uuid.UUID
	Email        string
	Username     string
	PasswordHash string
	City         string
	IsPro        bool
	CreatedAt    time.Time
}

// PlayerRepo defines persistence operations for players.
type PlayerRepo interface {
	Create(ctx context.Context, p Player) (Player, error)
	ByEmail(ctx context.Context, email string) (Player, error)
	ByID(ctx context.Context, id uuid.UUID) (Player, error)
	UpdateCity(ctx context.Context, id uuid.UUID, city string) error
	SetPro(ctx context.Context, id uuid.UUID, isPro bool) error
}
