package httpapi

import (
	"encoding/json"
	"net/http"

	jwtadapter "github.com/chessmaster-pro/chessmaster/internal/adapters/auth/jwt"
	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Deps holds all dependencies needed by the HTTP handlers.
type Deps struct {
	Players ports.PlayerRepo
	Games   ports.GameRepo
	Moves   ports.MoveRepo
	Ratings ports.RatingRepo
	Engine  ports.Engine
	Signer  *jwtadapter.Signer
	WS      http.Handler
}

// NewRouter constructs the chi router for the chess API.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth
	r.Post("/auth/register", d.handleRegister)
	r.Post("/auth/login", d.handleLogin)

	// Protected routes
	authMW := jwtadapter.Middleware(d.Signer)
	r.With(authMW).Get("/me", d.handleMe)

	// Games
	r.With(authMW).Post("/games", d.handleCreateGame)
	r.With(authMW).Post("/games/join", d.handleJoinGame)
	r.Get("/games/{id}", d.handleGetGame)
	r.Get("/games/{id}/moves", d.handleGetMoves)
	r.With(authMW).Post("/games/{id}/move", d.handlePostMove)

	// Players
	r.With(authMW).Get("/players/me/games", d.handleListMyGames)

	// Leaderboard
	r.Get("/leaderboard", d.handleLeaderboard)

	// WebSocket
	if d.WS != nil {
		r.Handle("/ws", d.WS)
	}

	return r
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
