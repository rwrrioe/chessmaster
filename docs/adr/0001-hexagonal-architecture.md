# ADR-0001: Hexagonal Architecture

Date: 2026-04-27
Status: Accepted

---

## Context

Chess is a rule-heavy domain with deterministic, side-effect-free logic. At the
same time the platform needs several external concerns: a relational database,
a WebSocket hub, a UCI chess engine, a Gemini REST API, and JWT authentication.
The risk of mixing these concerns is high: business logic becomes untestable
without real infrastructure, and swapping an adapter (e.g. replacing Stockfish
with a different engine) becomes a broad refactor.

The Go standard library and the language's interface system make hexagonal layout
natural: interfaces are satisfied implicitly, the compiler enforces the boundary,
and packages have clear, single responsibilities.

---

## Decision

Structure the codebase in three concentric layers with strict dependency direction:

```
domain  ←  ports  ←  adapters
```

- `internal/domain/chess` contains the complete chess rules engine. It imports
  nothing outside the standard library. It has no concept of HTTP, databases, or
  AI.
- `internal/ports` declares Go interfaces (`PlayerRepo`, `GameRepo`, `Engine`,
  `Coach`, etc.) and the shared value types that cross the boundary (`Player`,
  `Game`, `Move`, `Analysis`, …). Ports import only `context`, `time`, and
  `github.com/google/uuid`.
- `internal/adapters/*` implement the port interfaces. Each adapter is a
  separate package that may freely import its respective third-party library.
  Adapters never import each other.

There is no "service layer". The HTTP handlers in `httpapi` compose ports
directly through the `Deps` struct; this is sufficient for a startup prototype
and avoids an extra indirection that would add no testability benefit at this
scale.

---

## Consequences

**Positive**
- The chess engine and port definitions are tested without any real I/O; the
  in-memory `memrepo` adapters provide a zero-dependency test double for all
  higher-level tests.
- Adding a new storage backend (e.g. Redis cache) means writing one new adapter
  struct that satisfies an existing interface — no domain code changes.
- Developers can run the server locally without Docker by omitting `POSTGRES_URL`
  and relying on `memrepo`.

**Negative / trade-offs**
- The HTTP handlers carry some orchestration logic (reconstructing game state
  from stored moves, deciding AI turns). This would belong in a use-case layer
  in a larger system. Accepted for now; flagged as a follow-up item.
- Duplication of the Elo calculation formula exists in both `postgres/ratings.go`
  and `memrepo/ratings.go`. Acceptable at this scale; could be extracted to
  `pkg/elo` if a third implementation were added.
