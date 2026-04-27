package postgres

import (
	"context"
	"math"

	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ratings implements ports.RatingRepo against PostgreSQL.
type Ratings struct {
	pool *pgxpool.Pool
}

// NewRatings creates a Ratings repo backed by the given pool.
func NewRatings(pool *pgxpool.Pool) *Ratings { return &Ratings{pool: pool} }

// Get returns the rating for a player, inserting defaults if missing.
func (r *Ratings) Get(ctx context.Context, playerID uuid.UUID) (ports.Rating, error) {
	const upsert = `
		INSERT INTO ratings (player_id) VALUES ($1)
		ON CONFLICT (player_id) DO NOTHING`
	_, _ = r.pool.Exec(ctx, upsert, playerID)

	const q = `SELECT player_id, elo, games, wins, losses, draws FROM ratings WHERE player_id=$1`
	row := r.pool.QueryRow(ctx, q, playerID)
	var rt ports.Rating
	if err := row.Scan(&rt.PlayerID, &rt.Elo, &rt.Games, &rt.Wins, &rt.Losses, &rt.Draws); err != nil {
		return ports.Rating{}, err
	}
	return rt, nil
}

// ApplyResult updates Elo for both sides using K=32. Pass nil for AI side.
func (r *Ratings) ApplyResult(ctx context.Context, whiteID, blackID *uuid.UUID, result string) error {
	if whiteID == nil || blackID == nil {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint

	const lockQ = `
		INSERT INTO ratings (player_id) VALUES ($1), ($2)
		ON CONFLICT (player_id) DO NOTHING`
	if _, err = tx.Exec(ctx, lockQ, *whiteID, *blackID); err != nil {
		return err
	}

	const selectQ = `SELECT player_id, elo, games, wins, losses, draws FROM ratings WHERE player_id=ANY($1) FOR UPDATE`
	rows, err := tx.Query(ctx, selectQ, []uuid.UUID{*whiteID, *blackID})
	if err != nil {
		return err
	}

	byID := map[uuid.UUID]*ports.Rating{}
	for rows.Next() {
		var rt ports.Rating
		if err = rows.Scan(&rt.PlayerID, &rt.Elo, &rt.Games, &rt.Wins, &rt.Losses, &rt.Draws); err != nil {
			rows.Close()
			return err
		}
		rt2 := rt
		byID[rt.PlayerID] = &rt2
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return err
	}

	white := byID[*whiteID]
	black := byID[*blackID]

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
	default:
		sW, sB = 0.5, 0.5
		white.Draws++
		black.Draws++
	}

	white.Elo = newElo(white.Elo, eW, sW)
	black.Elo = newElo(black.Elo, eB, sB)
	white.Games++
	black.Games++

	const updQ = `UPDATE ratings SET elo=$2,games=$3,wins=$4,losses=$5,draws=$6,updated_at=NOW() WHERE player_id=$1`
	if _, err = tx.Exec(ctx, updQ, white.PlayerID, white.Elo, white.Games, white.Wins, white.Losses, white.Draws); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, updQ, black.PlayerID, black.Elo, black.Games, black.Wins, black.Losses, black.Draws); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Leaderboard returns top players ordered by Elo, optionally filtered by city.
func (r *Ratings) Leaderboard(ctx context.Context, city *string, limit int) ([]ports.LeaderboardEntry, error) {
	const q = `
		SELECT p.username, COALESCE(p.city,''), rt.elo, rt.games, rt.wins
		FROM ratings rt JOIN players p ON p.id=rt.player_id
		WHERE ($1::text IS NULL OR p.city=$1)
		ORDER BY rt.elo DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, q, city, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []ports.LeaderboardEntry
	for rows.Next() {
		var e ports.LeaderboardEntry
		if err = rows.Scan(&e.Username, &e.City, &e.Elo, &e.Games, &e.Wins); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func eloExpected(a, b int) float64 {
	return 1.0 / (1.0 + math.Pow(10, float64(b-a)/400.0))
}

func newElo(current int, expected, score float64) int {
	return int(math.Round(float64(current) + 32*(score-expected)))
}
