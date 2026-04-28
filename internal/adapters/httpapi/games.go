package httpapi

import (
	"encoding/json"
	"net/http"

	jwtadapter "github.com/chessmaster-pro/chessmaster/internal/adapters/auth/jwt"
	"github.com/chessmaster-pro/chessmaster/internal/domain/chess"
	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/chessmaster-pro/chessmaster/pkg/code"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type createGameRequest struct {
	Mode  string `json:"mode"`
	Color string `json:"color"`
}

type joinGameRequest struct {
	InviteCode string `json:"inviteCode"`
}

type moveRequest struct {
	UCI string `json:"uci"`
}

func (d *Deps) handleCreateGame(w http.ResponseWriter, r *http.Request) {
	playerID, ok := jwtadapter.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validModes := map[string]bool{"pvp": true, "ai_easy": true, "ai_medium": true, "ai_hard": true}
	if !validModes[req.Mode] {
		respondError(w, http.StatusBadRequest, "invalid mode")
		return
	}

	color := req.Color
	if color == "random" {
		color = "white" // simple default; could use crypto/rand
	}
	if color == "" {
		color = "white"
	}

	game := ports.Game{Mode: req.Mode, Status: "pending"}
	pid := playerID
	if color == "white" {
		game.WhiteID = &pid
	} else {
		game.BlackID = &pid
	}

	// For PvP games generate an invite code
	if req.Mode == "pvp" {
		c, err := code.Generate()
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal error")
			return
		}
		game.InviteCode = &c
	}

	// For AI games, create immediately as active
	if req.Mode != "pvp" {
		game.Status = "active"
	}

	created, err := d.Games.Create(r.Context(), game)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusCreated, gameResponse(created))
}

func (d *Deps) handleJoinGame(w http.ResponseWriter, r *http.Request) {
	playerID, ok := jwtadapter.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req joinGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.InviteCode == "" {
		respondError(w, http.StatusBadRequest, "inviteCode required")
		return
	}

	game, err := d.Games.ByInviteCode(r.Context(), req.InviteCode)
	if err != nil {
		respondError(w, http.StatusNotFound, "game not found")
		return
	}

	if game.Status != "pending" {
		respondError(w, http.StatusConflict, "game is not open")
		return
	}

	if err = d.Games.JoinAsBlack(r.Context(), game.ID, playerID); err != nil {
		respondError(w, http.StatusConflict, "cannot join game")
		return
	}

	updated, _ := d.Games.ByID(r.Context(), game.ID)
	respondJSON(w, http.StatusOK, gameResponse(updated))
}

func (d *Deps) handleGetGame(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid game id")
		return
	}

	game, err := d.Games.ByID(r.Context(), gameID)
	if err != nil {
		respondError(w, http.StatusNotFound, "game not found")
		return
	}

	moves, _ := d.Moves.ListByGame(r.Context(), gameID)
	g := replayGame(moves)
	fen, legal, side := boardSnapshot(g)
	respondJSON(w, http.StatusOK, map[string]any{
		"game":       gameResponse(game),
		"moves":      movesResponse(moves),
		"fen":        fen,
		"legalMoves": legal,
		"sideToMove": side,
	})
}

func (d *Deps) handleGetMoves(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid game id")
		return
	}

	moves, err := d.Moves.ListByGame(r.Context(), gameID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, movesResponse(moves))
}

func (d *Deps) handlePostMove(w http.ResponseWriter, r *http.Request) {
	playerID, ok := jwtadapter.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	gameID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid game id")
		return
	}

	var req moveRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil || req.UCI == "" {
		respondError(w, http.StatusBadRequest, "uci move required")
		return
	}

	game, err := d.Games.ByID(r.Context(), gameID)
	if err != nil {
		respondError(w, http.StatusNotFound, "game not found")
		return
	}

	if game.Status != "active" {
		respondError(w, http.StatusConflict, "game is not active")
		return
	}

	// Reconstruct game state from moves
	existingMoves, _ := d.Moves.ListByGame(r.Context(), gameID)
	g := chess.NewGame()
	for _, m := range existingMoves {
		pos, _ := chess.ParseFEN(g.PositionFEN())
		mv, _ := chess.ParseUCI(pos, m.UCI)
		_ = g.Move(mv)
	}

	pos, err := chess.ParseFEN(g.PositionFEN())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Validate turn
	isWhiteTurn := pos.SideToMove == chess.White
	isWhitePlayer := game.WhiteID != nil && *game.WhiteID == playerID
	isBlackPlayer := game.BlackID != nil && *game.BlackID == playerID

	if !isWhitePlayer && !isBlackPlayer {
		respondError(w, http.StatusForbidden, "you are not a participant")
		return
	}
	if isWhiteTurn != isWhitePlayer {
		respondError(w, http.StatusConflict, "not your turn")
		return
	}

	mv, err := chess.ParseUCI(pos, req.UCI)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid move")
		return
	}

	san := chess.MoveSAN(pos, mv)

	if err = g.Move(mv); err != nil {
		respondError(w, http.StatusBadRequest, "illegal move")
		return
	}

	fenAfter := g.PositionFEN()
	nextPly := len(existingMoves) + 1
	_ = d.Moves.Append(r.Context(), gameID, nextPly, req.UCI, san, fenAfter)

	status, result, pgn := gameOutcomeFromGame(g)
	_ = d.Games.UpdateStatus(r.Context(), gameID, status, result, pgn)

	if result != "" {
		_ = d.Ratings.ApplyResult(r.Context(), game.WhiteID, game.BlackID, result)
	}

	// If AI game and game still active, let the engine move
	if result == "" && d.Engine != nil && game.Mode != "pvp" {
		level := modeToLevel(game.Mode)
		uciMove, aiErr := d.Engine.BestMove(r.Context(), g.PositionFEN(), level)
		if aiErr == nil {
			aiPos, _ := chess.ParseFEN(g.PositionFEN())
			aiMv, _ := chess.ParseUCI(aiPos, uciMove)
			aiSAN := chess.MoveSAN(aiPos, aiMv)
			if g.Move(aiMv) == nil {
				aiPly := nextPly + 1
				_ = d.Moves.Append(r.Context(), gameID, aiPly, uciMove, aiSAN, g.PositionFEN())
				aiStatus, aiResult, aiPGN := gameOutcomeFromGame(g)
				_ = d.Games.UpdateStatus(r.Context(), gameID, aiStatus, aiResult, aiPGN)
				if aiResult != "" {
					_ = d.Ratings.ApplyResult(r.Context(), game.WhiteID, game.BlackID, aiResult)
				}
			}
		}
	}

	updated, _ := d.Games.ByID(r.Context(), gameID)
	allMoves, _ := d.Moves.ListByGame(r.Context(), gameID)
	fen, legal, side := boardSnapshot(g)
	respondJSON(w, http.StatusOK, map[string]any{
		"game":       gameResponse(updated),
		"moves":      movesResponse(allMoves),
		"fen":        fen,
		"legalMoves": legal,
		"sideToMove": side,
	})
}

func replayGame(moves []ports.Move) *chess.Game {
	g := chess.NewGame()
	for _, m := range moves {
		pos, err := chess.ParseFEN(g.PositionFEN())
		if err != nil {
			return g
		}
		mv, err := chess.ParseUCI(pos, m.UCI)
		if err != nil {
			return g
		}
		_ = g.Move(mv)
	}
	return g
}

func boardSnapshot(g *chess.Game) (fen string, legal []string, side string) {
	fen = g.PositionFEN()
	legal = []string{}
	for _, m := range g.LegalMoves() {
		legal = append(legal, m.UCI())
	}
	side = "white"
	if pos, err := chess.ParseFEN(fen); err == nil && pos.SideToMove == chess.Black {
		side = "black"
	}
	return
}

func (d *Deps) handleListMyGames(w http.ResponseWriter, r *http.Request) {
	playerID, ok := jwtadapter.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	games, err := d.Games.ListByPlayer(r.Context(), playerID, 20)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := make([]gameResp, len(games))
	for i, g := range games {
		resp[i] = gameResponse(g)
	}
	respondJSON(w, http.StatusOK, resp)
}

type gameResp struct {
	ID         string  `json:"id"`
	WhiteID    *string `json:"whiteId"`
	BlackID    *string `json:"blackId"`
	Mode       string  `json:"mode"`
	Status     string  `json:"status"`
	InviteCode *string `json:"inviteCode,omitempty"`
	PGN        string  `json:"pgn"`
	Result     *string `json:"result,omitempty"`
}

func gameResponse(g ports.Game) gameResp {
	resp := gameResp{
		ID:         g.ID.String(),
		Mode:       g.Mode,
		Status:     g.Status,
		InviteCode: g.InviteCode,
		PGN:        g.PGN,
		Result:     g.Result,
	}
	if g.WhiteID != nil {
		s := g.WhiteID.String()
		resp.WhiteID = &s
	}
	if g.BlackID != nil {
		s := g.BlackID.String()
		resp.BlackID = &s
	}
	return resp
}

type moveResp struct {
	Ply      int    `json:"ply"`
	UCI      string `json:"uci"`
	SAN      string `json:"san"`
	FENAfter string `json:"fenAfter"`
}

func movesResponse(moves []ports.Move) []moveResp {
	resp := make([]moveResp, len(moves))
	for i, m := range moves {
		resp[i] = moveResp{
			Ply:      m.Ply,
			UCI:      m.UCI,
			SAN:      m.SAN,
			FENAfter: m.FENAfter,
		}
	}
	return resp
}

func gameOutcomeFromGame(g *chess.Game) (status, result, pgn string) {
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

func modeToLevel(mode string) int {
	switch mode {
	case "ai_easy":
		return 1
	case "ai_medium":
		return 2
	case "ai_hard":
		return 3
	}
	return 2
}
