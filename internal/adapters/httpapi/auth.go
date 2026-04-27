package httpapi

import (
	"encoding/json"
	"net/http"

	jwtadapter "github.com/chessmaster-pro/chessmaster/internal/adapters/auth/jwt"
	"github.com/chessmaster-pro/chessmaster/internal/ports"
)

type registerRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	City     string `json:"city"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

func (d *Deps) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email, username and password are required")
		return
	}

	hash, err := jwtadapter.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	p, err := d.Players.Create(r.Context(), ports.Player{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hash,
		City:         req.City,
	})
	if err != nil {
		respondError(w, http.StatusConflict, "email already registered")
		return
	}

	tok, err := d.Signer.Issue(p.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusCreated, tokenResponse{Token: tok})
}

func (d *Deps) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := d.Players.ByEmail(r.Context(), req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err = jwtadapter.CheckPassword(p.PasswordHash, req.Password); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	tok, err := d.Signer.Issue(p.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondJSON(w, http.StatusOK, tokenResponse{Token: tok})
}

func (d *Deps) handleMe(w http.ResponseWriter, r *http.Request) {
	playerID, ok := jwtadapter.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	p, err := d.Players.ByID(r.Context(), playerID)
	if err != nil {
		respondError(w, http.StatusNotFound, "player not found")
		return
	}
	respondJSON(w, http.StatusOK, playerResponse(p))
}

type playerResp struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	City      string `json:"city"`
	IsPro     bool   `json:"isPro"`
	CreatedAt string `json:"createdAt"`
}

func playerResponse(p ports.Player) playerResp {
	return playerResp{
		ID:        p.ID.String(),
		Email:     p.Email,
		Username:  p.Username,
		City:      p.City,
		IsPro:     p.IsPro,
		CreatedAt: p.CreatedAt.String(),
	}
}
