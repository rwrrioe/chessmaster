# ADR-0003: Stockfish Per-Call Process Spawn

Date: 2026-04-27
Status: Accepted

---

## Context

The platform supports AI opponents at three skill levels (easy / medium / hard).
After each human move in an AI game the server must ask Stockfish for a reply.
Stockfish is a native binary that communicates over stdin/stdout using the UCI
protocol. There are two common ways to manage the process lifetime:

**Option A — Per-call spawn**
Spawn a fresh Stockfish process per move, write the UCI commands, read the
`bestmove` line, then let the process exit.

**Option B — Long-lived process (pool)**
Keep one (or several) Stockfish processes running. Send `position` + `go`
commands per move; reuse the same process across calls.

---

## Decision

Use Option A: per-call spawn.

The `stockfish.Engine.BestMove` method creates an `exec.Cmd`, pipes UCI commands
(`uci`, `setoption`, `isready`, `position fen …`, `go movetime …`), reads the
`bestmove` response, then returns. The process exits when stdin is closed.

A `context.WithTimeout` wraps each call with a deadline equal to twice the
configured `movetime` plus 3 seconds, so a hung Stockfish process is always
killed.

---

## Consequences

**Positive**
- Zero state shared between calls; no synchronisation, no pool bookkeeping.
- Simple error handling: if the process fails to start (binary missing, OS
  error), `BestMove` returns an error and the HTTP handler degrades gracefully
  — the human's move is still recorded and the AI turn is skipped.
- The entire adapter is ~115 lines including error handling. Easy to read and
  test with a fake binary (see `testdata/fakesf`).

**Negative / trade-offs**
- Process spawn cost (~5–20 ms on Linux) adds latency to every AI move. At the
  current scale (prototype / low concurrency) this is imperceptible to users.
- UCI initialisation (`uci` / `uciok` handshake, `isready` / `readyok`) is
  repeated on every call, adding ~10 ms on top of spawn cost.
- High concurrency (hundreds of simultaneous AI games) would cause significant
  overhead. A process pool with a fixed number of long-lived Stockfish workers
  (see Roadmap in README) is the natural next step.
- Stockfish does not benefit from hash table warmup between calls; analysis
  quality is slightly lower than a persistent process would provide at equivalent
  depth. Acceptable for easy/medium difficulty.

The per-call approach is the right default for a prototype: correctness and
simplicity first, optimise when load requires it.
