package memrepo

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/google/uuid"
)

// Moves is an in-memory MoveRepo.
type Moves struct {
	mu    sync.RWMutex
	moves []ports.Move
	nextID int64
}

// NewMoves returns an empty Moves repo.
func NewMoves() *Moves {
	return &Moves{nextID: 1}
}

// Append records a new move.
func (r *Moves) Append(_ context.Context, gameID uuid.UUID, ply int, uci, san, fenAfter string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.moves = append(r.moves, ports.Move{
		ID:       r.nextID,
		GameID:   gameID,
		Ply:      ply,
		UCI:      uci,
		SAN:      san,
		FENAfter: fenAfter,
		PlayedAt: time.Now(),
	})
	r.nextID++
	return nil
}

// ListByGame returns all moves for a game ordered by ply.
func (r *Moves) ListByGame(_ context.Context, gameID uuid.UUID) ([]ports.Move, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []ports.Move
	for _, m := range r.moves {
		if m.GameID == gameID {
			result = append(result, m)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Ply < result[j].Ply })
	return result, nil
}
