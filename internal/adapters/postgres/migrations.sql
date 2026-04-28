CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE players (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email        TEXT UNIQUE NOT NULL,
    username     TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    city         TEXT,
    is_pro       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_players_city ON players(city);

CREATE TYPE game_status AS ENUM ('pending', 'active', 'white_won', 'black_won', 'draw', 'aborted');
CREATE TYPE game_mode   AS ENUM ('pvp', 'ai_easy', 'ai_medium', 'ai_hard');

CREATE TABLE games (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    white_id    UUID REFERENCES players(id) ON DELETE SET NULL,
    black_id    UUID REFERENCES players(id) ON DELETE SET NULL,
    mode        game_mode   NOT NULL,
    status      game_status NOT NULL DEFAULT 'pending',
    invite_code TEXT UNIQUE,
    pgn         TEXT NOT NULL DEFAULT '',
    result      TEXT,
    started_at  TIMESTAMPTZ,
    ended_at    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_games_white  ON games(white_id);
CREATE INDEX idx_games_black  ON games(black_id);
CREATE INDEX idx_games_status ON games(status);

CREATE TABLE moves (
    id        BIGSERIAL PRIMARY KEY,
    game_id   UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    ply       INTEGER NOT NULL,
    uci       TEXT NOT NULL,
    san       TEXT NOT NULL,
    fen_after TEXT NOT NULL,
    played_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, ply)
);

CREATE INDEX idx_moves_game ON moves(game_id);

CREATE TABLE ratings (
    player_id  UUID PRIMARY KEY REFERENCES players(id) ON DELETE CASCADE,
    elo        INTEGER NOT NULL DEFAULT 1200,
    games      INTEGER NOT NULL DEFAULT 0,
    wins       INTEGER NOT NULL DEFAULT 0,
    losses     INTEGER NOT NULL DEFAULT 0,
    draws      INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ratings_elo ON ratings(elo DESC);
