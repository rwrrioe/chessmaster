package memrepo

import (
	"context"
	"math"
	"sort"
	"sync"

	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/google/uuid"
)

// Ratings is an in-memory RatingRepo.
type Ratings struct {
	mu      sync.Mutex
	ratings map[uuid.UUID]*ports.Rating
	// players is a reference to look up username/city for leaderboard.
	players *Players
}

// NewRatings returns an empty Ratings repo.
func NewRatings(players *Players) *Ratings {
	return &Ratings{
		ratings: make(map[uuid.UUID]*ports.Rating),
		players: players,
	}
}

func (r *Ratings) ensure(id uuid.UUID) *ports.Rating {
	if rt, ok := r.ratings[id]; ok {
		return rt
	}
	rt := &ports.Rating{PlayerID: id, Elo: 1200}
	r.ratings[id] = rt
	return rt
}

// Get returns the rating for a player, creating it with Elo 1200 if missing.
func (r *Ratings) Get(_ context.Context, playerID uuid.UUID) (ports.Rating, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rt := r.ensure(playerID)
	return *rt, nil
}

// ApplyResult updates Elo for both players (K=32). Pass nil for AI side.
func (r *Ratings) ApplyResult(_ context.Context, whiteID, blackID *uuid.UUID, result string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if whiteID == nil || blackID == nil {
		// AI game — no rating change
		return nil
	}

	white := r.ensure(*whiteID)
	black := r.ensure(*blackID)

	eW := eloExpected(white.Elo, black.Elo)
	eB := eloExpected(black.Elo, white.Elo)

	var sW, sB float64
	switch result {
	case "1-0":
		sW, sB = 1, 0
		white.Wins++
		black.Losses++
	case "0-1":
		sW, sB = 0, 1
		white.Losses++
		black.Wins++
	default: // draw
		sW, sB = 0.5, 0.5
		white.Draws++
		black.Draws++
	}

	white.Elo = newElo(white.Elo, eW, sW)
	black.Elo = newElo(black.Elo, eB, sB)
	white.Games++
	black.Games++

	return nil
}

// Leaderboard returns top players ordered by Elo, optionally filtered by city.
func (r *Ratings) Leaderboard(_ context.Context, city *string, limit int) ([]ports.LeaderboardEntry, error) {
	r.mu.Lock()
	ratingsCopy := make([]ports.Rating, 0, len(r.ratings))
	for _, rt := range r.ratings {
		ratingsCopy = append(ratingsCopy, *rt)
	}
	r.mu.Unlock()

	sort.Slice(ratingsCopy, func(i, j int) bool {
		return ratingsCopy[i].Elo > ratingsCopy[j].Elo
	})

	var entries []ports.LeaderboardEntry
	for _, rt := range ratingsCopy {
		p, err := r.players.ByID(context.Background(), rt.PlayerID)
		if err != nil {
			continue
		}
		if city != nil && p.City != *city {
			continue
		}
		entries = append(entries, ports.LeaderboardEntry{
			Username: p.Username,
			City:     p.City,
			Elo:      rt.Elo,
			Games:    rt.Games,
			Wins:     rt.Wins,
		})
		if len(entries) >= limit {
			break
		}
	}
	return entries, nil
}

func eloExpected(a, b int) float64 {
	return 1.0 / (1.0 + math.Pow(10, float64(b-a)/400.0))
}

func newElo(current int, expected, score float64) int {
	return int(math.Round(float64(current) + 32*(score-expected)))
}
