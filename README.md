# ChessMaster Pro

> Premium chess with stakes — bet Pawns, play live, get AI-coached. Built for Kazakhstan.

ChessMaster Pro is a real-time multiplayer chess platform where every game can carry a Pawn wager. Players earn internal currency through wins, streaks, and daily logins; Pro subscribers receive a 2x multiplier. After each game an AI Coach (Google Gemini) analyses mistakes and suggests improvements — the first chess platform in the region that combines live stakes with structured coaching.

---

## Stack

| Layer | Technology |
|---|---|
| Backend language | Go 1.22+ |
| Database | PostgreSQL 16 |
| DB driver | pgx/v5 (raw SQL, no ORM) |
| HTTP router | chi v5 |
| WebSocket | coder/websocket |
| Auth | golang-jwt/jwt v5 (HS256) |
| Chess AI | Stockfish (UCI, per-call spawn) |
| Game coach | Google Gemini 2.0 Flash |
| Frontend | Next.js 14 App Router |
| Styling | Tailwind CSS 3 |
| Language | TypeScript (strict) |
| Infra | Docker + docker-compose |

---

## Architecture

Hexagonal architecture: domain logic has zero external dependencies; adapters plug in at the edges.

```mermaid
graph LR
    subgraph Clients
        B[Browser HTTP]
        WS[Browser WS]
    end

    subgraph Adapters-In
        H[httpapi]
        W[wsroom]
    end

    subgraph Ports
        PR[PlayerRepo]
        GR[GameRepo]
        MR[MoveRepo]
        RT[RatingRepo]
        EN[Engine]
        CO[Coach]
    end

    subgraph Domain
        CH[domain/chess]
    end

    subgraph Adapters-Out
        PG[(postgres)]
        MEM[(memrepo)]
        SF[stockfish]
        GM[gemini]
        JWT[auth/jwt]
    end

    B --> H
    WS --> W
    H --> PR
    H --> GR
    H --> MR
    H --> RT
    H --> EN
    H --> CO
    W --> GR
    W --> MR
    W --> RT
    H --> CH
    W --> CH
    PR --> PG
    PR --> MEM
    GR --> PG
    GR --> MEM
    MR --> PG
    MR --> MEM
    RT --> PG
    RT --> MEM
    EN --> SF
    CO --> GM
    H --> JWT
    W --> JWT
```

---

## Project Layout

```
.
├── cmd/
│   └── api/            # HTTP + WS server entrypoint
├── internal/
│   ├── domain/
│   │   └── chess/      # Pure chess rules engine (board, moves, FEN, PGN, status)
│   ├── ports/          # Interface definitions (PlayerRepo, GameRepo, Engine, Coach …)
│   └── adapters/
│       ├── httpapi/    # Chi HTTP handlers + router
│       ├── wsroom/     # WebSocket hub and game rooms
│       ├── postgres/   # pgx/v5 implementations of all repos
│       ├── memrepo/    # In-memory implementations (dev / tests)
│       ├── stockfish/  # UCI engine adapter (per-call spawn)
│       ├── gemini/     # Gemini REST coach adapter
│       └── auth/jwt/   # HS256 signer + bcrypt helpers + middleware
├── pkg/
│   └── code/           # Cryptographically random invite code generator
├── migrations/         # golang-migrate SQL migration files
├── web/                # Next.js 14 frontend
├── docs/
│   ├── openapi.yaml    # OpenAPI 3.1 spec
│   ├── websocket.md    # WebSocket protocol spec
│   └── adr/            # Architecture Decision Records
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

---

## Getting Started

### Option 1 — Docker Compose (recommended)

Requires Docker with Compose v2.

```bash
# set secrets (only GEMINI_API_KEY is optional)
export JWT_SECRET=your-secret-here
export GEMINI_API_KEY=your-gemini-key   # optional; disables AI coach if unset

make docker-up          # builds + starts postgres, migrate, stockfish, api, frontend
# api  → http://localhost:8080
# web  → http://localhost:3000
```

### Option 2 — Local Go + Postgres

Prerequisites: Go 1.22+, PostgreSQL 16, `golang-migrate` CLI, Stockfish binary.

```bash
export POSTGRES_URL=postgres://chess:chess@localhost:5432/chess?sslmode=disable
export JWT_SECRET=dev-secret
export PORT=8080
export STOCKFISH_PATH=/usr/games/stockfish   # or wherever stockfish lives
export GEMINI_API_KEY=your-key               # optional

make migrate      # run SQL migrations
make run          # go run ./cmd/api
```

### Option 3 — Dev mode with in-memory repos

No Postgres needed. Omit `POSTGRES_URL` and the server boots with in-memory repositories.

```bash
JWT_SECRET=dev PORT=8080 go run ./cmd/api
```

### Frontend (development)

```bash
cd web
npm install
npm run dev      # http://localhost:3000
```

---

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `POSTGRES_URL` | No | — | PostgreSQL DSN; omit for in-memory mode |
| `JWT_SECRET` | Yes | `dev-secret-change-me` | HMAC secret for JWT tokens |
| `PORT` | No | `8080` | HTTP listen port |
| `STOCKFISH_PATH` | No | `stockfish` | Path to Stockfish binary |
| `GEMINI_API_KEY` | No | — | Google AI API key; omit to disable coach |

Get a free Gemini API key at <https://aistudio.google.com/app/apikey>.

---

## API

Full spec: [`docs/openapi.yaml`](docs/openapi.yaml)

| Method | Path | Auth | Summary |
|---|---|---|---|
| GET | `/healthz` | — | Health probe |
| POST | `/auth/register` | — | Register new player, returns JWT |
| POST | `/auth/login` | — | Login, returns JWT |
| GET | `/me` | Bearer | Current player profile |
| POST | `/me/upgrade` | Bearer | Upgrade account to Pro |
| POST | `/games` | Bearer | Create new game (pvp / ai_easy / ai_medium / ai_hard) |
| POST | `/games/join` | Bearer | Join a PvP game by invite code |
| GET | `/games/{id}` | — | Get game + move list |
| GET | `/games/{id}/moves` | — | Get move list for a game |
| POST | `/games/{id}/move` | Bearer | Submit a move (UCI notation) |
| POST | `/games/{id}/coach` | Bearer | Request AI coach analysis |
| GET | `/players/me/games` | Bearer | List authenticated player's games |
| GET | `/leaderboard` | — | Top players, optional `?city=&limit=` |

---

## WebSocket

Protocol spec: [`docs/websocket.md`](docs/websocket.md)

Connect to `ws://host/ws` for live multiplayer. Authentication happens via the first `join` message (no HTTP headers needed).

---

## Testing

```bash
# Unit + integration (unit tests run anywhere; integration tests require POSTGRES_TEST_URL)
make test
# or
go test ./... -count=1

# With race detector
make test-race

# Postgres integration tests
POSTGRES_TEST_URL=postgres://chess:chess@localhost:5432/chess_test?sslmode=disable \
  go test ./internal/adapters/postgres/... -count=1
```

The chess engine is validated by a perft(3) test: starting position must yield exactly 8 902 nodes at depth 3. This confirms move generation correctness without an external oracle.

---

## ADRs

| # | Title |
|---|---|
| [ADR-0001](docs/adr/0001-hexagonal-architecture.md) | Hexagonal Architecture |
| [ADR-0002](docs/adr/0002-pgx-no-orm.md) | pgx Without ORM |
| [ADR-0003](docs/adr/0003-stockfish-per-call-spawn.md) | Stockfish Per-Call Spawn |
| [ADR-0004](docs/adr/0004-gemini-coach-stdlib-http.md) | Gemini Coach via stdlib net/http |
| [ADR-0005](docs/adr/0005-in-memory-repos-for-dev.md) | In-Memory Repos for Dev |

---

## Roadmap

- Owner-only AI coach (restrict `/games/{id}/coach` to game participants)
- Stockfish process pool to eliminate per-move spawn latency
- Stripe Pro upgrade with real payment flow
- Pawn betting UI with pre-game wager modal
- Frontend end-to-end tests (Playwright)
- Daily login rewards and achievement system
- Mobile PWA with push notifications

---

## License

MIT — see [LICENSE](LICENSE) (placeholder; not yet committed).
