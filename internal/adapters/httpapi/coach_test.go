package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chessmaster-pro/chessmaster/internal/ports"
	"github.com/google/uuid"
)

// fakeCoach is a test double that returns a canned Analysis or error.
type fakeCoach struct {
	a   ports.Analysis
	err error
}

func (f *fakeCoach) Analyze(_ context.Context, _ string) (ports.Analysis, error) {
	return f.a, f.err
}

// createGame helper creates an ai_easy game and returns its ID string.
func createGame(t *testing.T, srv http.Handler, tok string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"mode": "ai_easy", "color": "white"})
	req := httptest.NewRequest(http.MethodPost, "/games", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("createGame: got %d body: %s", rr.Code, rr.Body.String())
	}
	var gr map[string]any
	json.NewDecoder(rr.Body).Decode(&gr) //nolint
	return gr["id"].(string)
}

func TestCoach_Unauthorized(t *testing.T) {
	t.Parallel()
	srv := NewRouter(newTestDeps())

	req := httptest.NewRequest(http.MethodPost, "/games/00000000-0000-0000-0000-000000000001/coach", nil)
	// no Authorization header
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestCoach_NilCoach_503(t *testing.T) {
	t.Parallel()
	deps := newTestDeps() // Coach field is nil
	srv := NewRouter(deps)

	tok := registerAndLogin(t, srv, "coach503@chess.com", "coach503", "pass")
	gameID := createGame(t, srv, tok)

	req := httptest.NewRequest(http.MethodPost, "/games/"+gameID+"/coach", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "not configured") {
		t.Fatalf("body should mention 'not configured', got: %s", rr.Body.String())
	}
}

func TestCoach_GameNotFound_404(t *testing.T) {
	t.Parallel()
	deps := newTestDeps()
	deps.Coach = &fakeCoach{}
	srv := NewRouter(deps)

	tok := registerAndLogin(t, srv, "coach404@chess.com", "coach404", "pass")

	req := httptest.NewRequest(http.MethodPost, "/games/00000000-0000-0000-0000-000000000099/coach", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body: %s", rr.Code, rr.Body.String())
	}
}

func TestCoach_EmptyPGN_409(t *testing.T) {
	t.Parallel()
	deps := newTestDeps()
	deps.Coach = &fakeCoach{}
	srv := NewRouter(deps)

	tok := registerAndLogin(t, srv, "coach409@chess.com", "coach409", "pass")
	// createGame creates a fresh game with no moves → PGN is empty
	gameID := createGame(t, srv, tok)

	req := httptest.NewRequest(http.MethodPost, "/games/"+gameID+"/coach", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "no moves") {
		t.Fatalf("body should mention 'no moves', got: %s", rr.Body.String())
	}
}

func TestCoach_Success_200(t *testing.T) {
	t.Parallel()

	canned := ports.Analysis{
		Summary: "Well played game.",
		Mistakes: []ports.Mistake{
			{Ply: 5, Move: "e2e4", Severity: ports.SevInaccuracy, Comment: "Slight inaccuracy."},
		},
	}

	deps := newTestDeps()
	deps.Coach = &fakeCoach{a: canned}
	srv := NewRouter(deps)

	tok := registerAndLogin(t, srv, "coachOK@chess.com", "coachOK", "pass")
	gameID := createGame(t, srv, tok)

	// Inject a non-empty PGN via UpdateStatus so the handler sees game.PGN != "".
	id, err := uuid.Parse(gameID)
	if err != nil {
		t.Fatalf("parse game uuid: %v", err)
	}
	if err = deps.Games.UpdateStatus(context.Background(), id, "active", "", "1. e4 e5 2. Nf3"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/games/"+gameID+"/coach", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body: %s", rr.Code, rr.Body.String())
	}

	var got ports.Analysis
	if err = json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Summary != canned.Summary {
		t.Fatalf("summary: got %q want %q", got.Summary, canned.Summary)
	}
	if len(got.Mistakes) != 1 || got.Mistakes[0].Ply != 5 {
		t.Fatalf("unexpected mistakes: %+v", got.Mistakes)
	}
}
