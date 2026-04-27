package httpapi

import (
	"net/http"
	"strconv"
)

func (d *Deps) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	limitStr := r.URL.Query().Get("limit")

	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}

	var cityPtr *string
	if city != "" {
		cityPtr = &city
	}

	entries, err := d.Ratings.Leaderboard(r.Context(), cityPtr, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, entries)
}
