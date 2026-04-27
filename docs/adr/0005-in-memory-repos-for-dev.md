# ADR-0005: In-Memory Repos for Dev and Testing

Date: 2026-04-27
Status: Accepted

---

## Context

Running the full stack locally requires Docker (for PostgreSQL), `golang-migrate`,
and the correct environment variables. This creates friction for:

- A developer who wants to experiment with the HTTP API without Docker.
- Unit and integration tests that run in CI without a live database.
- The test suite for `httpapi` handlers, which needs repositories but not real
  SQL.

---

## Decision

Provide a complete second implementation of all port interfaces in
`internal/adapters/memrepo`:

| Port interface | memrepo type |
|---|---|
| `ports.PlayerRepo` | `memrepo.Players` |
| `ports.GameRepo` | `memrepo.Games` |
| `ports.MoveRepo` | `memrepo.Moves` |
| `ports.RatingRepo` | `memrepo.Ratings` |

All four are backed by in-memory maps protected by `sync.RWMutex` (or
`sync.Mutex` where writes dominate). They generate UUIDs and timestamps
locally using `github.com/google/uuid` and `time.Now()`.

`cmd/api/main.go` branches on `POSTGRES_URL`:
- Present → use `postgres` adapters.
- Absent → use `memrepo` adapters and log `"running with in-memory repos (dev)"`.

This makes the server runnable with a single command and no external processes.

---

## Consequences

**Positive**
- Zero-infrastructure dev mode: `JWT_SECRET=dev go run ./cmd/api` is enough.
- HTTP handler tests (`httpapi/handlers_test.go`, `httpapi/router_test.go`) use
  `memrepo` directly; they execute in milliseconds and run anywhere.
- `memrepo.Ratings` requires a reference to `*memrepo.Players` to resolve
  usernames for the leaderboard. This is the only cross-repo dependency and is
  passed explicitly at construction time — no global state.

**Negative / trade-offs**
- Behaviour differences between memrepo and postgres can mask bugs. For
  example, the memrepo leaderboard iterates a map (unordered before the sort),
  while the postgres query uses `ORDER BY elo DESC`. Both produce the same
  result but the code paths are different.
- Data is lost on process restart; not suitable for any persistent staging
  environment.
- The Elo formula is duplicated between `memrepo/ratings.go` and
  `postgres/ratings.go`. An extraction to `pkg/elo` would eliminate the drift
  risk, but is deferred until a third implementation would be needed.
