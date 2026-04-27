package httpapi

import (
	"context"
	"net/http"
	"time"

	jwtadapter "github.com/chessmaster-pro/chessmaster/internal/adapters/auth/jwt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// handleCoach analyses a game via the AI coach.
// Any authenticated user may request coaching for any game.
func (d *Deps) handleCoach(w http.ResponseWriter, r *http.Request) {
	_, ok := jwtadapter.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if d.Coach == nil {
		respondError(w, http.StatusServiceUnavailable, "ai coach not configured")
		return
	}

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

	if game.PGN == "" {
		respondError(w, http.StatusConflict, "no moves yet")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	analysis, err := d.Coach.Analyze(ctx, game.PGN)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "coach error")
		return
	}

	respondJSON(w, http.StatusOK, analysis)
}
