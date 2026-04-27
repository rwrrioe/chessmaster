package postgres

import (
	"context"
	"errors"

	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Players implements ports.PlayerRepo against PostgreSQL.
type Players struct {
	pool *pgxpool.Pool
}

// NewPlayers creates a Players repo backed by the given pool.
func NewPlayers(pool *pgxpool.Pool) *Players { return &Players{pool: pool} }

// Create inserts a new player row and returns it with the generated ID and timestamp.
func (r *Players) Create(ctx context.Context, p ports.Player) (ports.Player, error) {
	const q = `
		INSERT INTO players (email, username, password_hash, city, is_pro)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, username, password_hash, city, is_pro, created_at`
	row := r.pool.QueryRow(ctx, q, p.Email, p.Username, p.PasswordHash, nullStr(p.City), p.IsPro)
	return scanPlayer(row)
}

// ByEmail fetches a player by email address.
func (r *Players) ByEmail(ctx context.Context, email string) (ports.Player, error) {
	const q = `SELECT id, email, username, password_hash, city, is_pro, created_at FROM players WHERE email=$1`
	row := r.pool.QueryRow(ctx, q, email)
	p, err := scanPlayer(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.Player{}, errNotFound
	}
	return p, err
}

// ByID fetches a player by UUID.
func (r *Players) ByID(ctx context.Context, id uuid.UUID) (ports.Player, error) {
	const q = `SELECT id, email, username, password_hash, city, is_pro, created_at FROM players WHERE id=$1`
	row := r.pool.QueryRow(ctx, q, id)
	p, err := scanPlayer(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.Player{}, errNotFound
	}
	return p, err
}

// UpdateCity sets the city for a player.
func (r *Players) UpdateCity(ctx context.Context, id uuid.UUID, city string) error {
	const q = `UPDATE players SET city=$2 WHERE id=$1`
	tag, err := r.pool.Exec(ctx, q, id, city)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errNotFound
	}
	return nil
}

// SetPro sets the is_pro flag for a player.
func (r *Players) SetPro(ctx context.Context, id uuid.UUID, isPro bool) error {
	const q = `UPDATE players SET is_pro=$2 WHERE id=$1`
	tag, err := r.pool.Exec(ctx, q, id, isPro)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errNotFound
	}
	return nil
}

func scanPlayer(row pgx.Row) (ports.Player, error) {
	var p ports.Player
	var city *string
	if err := row.Scan(&p.ID, &p.Email, &p.Username, &p.PasswordHash, &city, &p.IsPro, &p.CreatedAt); err != nil {
		return ports.Player{}, err
	}
	if city != nil {
		p.City = *city
	}
	return p, nil
}
