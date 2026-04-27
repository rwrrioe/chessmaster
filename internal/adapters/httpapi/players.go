package httpapi

import (
	"net/http"

	jwtadapter "github.com/chessmaster-pro/chessmaster/internal/adapters/auth/jwt"
)

// handleUpgrade sets isPro = true for the authenticated player and returns
// the updated player profile.
func (d *Deps) handleUpgrade(w http.ResponseWriter, r *http.Request) {
	playerID, ok := jwtadapter.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := d.Players.SetPro(r.Context(), playerID, true); err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	p, err := d.Players.ByID(r.Context(), playerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, playerResponse(p))
}
