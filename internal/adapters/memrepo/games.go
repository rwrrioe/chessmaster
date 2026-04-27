package memrepo

import (
	"context"
	"sync"
	"time"

	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/google/uuid"
)

// Games is an in-memory GameRepo.
type Games struct {
	mu           sync.RWMutex
	byID         map[uuid.UUID]ports.Game
	byInviteCode map[string]uuid.UUID
}

// NewGames returns an empty Games repo.
func NewGames() *Games {
	return &Games{
		byID:         make(map[uuid.UUID]ports.Game),
		byInviteCode: make(map[string]uuid.UUID),
	}
}

// Create inserts a new game, generating an ID.
func (r *Games) Create(_ context.Context, g ports.Game) (ports.Game, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	g.ID = uuid.New()
	g.CreatedAt = time.Now()
	r.byID[g.ID] = g
	if g.InviteCode != nil {
		r.byInviteCode[*g.InviteCode] = g.ID
	}
	return g, nil
}

// ByID retrieves a game by ID.
func (r *Games) ByID(_ context.Context, id uuid.UUID) (ports.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.byID[id]
	if !ok {
		return ports.Game{}, ErrNotFound
	}
	return g, nil
}

// ByInviteCode retrieves a game by invite code.
func (r *Games) ByInviteCode(_ context.Context, code string) (ports.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byInviteCode[code]
	if !ok {
		return ports.Game{}, ErrNotFound
	}
	return r.byID[id], nil
}

// UpdateStatus sets status, result, and pgn fields.
func (r *Games) UpdateStatus(_ context.Context, id uuid.UUID, status, result, pgn string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.byID[id]
	if !ok {
		return ErrNotFound
	}
	g.Status = status
	g.PGN = pgn
	if result != "" {
		g.Result = &result
	}
	now := time.Now()
	if g.StartedAt == nil && (status == "active" || status == "white_won" || status == "black_won" || status == "draw") {
		g.StartedAt = &now
	}
	if status == "white_won" || status == "black_won" || status == "draw" || status == "aborted" {
		g.EndedAt = &now
	}
	r.byID[id] = g
	return nil
}

// JoinAsBlack sets the black player for a pending game.
func (r *Games) JoinAsBlack(_ context.Context, id uuid.UUID, blackID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.byID[id]
	if !ok {
		return ErrNotFound
	}
	g.BlackID = &blackID
	g.Status = "active"
	now := time.Now()
	g.StartedAt = &now
	r.byID[id] = g
	return nil
}

// ListByPlayer returns up to limit games where the player is white or black.
func (r *Games) ListByPlayer(_ context.Context, playerID uuid.UUID, limit int) ([]ports.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []ports.Game
	for _, g := range r.byID {
		if (g.WhiteID != nil && *g.WhiteID == playerID) ||
			(g.BlackID != nil && *g.BlackID == playerID) {
			result = append(result, g)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}
