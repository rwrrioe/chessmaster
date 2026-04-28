package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	jwtadapter "github.com/chessmaster-pro/chessmaster/internal/adapters/auth/jwt"
	"github.com/chessmaster-pro/chessmaster/internal/adapters/gemini"
	"github.com/chessmaster-pro/chessmaster/internal/adapters/httpapi"
	"github.com/chessmaster-pro/chessmaster/internal/adapters/memrepo"
	pgadapter "github.com/chessmaster-pro/chessmaster/internal/adapters/postgres"
	"github.com/chessmaster-pro/chessmaster/internal/adapters/stockfish"
	"github.com/chessmaster-pro/chessmaster/internal/adapters/wsroom"
	"github.com/chessmaster-pro/chessmaster/internal/ports"
)

func main() {
	port := envOr("PORT", "8080")
	jwtSecret := envOr("JWT_SECRET", "dev-secret-change-me")

	signer := jwtadapter.NewSigner(jwtSecret)

	var (
		players ports.PlayerRepo
		games   ports.GameRepo
		moves   ports.MoveRepo
		ratings ports.RatingRepo
	)

	if pgURL := os.Getenv("POSTGRES_URL"); pgURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pool, err := pgadapter.New(ctx, pgURL)
		if err != nil {
			log.Fatalf("postgres: %v", err)
		}
		if err = pgadapter.EnsureSchema(ctx, pool); err != nil {
			log.Fatalf("schema: %v", err)
		}
		players = pgadapter.NewPlayers(pool)
		games = pgadapter.NewGames(pool)
		moves = pgadapter.NewMoves(pool)
		ratings = pgadapter.NewRatings(pool)
		log.Println("using postgres repos")
	} else {
		log.Println("running with in-memory repos (dev)")
		memPlayers := memrepo.NewPlayers()
		players = memPlayers
		games = memrepo.NewGames()
		moves = memrepo.NewMoves()
		ratings = memrepo.NewRatings(memPlayers)
	}

	var engine ports.Engine
	sfPath := os.Getenv("STOCKFISH_PATH")
	if sfPath == "" {
		sfPath = "stockfish"
	}
	eng := &stockfish.Engine{Path: sfPath}
	// Attempt to verify binary exists; fall back to nil (no AI)
	if _, err := os.Stat(sfPath); err == nil {
		engine = eng
		log.Printf("stockfish engine at %s", sfPath)
	} else {
		log.Println("stockfish not found; AI moves disabled")
	}

	hub := wsroom.NewHub(games, moves, ratings, signer)

	// Wire AI coach. Set GEMINI_API_KEY to enable; omit to run without coaching.
	var coach ports.Coach
	if geminiKey := os.Getenv("GEMINI_API_KEY"); geminiKey != "" {
		c := gemini.New(geminiKey)
		if model := os.Getenv("GEMINI_MODEL"); model != "" {
			c = c.WithModel(model)
			log.Printf("gemini coach enabled (model=%s)", model)
		} else {
			log.Println("gemini coach enabled (default model)")
		}
		coach = c
	} else {
		log.Println("GEMINI_API_KEY not set; AI coach disabled")
	}

	deps := httpapi.Deps{
		Players: players,
		Games:   games,
		Moves:   moves,
		Ratings: ratings,
		Engine:  engine,
		Coach:   coach,
		Signer:  signer,
		WS:      hub,
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewRouter(deps),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("api listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
