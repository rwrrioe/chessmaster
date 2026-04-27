// Package memrepo provides in-memory implementations of all port interfaces for testing and dev.
package memrepo

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/google/uuid"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict is returned when a uniqueness constraint is violated.
var ErrConflict = errors.New("conflict")

// Players is an in-memory PlayerRepo.
type Players struct {
	mu      sync.RWMutex
	byID    map[uuid.UUID]ports.Player
	byEmail map[string]uuid.UUID
}

// NewPlayers returns an empty Players repo.
func NewPlayers() *Players {
	return &Players{
		byID:    make(map[uuid.UUID]ports.Player),
		byEmail: make(map[string]uuid.UUID),
	}
}

// Create inserts a new player, generating an ID and timestamps.
func (r *Players) Create(_ context.Context, p ports.Player) (ports.Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byEmail[p.Email]; ok {
		return ports.Player{}, ErrConflict
	}
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	r.byID[p.ID] = p
	r.byEmail[p.Email] = p.ID
	return p, nil
}

// ByEmail looks up a player by email.
func (r *Players) ByEmail(_ context.Context, email string) (ports.Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[email]
	if !ok {
		return ports.Player{}, ErrNotFound
	}
	return r.byID[id], nil
}

// ByID looks up a player by UUID.
func (r *Players) ByID(_ context.Context, id uuid.UUID) (ports.Player, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byID[id]
	if !ok {
		return ports.Player{}, ErrNotFound
	}
	return p, nil
}

// UpdateCity sets the city for a player.
func (r *Players) UpdateCity(_ context.Context, id uuid.UUID, city string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.byID[id]
	if !ok {
		return ErrNotFound
	}
	p.City = city
	r.byID[id] = p
	return nil
}

// SetPro sets the is_pro flag for a player.
func (r *Players) SetPro(_ context.Context, id uuid.UUID, isPro bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.byID[id]
	if !ok {
		return ErrNotFound
	}
	p.IsPro = isPro
	r.byID[id] = p
	return nil
}
