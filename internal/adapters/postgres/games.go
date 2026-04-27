package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Games implements ports.GameRepo against PostgreSQL.
type Games struct {
	pool *pgxpool.Pool
}

// NewGames creates a Games repo backed by the given pool.
func NewGames(pool *pgxpool.Pool) *Games { return &Games{pool: pool} }

// Create inserts a new game and returns it with generated fields.
func (r *Games) Create(ctx context.Context, g ports.Game) (ports.Game, error) {
	const q = `
		INSERT INTO games (white_id, black_id, mode, status, invite_code, pgn)
		VALUES ($1, $2, $3, $4, $5, '')
		RETURNING id, white_id, black_id, mode, status, invite_code, pgn, result, started_at, ended_at, created_at`
	row := r.pool.QueryRow(ctx, q, g.WhiteID, g.BlackID, g.Mode, g.Status, g.InviteCode)
	return scanGame(row)
}

// ByID fetches a game by UUID.
func (r *Games) ByID(ctx context.Context, id uuid.UUID) (ports.Game, error) {
	const q = `SELECT id, white_id, black_id, mode, status, invite_code, pgn, result, started_at, ended_at, created_at FROM games WHERE id=$1`
	row := r.pool.QueryRow(ctx, q, id)
	g, err := scanGame(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.Game{}, errNotFound
	}
	return g, err
}

// ByInviteCode fetches a game by its invite code.
func (r *Games) ByInviteCode(ctx context.Context, code string) (ports.Game, error) {
	const q = `SELECT id, white_id, black_id, mode, status, invite_code, pgn, result, started_at, ended_at, created_at FROM games WHERE invite_code=$1`
	row := r.pool.QueryRow(ctx, q, code)
	g, err := scanGame(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.Game{}, errNotFound
	}
	return g, err
}

// UpdateStatus sets the status, result, and pgn on a game.
func (r *Games) UpdateStatus(ctx context.Context, id uuid.UUID, status, result, pgn string) error {
	const q = `
		UPDATE games
		SET status=$2,
		    result=CASE WHEN $3='' THEN result ELSE $3 END,
		    pgn=$4,
		    ended_at=CASE WHEN $2 IN ('white_won','black_won','draw','aborted') THEN NOW() ELSE ended_at END,
		    started_at=CASE WHEN started_at IS NULL AND $2='active' THEN NOW() ELSE started_at END
		WHERE id=$1`
	tag, err := r.pool.Exec(ctx, q, id, status, result, pgn)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errNotFound
	}
	return nil
}

// JoinAsBlack sets the black player and transitions the game to active.
func (r *Games) JoinAsBlack(ctx context.Context, id uuid.UUID, blackID uuid.UUID) error {
	const q = `UPDATE games SET black_id=$2, status='active', started_at=NOW() WHERE id=$1 AND black_id IS NULL`
	tag, err := r.pool.Exec(ctx, q, id, blackID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errNotFound
	}
	return nil
}

// ListByPlayer returns up to limit games where the player is white or black, newest first.
func (r *Games) ListByPlayer(ctx context.Context, playerID uuid.UUID, limit int) ([]ports.Game, error) {
	const q = `
		SELECT id, white_id, black_id, mode, status, invite_code, pgn, result, started_at, ended_at, created_at
		FROM games WHERE white_id=$1 OR black_id=$1
		ORDER BY created_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, q, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var games []ports.Game
	for rows.Next() {
		g, err := scanGame(rows)
		if err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, rows.Err()
}

type gameScanner interface {
	Scan(dest ...any) error
}

func scanGame(row gameScanner) (ports.Game, error) {
	var g ports.Game
	var startedAt, endedAt *time.Time
	err := row.Scan(
		&g.ID, &g.WhiteID, &g.BlackID,
		&g.Mode, &g.Status, &g.InviteCode,
		&g.PGN, &g.Result,
		&startedAt, &endedAt, &g.CreatedAt,
	)
	if err != nil {
		return ports.Game{}, err
	}
	g.StartedAt = startedAt
	g.EndedAt = endedAt
	return g, nil
}
