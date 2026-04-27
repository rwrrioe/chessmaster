# ADR-0002: pgx Without ORM

Date: 2026-04-27
Status: Accepted

---

## Context

The schema is small and stable: five tables (`players`, `games`, `moves`,
`ratings`, `pawns_ledger`). The queries are predominantly straightforward
INSERT/SELECT/UPDATE statements, with two slightly non-trivial cases: the Elo
`FOR UPDATE` lock in a transaction, and the leaderboard join between `ratings`
and `players`.

Options considered:

| Approach | Pros | Cons |
|---|---|---|
| GORM | Least SQL to write | Opaque SQL, migrations via tags, hidden N+1, import weight |
| sqlx | Thin reflection helper | Adds a dependency for marginal gain; struct tags drive scanning |
| ent | Code-gen, strong types | Heavy code-gen pipeline, steep learning curve |
| pgx/v5 raw | Full control, zero magic, native PostgreSQL types | More boilerplate per query |

The project already adopted pgx/v5 as the driver for its `pgxpool` connection
pool (the only practical choice for Go + PostgreSQL 16 without cgo). Given the
small query surface, writing SQL directly costs less than introducing a second
abstraction on top.

---

## Decision

Use `pgx/v5` directly with raw SQL string constants. No ORM, no `sqlx`, no
code generation.

Each repo struct holds a `*pgxpool.Pool` (not a single `*pgx.Conn`) so all
repository operations are safe for concurrent use from multiple HTTP goroutines
without additional pooling logic.

Conventions adopted:
- Each public repo type (`Players`, `Games`, `Moves`, `Ratings`) lives in its
  own file named after the resource.
- SQL is defined as `const q = ...` local to each method; never built with
  `fmt.Sprintf`.
- `pgx.Row` and `pgx.Rows` are scanned into port types via small private helper
  functions (`scanPlayer`, `scanGame`).
- The `FOR UPDATE` advisory lock in `ApplyResult` uses a single transaction with
  explicit row locking to prevent concurrent Elo races.

---

## Consequences

**Positive**
- SQL is auditable at a glance; no ORM query translation layer to debug.
- pgx/v5 supports PostgreSQL-native types (UUIDs, arrays, JSONB) without
  reflection hacks.
- Migrations remain SQL files managed by `golang-migrate`, keeping schema
  evolution visible in version control.

**Negative / trade-offs**
- Each new column requires updating both the INSERT/SELECT query and the
  corresponding scan call. Acceptable overhead for a table count this small.
- No query builder means string duplication of column lists (e.g. the SELECT
  column list in `ByID` and `ListByPlayer`). A future extraction to named
  constants would reduce drift.
