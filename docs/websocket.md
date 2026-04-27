# WebSocket Protocol

ChessMaster Pro uses a single WebSocket endpoint for live multiplayer.

```
ws://host/ws
```

The same HTTP server handles both REST and WebSocket traffic; the `WS` handler is
registered at `/ws` by `httpapi.NewRouter`.

---

## Connection Lifecycle

1. Client opens a WebSocket connection to `ws://host/ws`.
2. Client **must** send a `join` message as the very first frame. The server
   rejects any other message type at this stage and closes the connection.
3. The server authenticates the JWT embedded in the `join` message and verifies
   that the player is a participant in the requested game.
4. On success the server sends an initial `state` message describing the current
   board.
5. Both clients exchange `move` (and optionally `resign`) messages until the
   game ends.
6. When the game ends the server broadcasts a final `state` message followed by a
   `gameOver` message to all connected clients.
7. Either side may close the TCP connection at any time; the server cleans up the
   room entry silently.

All messages are JSON objects delivered as WebSocket **text frames**.

---

## Inbound Messages (client → server)

### `join`

Authenticates the player and subscribes them to a game room. Must be the first
message sent on every new connection.

```json
{
  "type":   "join",
  "gameId": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "token":  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

| Field | Type | Description |
|---|---|---|
| `type` | string | Must be `"join"` |
| `gameId` | string (UUID) | The game to join |
| `token` | string | Bearer JWT obtained from `/auth/login` or `/auth/register` |

### `move`

Submits a chess move. The `uci` field is ignored on the server if it is not the
sender's turn.

```json
{
  "type": "move",
  "uci":  "e2e4"
}
```

| Field | Type | Description |
|---|---|---|
| `type` | string | Must be `"move"` |
| `uci` | string | Move in UCI notation (e.g. `e2e4`, `e7e8q`, `e1g1` for castling) |

The `gameId` and `token` fields are ignored after the initial `join`.

### `resign`

Forfeits the game immediately. The resigning player loses regardless of the
board position.

```json
{
  "type": "resign"
}
```

---

## Outbound Messages (server → client)

### `state`

Sent after every `join` and after every legal move. Broadcast to all clients in
the room.

```json
{
  "type":       "state",
  "fen":        "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
  "pgn":        "1. e4 *",
  "status":     "Ongoing",
  "legalMoves": ["e7e5", "d7d5", "g8f6"],
  "sideToMove": "black"
}
```

| Field | Type | Description |
|---|---|---|
| `type` | string | Always `"state"` |
| `fen` | string | FEN of the current position |
| `pgn` | string | Full PGN of the game so far |
| `status` | string | One of: `Ongoing`, `Check`, `Checkmate`, `Stalemate`, `DrawFiftyMove`, `DrawInsufficientMaterial`, `DrawByRepetition` |
| `legalMoves` | array of string | All legal moves for the side to move, in UCI notation |
| `sideToMove` | string | `"white"` or `"black"` |

### `error`

Sent only to the client that caused the error; not broadcast.

```json
{
  "type":    "error",
  "message": "not your turn"
}
```

| Field | Type | Description |
|---|---|---|
| `type` | string | Always `"error"` |
| `message` | string | Human-readable error description |

### `gameOver`

Broadcast to all clients in the room immediately after the terminal `state` message.
Only sent when the game has ended (checkmate, stalemate, resignation, draw).

```json
{
  "type":   "gameOver",
  "result": "1-0"
}
```

| Field | Type | Description |
|---|---|---|
| `type` | string | Always `"gameOver"` |
| `result` | string | `"1-0"` (white wins), `"0-1"` (black wins), or `"1/2-1/2"` (draw) |

---

## Example Flow — Scholar's Mate (4-move sequence)

```
Client A (white)           Server                   Client B (black)
     |                        |                           |
     |--- join {gameId,tok} -->|                           |
     |<-- state (initial) ----|--- join {gameId,tok} ---->|
     |                        |<-- state (initial) -------|
     |--- move {uci:e2e4} --->|                           |
     |<-- state (e4 played) --|--- state (e4 played) ---->|
     |                        |<-- move {uci:e7e5} -------|
     |<-- state (e5 played) --|--- state (e5 played) ---->|
     |--- move {uci:f1c4} --->|                           |
     |<-- state (Bc4 played)--|--- state (Bc4 played) --->|
     |                        |<-- move {uci:b8c6} -------|
     |<-- state (Nc6 played)--|--- state (Nc6 played) --->|
     |--- move {uci:d1h5} --->|                           |
     |<-- state (Qh5 played)--|--- state (Qh5 played) --->|
     |                        |<-- move {uci:a7a6} -------|
     |<-- state (a6 played) --|--- state (a6 played) ---->|
     |--- move {uci:h5f7} --->|  (checkmate)              |
     |<-- state (status:Checkmate)                        |
     |<-- gameOver {result:1-0}                           |
     |                        |--- state (Checkmate) ---->|
     |                        |--- gameOver {1-0} ------->|
```

---

## Error Cases

| Scenario | Server response |
|---|---|
| First message is not `join` | `error: "first message must be join"`, connection closed |
| JWT is missing or expired | `error: "unauthorized"`, connection closed |
| `gameId` is not a valid UUID | `error: "invalid gameId"`, connection closed |
| Game does not exist | `error: "game not found"`, connection closed |
| Player is not in the game | `error: "you are not a participant in this game"`, connection closed |
| Move played out of turn | `error: "not your turn"` (connection stays open) |
| UCI string is malformed | `error: "invalid move: <details>"` (connection stays open) |
| Move is illegal in the position | `error: "illegal move: <details>"` (connection stays open) |

---

## Notes

- The server stores every move in the database via `MoveRepo.Append` and updates
  the game row via `GameRepo.UpdateStatus`, so REST endpoints stay consistent with
  the live board.
- Elo ratings are updated atomically in the database as soon as a game-ending
  move or resignation is processed.
- The `/ws` endpoint is not available in dev mode when `WS` is nil
  (it will 404). In practice `main.go` always wires the hub.
