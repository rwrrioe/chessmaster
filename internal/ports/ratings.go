package ports

import (
	"context"

	"github.com/google/uuid"
)

// Rating holds the Elo data for a player.
type Rating struct {
	PlayerID uuid.UUID
	Elo      int
	Games    int
	Wins     int
	Losses   int
	Draws    int
}

// LeaderboardEntry is a row returned by the leaderboard query.
type LeaderboardEntry struct {
	Username string
	City     string
	Elo      int
	Games    int
	Wins     int
}

// RatingRepo defines persistence operations for Elo ratings.
type RatingRepo interface {
	// Get returns the rating for a player, auto-creating it with Elo 1200 if missing.
	Get(ctx context.Context, playerID uuid.UUID) (Rating, error)
	// ApplyResult updates Elo for both players using K=32. Pass nil for AI side.
	ApplyResult(ctx context.Context, whiteID, blackID *uuid.UUID, result string) error
	// Leaderboard returns the top players, optionally filtered by city.
	Leaderboard(ctx context.Context, city *string, limit int) ([]LeaderboardEntry, error)
}
