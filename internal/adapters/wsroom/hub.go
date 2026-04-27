// Package wsroom implements the WebSocket game room hub.
package wsroom

import (
	"context"
	"net/http"
	"sync"

	"github.com/chessmaster-pro/chessmaster/internal/adapters/auth/jwt"
	"github.com/chessmaster-pro/chessmaster/internal/domain/chess"
	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
)

// Hub manages active game rooms.
type Hub struct {
	mu    sync.Mutex
	rooms map[uuid.UUID]*Room

	Games   ports.GameRepo
	Moves   ports.MoveRepo
	Ratings ports.RatingRepo
	Signer  *jwt.Signer
}

// NewHub creates an empty Hub.
func NewHub(games ports.GameRepo, moves ports.MoveRepo, ratings ports.RatingRepo, signer *jwt.Signer) *Hub {
	return &Hub{
		rooms:   make(map[uuid.UUID]*Room),
		Games:   games,
		Moves:   moves,
		Ratings: ratings,
		Signer:  signer,
	}
}

func (h *Hub) room(gameID uuid.UUID) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	r, ok := h.rooms[gameID]
	if !ok {
		r = &Room{}
		h.rooms[gameID] = r
	}
	return r
}

// ServeHTTP upgrades the connection to WebSocket and handles the game session.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // CORS handled at router level
	})
	if err != nil {
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()

	// First message must be a join
	var msg inboundMsg
	if err = wsjson.Read(ctx, conn, &msg); err != nil {
		return
	}
	if msg.Type != "join" {
		writeError(ctx, conn, "first message must be join")
		return
	}

	playerID, err := h.Signer.Parse(msg.Token)
	if err != nil {
		writeError(ctx, conn, "unauthorized")
		return
	}

	gameID, err := uuid.Parse(msg.GameID)
	if err != nil {
		writeError(ctx, conn, "invalid gameId")
		return
	}

	game, err := h.Games.ByID(ctx, gameID)
	if err != nil {
		writeError(ctx, conn, "game not found")
		return
	}

	// Verify player is part of this game
	isWhite := game.WhiteID != nil && *game.WhiteID == playerID
	isBlack := game.BlackID != nil && *game.BlackID == playerID
	if !isWhite && !isBlack {
		writeError(ctx, conn, "you are not a participant in this game")
		return
	}

	room := h.room(gameID)
	room.mu.Lock()
	if room.game == nil {
		room.game = chess.NewGame()
		room.gameID = gameID
		room.whiteID = game.WhiteID
		room.blackID = game.BlackID
	}
	room.addClient(conn)
	room.mu.Unlock()

	defer func() {
		room.mu.Lock()
		room.removeClient(conn)
		room.mu.Unlock()
	}()

	// Send initial state
	h.sendState(ctx, conn, room)

	// Message loop
	for {
		var m inboundMsg
		if err = wsjson.Read(ctx, conn, &m); err != nil {
			break
		}
		switch m.Type {
		case "move":
			h.handleMove(ctx, conn, room, game, playerID, m.UCI)
			// refresh game for updated status/pgn
			game, _ = h.Games.ByID(ctx, gameID)
		case "resign":
			result := "0-1"
			status := "black_won"
			if isWhite {
				result = "0-1"
				status = "black_won"
			} else {
				result = "1-0"
				status = "white_won"
			}
			_ = h.Games.UpdateStatus(ctx, gameID, status, result, room.game.PGN())
			_ = h.Ratings.ApplyResult(ctx, game.WhiteID, game.BlackID, result)
			h.broadcastState(ctx, room)
			return
		}
	}
}

type inboundMsg struct {
	Type   string `json:"type"`
	GameID string `json:"gameId"`
	Token  string `json:"token"`
	UCI    string `json:"uci"`
}

type stateMsg struct {
	Type        string   `json:"type"`
	FEN         string   `json:"fen"`
	PGN         string   `json:"pgn"`
	Status      string   `json:"status"`
	LegalMoves  []string `json:"legalMoves"`
	SideToMove  string   `json:"sideToMove"`
}

type errorMsg struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type gameOverMsg struct {
	Type   string `json:"type"`
	Result string `json:"result"`
}

func (h *Hub) handleMove(ctx context.Context, conn *websocket.Conn, room *Room, game ports.Game, playerID uuid.UUID, uci string) {
	room.mu.Lock()
	defer room.mu.Unlock()

	g := room.game
	pos := g.PositionFEN()
	parsedPos, err := chess.ParseFEN(pos)
	if err != nil {
		writeError(ctx, conn, "internal error parsing FEN")
		return
	}

	// Verify it's this player's turn
	isWhiteTurn := parsedPos.SideToMove == chess.White
	isWhitePlayer := game.WhiteID != nil && *game.WhiteID == playerID
	if isWhiteTurn != isWhitePlayer {
		writeError(ctx, conn, "not your turn")
		return
	}

	mv, err := chess.ParseUCI(parsedPos, uci)
	if err != nil {
		writeError(ctx, conn, "invalid move: "+err.Error())
		return
	}

	// Compute SAN before applying move
	san := chess.MoveSAN(parsedPos, mv)

	ply := len(g.LegalMoves()) // approximate; use move count
	_ = ply
	// Better: use position history length
	moves, _ := h.Moves.ListByGame(ctx, room.gameID)
	nextPly := len(moves) + 1

	if err = g.Move(mv); err != nil {
		writeError(ctx, conn, "illegal move: "+err.Error())
		return
	}

	fenAfter := g.PositionFEN()
	_ = h.Moves.Append(ctx, room.gameID, nextPly, uci, san, fenAfter)

	status, result, pgn := gameOutcome(g)
	_ = h.Games.UpdateStatus(ctx, room.gameID, status, result, pgn)

	if result != "" {
		_ = h.Ratings.ApplyResult(ctx, game.WhiteID, game.BlackID, result)
	}

	h.broadcastStateLocked(ctx, room)
}

func (h *Hub) sendState(ctx context.Context, conn *websocket.Conn, room *Room) {
	room.mu.Lock()
	defer room.mu.Unlock()
	msg := buildStateMsg(room.game)
	_ = wsjson.Write(ctx, conn, msg)
}

func (h *Hub) broadcastState(ctx context.Context, room *Room) {
	room.mu.Lock()
	defer room.mu.Unlock()
	h.broadcastStateLocked(ctx, room)
}

func (h *Hub) broadcastStateLocked(ctx context.Context, room *Room) {
	msg := buildStateMsg(room.game)
	for _, c := range room.clients {
		_ = wsjson.Write(ctx, c, msg)
	}
	// If game is over, also send gameOver
	if msg.Status != "active" && msg.Status != "pending" && msg.Status != "Ongoing" && msg.Status != "Check" {
		result := statusToResult(room.game)
		if result != "" {
			over := gameOverMsg{Type: "gameOver", Result: result}
			for _, c := range room.clients {
				_ = wsjson.Write(ctx, c, over)
			}
		}
	}
}

func buildStateMsg(g *chess.Game) stateMsg {
	legalMoves := g.LegalMoves()
	uciMoves := make([]string, len(legalMoves))
	for i, m := range legalMoves {
		uciMoves[i] = m.UCI()
	}

	pos, _ := chess.ParseFEN(g.PositionFEN())
	side := "white"
	if pos != nil && pos.SideToMove == chess.Black {
		side = "black"
	}

	return stateMsg{
		Type:       "state",
		FEN:        g.PositionFEN(),
		PGN:        g.PGN(),
		Status:     g.Status().String(),
		LegalMoves: uciMoves,
		SideToMove: side,
	}
}

func gameOutcome(g *chess.Game) (status, result, pgn string) {
	pgn = g.PGN()
	switch g.Status() {
	case chess.Checkmate:
		pos, _ := chess.ParseFEN(g.PositionFEN())
		if pos != nil && pos.SideToMove == chess.Black {
			return "white_won", "1-0", pgn
		}
		return "black_won", "0-1", pgn
	case chess.Stalemate, chess.DrawFiftyMove, chess.DrawInsufficientMaterial, chess.DrawByRepetition:
		return "draw", "1/2-1/2", pgn
	default:
		return "active", "", pgn
	}
}

func statusToResult(g *chess.Game) string {
	switch g.Status() {
	case chess.Checkmate:
		pos, _ := chess.ParseFEN(g.PositionFEN())
		if pos != nil && pos.SideToMove == chess.Black {
			return "1-0"
		}
		return "0-1"
	case chess.Stalemate, chess.DrawFiftyMove, chess.DrawInsufficientMaterial, chess.DrawByRepetition:
		return "1/2-1/2"
	}
	return ""
}

func writeError(ctx context.Context, conn *websocket.Conn, msg string) {
	_ = wsjson.Write(ctx, conn, errorMsg{Type: "error", Message: msg})
}

// Room holds a game in progress and its connected clients.
type Room struct {
	mu      sync.Mutex
	game    *chess.Game
	gameID  uuid.UUID
	whiteID *uuid.UUID
	blackID *uuid.UUID
	clients []*websocket.Conn
}

func (rm *Room) addClient(c *websocket.Conn) {
	rm.clients = append(rm.clients, c)
}

func (rm *Room) removeClient(c *websocket.Conn) {
	for i, client := range rm.clients {
		if client == c {
			rm.clients = append(rm.clients[:i], rm.clients[i+1:]...)
			return
		}
	}
}

