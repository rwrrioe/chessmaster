package postgres_test

import (
	"context"
	"os"
	"testing"

	pgadapter "github.com/chessmaster-pro/chessmaster/internal/adapters/postgres"
	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// migrationSQL is loaded at test time from the repo root.
const migrationSQL = `
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE IF NOT EXISTS players (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email        TEXT UNIQUE NOT NULL,
    username     TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    city         TEXT,
    is_pro       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_players_city ON players(city);
DO $$ BEGIN
  CREATE TYPE game_status AS ENUM ('pending','active','white_won','black_won','draw','aborted');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN
  CREATE TYPE game_mode AS ENUM ('pvp','ai_easy','ai_medium','ai_hard');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
CREATE TABLE IF NOT EXISTS games (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    white_id    UUID REFERENCES players(id) ON DELETE SET NULL,
    black_id    UUID REFERENCES players(id) ON DELETE SET NULL,
    mode        game_mode NOT NULL,
    status      game_status NOT NULL DEFAULT 'pending',
    invite_code TEXT UNIQUE,
    pgn         TEXT NOT NULL DEFAULT '',
    result      TEXT,
    started_at  TIMESTAMPTZ,
    ended_at    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS moves (
    id        BIGSERIAL PRIMARY KEY,
    game_id   UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    ply       INTEGER NOT NULL,
    uci       TEXT NOT NULL,
    san       TEXT NOT NULL,
    fen_after TEXT NOT NULL,
    played_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, ply)
);
CREATE TABLE IF NOT EXISTS ratings (
    player_id  UUID PRIMARY KEY REFERENCES players(id) ON DELETE CASCADE,
    elo        INTEGER NOT NULL DEFAULT 1200,
    games      INTEGER NOT NULL DEFAULT 0,
    wins       INTEGER NOT NULL DEFAULT 0,
    losses     INTEGER NOT NULL DEFAULT 0,
    draws      INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("POSTGRES_TEST_URL")
	if url == "" {
		t.Skip("POSTGRES_TEST_URL not set; skipping integration tests")
	}
	ctx := context.Background()
	pool, err := pgadapter.New(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, migrationSQL)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pool
}

func TestIntegration_Players(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := pgadapter.NewPlayers(pool)

	suffix := uuid.New().String()[:8]
	p, err := repo.Create(ctx, ports.Player{
		Email:        "test_" + suffix + "@example.com",
		Username:     "user_" + suffix,
		PasswordHash: "hash",
		City:         "Almaty",
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.ID == (uuid.UUID{}) {
		t.Fatal("id not generated")
	}

	got, err := repo.ByEmail(ctx, p.Email)
	if err != nil {
		t.Fatal(err)
	}
	if got.Username != p.Username {
		t.Fatalf("username mismatch: %s", got.Username)
	}

	got2, err := repo.ByID(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got2.City != "Almaty" {
		t.Fatalf("city: %s", got2.City)
	}

	if err = repo.SetPro(ctx, p.ID, true); err != nil {
		t.Fatal(err)
	}
}

func TestIntegration_Games(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	players := pgadapter.NewPlayers(pool)
	games := pgadapter.NewGames(pool)
	moves := pgadapter.NewMoves(pool)

	suffix := uuid.New().String()[:8]
	p1, _ := players.Create(ctx, ports.Player{Email: "g1_" + suffix + "@x.com", Username: "g1_" + suffix, PasswordHash: "h"})
	p2, _ := players.Create(ctx, ports.Player{Email: "g2_" + suffix + "@x.com", Username: "g2_" + suffix, PasswordHash: "h"})

	code := "INV" + suffix[:5]
	g, err := games.Create(ctx, ports.Game{
		WhiteID:    &p1.ID,
		Mode:       "pvp",
		Status:     "pending",
		InviteCode: &code,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = games.JoinAsBlack(ctx, g.ID, p2.ID); err != nil {
		t.Fatal(err)
	}

	if err = moves.Append(ctx, g.ID, 1, "e2e4", "e4", "fen"); err != nil {
		t.Fatal(err)
	}

	ms, err := moves.ListByGame(ctx, g.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 1 {
		t.Fatalf("expected 1 move, got %d", len(ms))
	}

	if err = games.UpdateStatus(ctx, g.ID, "white_won", "1-0", "pgn"); err != nil {
		t.Fatal(err)
	}

	gs, err := games.ListByPlayer(ctx, p1.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(gs) == 0 {
		t.Fatal("no games returned")
	}
}

func TestIntegration_Ratings(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	players := pgadapter.NewPlayers(pool)
	ratings := pgadapter.NewRatings(pool)

	suffix := uuid.New().String()[:8]
	p1, _ := players.Create(ctx, ports.Player{Email: "r1_" + suffix + "@x.com", Username: "r1_" + suffix, PasswordHash: "h"})
	p2, _ := players.Create(ctx, ports.Player{Email: "r2_" + suffix + "@x.com", Username: "r2_" + suffix, PasswordHash: "h"})

	r, err := ratings.Get(ctx, p1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if r.Elo != 1200 {
		t.Fatalf("default elo: %d", r.Elo)
	}

	if err = ratings.ApplyResult(ctx, &p1.ID, &p2.ID, "1-0"); err != nil {
		t.Fatal(err)
	}

	r1, _ := ratings.Get(ctx, p1.ID)
	r2, _ := ratings.Get(ctx, p2.ID)
	if r1.Elo <= 1200 {
		t.Fatalf("white elo should rise: %d", r1.Elo)
	}
	if r2.Elo >= 1200 {
		t.Fatalf("black elo should fall: %d", r2.Elo)
	}
}
