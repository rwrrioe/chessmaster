package postgres

import (
	"context"

	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Moves implements ports.MoveRepo against PostgreSQL.
type Moves struct {
	pool *pgxpool.Pool
}

// NewMoves creates a Moves repo backed by the given pool.
func NewMoves(pool *pgxpool.Pool) *Moves { return &Moves{pool: pool} }

// Append inserts a single move row.
func (r *Moves) Append(ctx context.Context, gameID uuid.UUID, ply int, uci, san, fenAfter string) error {
	const q = `INSERT INTO moves (game_id, ply, uci, san, fen_after) VALUES ($1,$2,$3,$4,$5)`
	_, err := r.pool.Exec(ctx, q, gameID, ply, uci, san, fenAfter)
	return err
}

// ListByGame returns all moves for a game ordered by ply.
func (r *Moves) ListByGame(ctx context.Context, gameID uuid.UUID) ([]ports.Move, error) {
	const q = `SELECT id, game_id, ply, uci, san, fen_after, played_at FROM moves WHERE game_id=$1 ORDER BY ply`
	rows, err := r.pool.Query(ctx, q, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var moves []ports.Move
	for rows.Next() {
		var m ports.Move
		if err := rows.Scan(&m.ID, &m.GameID, &m.Ply, &m.UCI, &m.SAN, &m.FENAfter, &m.PlayedAt); err != nil {
			return nil, err
		}
		moves = append(moves, m)
	}
	return moves, rows.Err()
}
