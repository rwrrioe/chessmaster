package memrepo_test

import (
	"context"
	"testing"

	"github.com/chessmaster-pro/chessmaster/internal/adapters/memrepo"
	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/google/uuid"
)

func TestPlayers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := memrepo.NewPlayers()

	p, err := repo.Create(ctx, ports.Player{Email: "a@b.com", Username: "alice", PasswordHash: "hash"})
	if err != nil {
		t.Fatal(err)
	}
	if p.ID == (uuid.UUID{}) {
		t.Fatal("id not generated")
	}

	got, err := repo.ByEmail(ctx, "a@b.com")
	if err != nil {
		t.Fatal(err)
	}
	if got.Username != "alice" {
		t.Fatalf("got username %q", got.Username)
	}

	got2, err := repo.ByID(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got2.Email != "a@b.com" {
		t.Fatal("ByID wrong email")
	}

	// conflict on duplicate email
	_, err = repo.Create(ctx, ports.Player{Email: "a@b.com", Username: "other"})
	if err != memrepo.ErrConflict {
		t.Fatalf("want ErrConflict, got %v", err)
	}

	// UpdateCity
	if err = repo.UpdateCity(ctx, p.ID, "Almaty"); err != nil {
		t.Fatal(err)
	}
	got3, _ := repo.ByID(ctx, p.ID)
	if got3.City != "Almaty" {
		t.Fatal("city not updated")
	}

	// SetPro
	if err = repo.SetPro(ctx, p.ID, true); err != nil {
		t.Fatal(err)
	}
	got4, _ := repo.ByID(ctx, p.ID)
	if !got4.IsPro {
		t.Fatal("isPro not set")
	}
}

func TestGames(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := memrepo.NewGames()

	whiteID := uuid.New()
	code := "ABCD1234"
	g, err := repo.Create(ctx, ports.Game{
		WhiteID:    &whiteID,
		Mode:       "pvp",
		Status:     "pending",
		InviteCode: &code,
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := repo.ByID(ctx, g.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != "pvp" {
		t.Fatal("mode mismatch")
	}

	got2, err := repo.ByInviteCode(ctx, code)
	if err != nil {
		t.Fatal(err)
	}
	if got2.ID != g.ID {
		t.Fatal("invite code lookup wrong game")
	}

	blackID := uuid.New()
	if err = repo.JoinAsBlack(ctx, g.ID, blackID); err != nil {
		t.Fatal(err)
	}
	got3, _ := repo.ByID(ctx, g.ID)
	if got3.Status != "active" {
		t.Fatalf("status after join: %s", got3.Status)
	}

	if err = repo.UpdateStatus(ctx, g.ID, "white_won", "1-0", "pgn"); err != nil {
		t.Fatal(err)
	}
	got4, _ := repo.ByID(ctx, g.ID)
	if *got4.Result != "1-0" {
		t.Fatal("result not set")
	}

	games, err := repo.ListByPlayer(ctx, whiteID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(games))
	}
}

func TestMoves(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := memrepo.NewMoves()

	gid := uuid.New()
	if err := repo.Append(ctx, gid, 1, "e2e4", "e4", "fen1"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Append(ctx, gid, 2, "e7e5", "e5", "fen2"); err != nil {
		t.Fatal(err)
	}

	moves, err := repo.ListByGame(ctx, gid)
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 2 {
		t.Fatalf("expected 2 moves, got %d", len(moves))
	}
	if moves[0].UCI != "e2e4" {
		t.Fatal("wrong first move")
	}
}

func TestRatings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	players := memrepo.NewPlayers()
	repo := memrepo.NewRatings(players)

	wID := uuid.New()
	bID := uuid.New()

	// auto-create
	r, err := repo.Get(ctx, wID)
	if err != nil {
		t.Fatal(err)
	}
	if r.Elo != 1200 {
		t.Fatalf("default elo: %d", r.Elo)
	}

	// apply white win
	if err = repo.ApplyResult(ctx, &wID, &bID, "1-0"); err != nil {
		t.Fatal(err)
	}
	rW, _ := repo.Get(ctx, wID)
	rB, _ := repo.Get(ctx, bID)
	if rW.Elo <= 1200 {
		t.Fatalf("white elo should have risen: %d", rW.Elo)
	}
	if rB.Elo >= 1200 {
		t.Fatalf("black elo should have fallen: %d", rB.Elo)
	}

	// AI game — no change
	if err = repo.ApplyResult(ctx, &wID, nil, "1-0"); err != nil {
		t.Fatal(err)
	}
}

func TestLeaderboard(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	players := memrepo.NewPlayers()
	repo := memrepo.NewRatings(players)

	p1, _ := players.Create(ctx, ports.Player{Email: "c@d.com", Username: "charlie", City: "Almaty"})
	p2, _ := players.Create(ctx, ports.Player{Email: "e@f.com", Username: "eve", City: "Nur-Sultan"})

	// give charlie a win
	repo.ApplyResult(ctx, &p1.ID, &p2.ID, "1-0") //nolint

	city := "Almaty"
	entries, err := repo.Leaderboard(ctx, &city, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry for Almaty, got %d", len(entries))
	}
	if entries[0].Username != "charlie" {
		t.Fatalf("wrong username: %s", entries[0].Username)
	}

	all, _ := repo.Leaderboard(ctx, nil, 10)
	if len(all) != 2 {
		t.Fatalf("expected 2 entries total, got %d", len(all))
	}
}
