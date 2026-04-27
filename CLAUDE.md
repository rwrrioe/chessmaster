# ChessMaster Pro — Go + PostgreSQL

## Skills (read all before starting)
@.claude/skills/senior-backend/SKILL.md
@.claude/skills/senior-frontend/SKILL.md
@.claude/skills/design/SKILL.md

## Product Vision
Premium chess platform with internal currency "Pawns" (Пешки).
Not a toy — a startup prototype for competitive players.
Unique angle: bet Pawns before each game + AI Coach after game.
Target: Kazakhstan players (Almaty, Astana, Shymkent).

## Unique Niche — Chess with Stakes
- Every user starts with 100 Pawns
- Before game: both players bet (10/25/50/100 Pawns)
- Winner takes all
- Pawns earned also by: daily login, achievements, streak
- Leaderboard shows Pawns + city ranking
- Pro users get 2x Pawn multiplier (monetization hook)

## Architecture
Hexagonal. Three layers only: domain → ports → adapters.

cmd/
  api/          → HTTP server entrypoint
  ws/           → WebSocket server entrypoint
internal/
  domain/
    game/       → Game, Move, Board, Piece, Rules
    player/     → Player, Rating, Pawns, Stats
    coach/      → MistakeDetector, Analysis
    betting/    → Bet, BettingPool, Transaction
  ports/
    in/         → HTTP + WS handler interfaces
    out/        → repository interfaces
  adapters/
    postgres/   → pgx/v5, raw SQL, no ORM
    http/       → Chi router
    ws/         → gorilla/websocket
    stockfish/  → engine adapter
    gemini/     → AI coach via Gemini API
frontend/
  app/
  components/
    board/
    coach/
    leaderboard/
    betting/
    pro/
docker/
Makefile
docker-compose.yml

## Backend Rules
- Go 1.22+, PostgreSQL 16
- pgx/v5 only, zero ORM
- golang-migrate for migrations
- TDD: failing test first, then implement
- Table-driven tests
- Minimal comments — exported functions only
- Conventional commits: feat/fix/test/docs/chore
- Auto-push after each feature

## Frontend Rules
- Next.js 14 App Router, TypeScript strict
- NO Inter/Roboto/Arial/Space Grotesk
- NO purple gradients, NO generic AI look
- Premium dark theme — obsidian black, gold accent (#F5C518)
- Chess pieces: custom SVG, not default unicode
- Framer Motion for board animations
- Mobile-first responsive

## Features — build in this order

### Phase 1 — Foundation
- Go module + Next.js init
- docker-compose: postgres, api, frontend, stockfish
- Makefile: test, build, run, migrate, push, docker-up, lint
- Migrations: players, games, moves, pawns_ledger, bets tables
- Git init + push to GitHub

### Phase 2 — Chess Domain (TDD)
- Board, Piece, Move, Game entities
- Full rules: castling, en passant, checkmate, stalemate, draw
- Every rule: table-driven tests

### Phase 3 — Backend APIs
- JWT auth (register/login)
- WebSocket multiplayer — play by link
- Stockfish adapter (easy/medium/hard)
- Game history API
- Pawns ledger (earn/spend/transfer)
- Betting API (create bet, accept, resolve)

### Phase 4 — AI Coach (Gemini)
- Send PGN to Gemini after game ends
- Return: mistakes, better moves, summary
- model: gemini-2.0-flash
- env: GEMINI_API_KEY (user provides own key)

### Phase 5 — Social Layer
- Leaderboard by city (Almaty, Astana, Shymkent)
- Player profiles with stats
- City ranking: most Pawns collected

### Phase 6 — Monetization
- Upgrade to Pro page
- Pro benefits: 2x Pawns multiplier, custom piece skins
- Stripe UI ready (no real payment if no time)

### Phase 7 — Frontend UI
- Obsidian dark board with gold accents
- Drag & drop pieces with Framer Motion
- Betting modal before game starts
- AI Coach panel after game (Gemini response)
- Leaderboard by city
- Pro upgrade page

### Phase 8 — Docs & Review
- Code review: security, performance, idiomatic Go
- README.md: product, architecture, how to run, screenshots
- OpenAPI 3.0 spec
- ADR in docs/adr/

## Database Schema
players: id, username, email, password_hash, city, rating, pawns, is_pro, created_at
games: id, white_id, black_id, pgn, result, mode, created_at
moves: id, game_id, move_number, from_sq, to_sq, notation, created_at
pawns_ledger: id, player_id, amount, type, description, created_at
bets: id, game_id, white_bet, black_bet, winner_id, status, created_at

## Environment
GEMINI_API_KEY=your_key_here
POSTGRES_URL=postgres://chess:chess@localhost:5432/chess?sslmode=disable
JWT_SECRET=your_secret_here
STOCKFISH_PATH=/usr/games/stockfish
PORT=8080
FRONTEND_URL=http://localhost:3000

## Makefile targets
test:     go test ./...
build:    go build ./cmd/api
run:      docker-compose up
migrate:  golang-migrate up
push:     git add -A && git commit -m "$(msg)" && git push origin main
lint:     golangci-lint run
docker-up: docker-compose up -d
docker-down: docker-compose down

## Git
- Init repo on GitHub: chess-master-pro
- Set remote origin
- Auto-push after each phase with conventional commit
- Branch: main

## Definition of Done
make test passes → git commit → git push → next phase

## README must include
- Product description (what, for whom, why valuable)
- Architecture diagram (ASCII)
- How to run locally
- How to get Gemini API key
- Screenshots or demo link
- Unique niche explanation

## Model Strategy (cost optimization)

### Use Opus for:
- Architecture decisions
- Code review (make review)
- Complex debugging
- Planning new phases
- Security audit

### Use Sonnet for:
- Feature implementation
- Writing tests
- API handlers
- Frontend components
- Database queries

### Use Haiku for:
- File scaffolding
- Boilerplate generation
- Simple refactoring
- Comment writing
- Makefile targets
- Docker/config files

## Opus Plan Mode
Before each phase run:
"Think carefully and plan the full implementation before writing any code.
Identify risks, edge cases, and dependencies first."

Use Opus only for: plan → review → architecture
Switch to Sonnet for: implementation
Switch to Haiku for: scaffolding and boilerplate