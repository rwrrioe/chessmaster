package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// registerAndLogin is a test helper that creates a user and returns their JWT.
func registerAndLogin(t *testing.T, srv http.Handler, email, username, password string) string {
	t.Helper()

	body, _ := json.Marshal(map[string]string{
		"email": email, "username": username, "password": password, "city": "Almaty",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("register: got %d, body: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp) //nolint
	return resp["token"]
}

func TestRegisterLogin(t *testing.T) {
	t.Parallel()
	srv := NewRouter(newTestDeps())

	tok := registerAndLogin(t, srv, "alice@chess.com", "alice", "pass123")
	if tok == "" {
		t.Fatal("no token on register")
	}

	// Login with correct credentials
	body, _ := json.Marshal(map[string]string{"email": "alice@chess.com", "password": "pass123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("login: got %d", rr.Code)
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp) //nolint
	if resp["token"] == "" {
		t.Fatal("no token on login")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	t.Parallel()
	srv := NewRouter(newTestDeps())
	registerAndLogin(t, srv, "bob@chess.com", "bob", "pass123")

	body, _ := json.Marshal(map[string]string{"email": "bob@chess.com", "password": "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthRequiredEndpointWithoutToken(t *testing.T) {
	t.Parallel()
	srv := NewRouter(newTestDeps())

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/me"},
		{http.MethodPost, "/games"},
		{http.MethodGet, "/players/me/games"},
	}
	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rr := httptest.NewRecorder()
		srv.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s: expected 401, got %d", tc.method, tc.path, rr.Code)
		}
	}
}

func TestCreateAndJoinPvP(t *testing.T) {
	t.Parallel()
	srv := NewRouter(newTestDeps())

	tok1 := registerAndLogin(t, srv, "white@chess.com", "whitePlayer", "pass")
	tok2 := registerAndLogin(t, srv, "black@chess.com", "blackPlayer", "pass")

	// Create pvp game as white
	body, _ := json.Marshal(map[string]string{"mode": "pvp", "color": "white"})
	req := httptest.NewRequest(http.MethodPost, "/games", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok1)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create game: got %d body: %s", rr.Code, rr.Body.String())
	}

	var gameResp map[string]any
	json.NewDecoder(rr.Body).Decode(&gameResp) //nolint
	inviteCode, _ := gameResp["inviteCode"].(string)
	if inviteCode == "" {
		t.Fatalf("no inviteCode in response: %v", gameResp)
	}

	// Join with black player
	body2, _ := json.Marshal(map[string]string{"inviteCode": inviteCode})
	req2 := httptest.NewRequest(http.MethodPost, "/games/join", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+tok2)
	rr2 := httptest.NewRecorder()
	srv.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("join game: got %d body: %s", rr2.Code, rr2.Body.String())
	}

	var joinResp map[string]any
	json.NewDecoder(rr2.Body).Decode(&joinResp) //nolint
	if joinResp["status"] != "active" {
		t.Fatalf("expected active game, got: %v", joinResp["status"])
	}
}

func TestGetGameAndMoves(t *testing.T) {
	t.Parallel()
	srv := NewRouter(newTestDeps())

	tok := registerAndLogin(t, srv, "mover@chess.com", "mover", "pass")

	// Create AI game to easily test moves
	body, _ := json.Marshal(map[string]string{"mode": "ai_easy", "color": "white"})
	req := httptest.NewRequest(http.MethodPost, "/games", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create game: got %d", rr.Code)
	}
	var gr map[string]any
	json.NewDecoder(rr.Body).Decode(&gr) //nolint
	gameID := gr["id"].(string)

	// Get game
	req2 := httptest.NewRequest(http.MethodGet, "/games/"+gameID, nil)
	rr2 := httptest.NewRecorder()
	srv.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("get game: got %d", rr2.Code)
	}

	// Get moves (should be empty)
	req3 := httptest.NewRequest(http.MethodGet, "/games/"+gameID+"/moves", nil)
	rr3 := httptest.NewRecorder()
	srv.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusOK {
		t.Fatalf("get moves: got %d", rr3.Code)
	}
}

func TestLeaderboard(t *testing.T) {
	t.Parallel()
	srv := NewRouter(newTestDeps())

	// No filter
	req := httptest.NewRequest(http.MethodGet, "/leaderboard", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("leaderboard: got %d", rr.Code)
	}

	// With city filter
	req2 := httptest.NewRequest(http.MethodGet, "/leaderboard?city=Almaty&limit=10", nil)
	rr2 := httptest.NewRecorder()
	srv.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("leaderboard city filter: got %d", rr2.Code)
	}
}

func TestListMyGames(t *testing.T) {
	t.Parallel()
	srv := NewRouter(newTestDeps())

	tok := registerAndLogin(t, srv, "gamer@chess.com", "gamer", "pass")

	// Create a game
	body, _ := json.Marshal(map[string]string{"mode": "ai_easy", "color": "white"})
	req := httptest.NewRequest(http.MethodPost, "/games", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	httptest.NewRecorder()
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	// List games
	req2 := httptest.NewRequest(http.MethodGet, "/players/me/games", nil)
	req2.Header.Set("Authorization", "Bearer "+tok)
	rr2 := httptest.NewRecorder()
	srv.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("list games: got %d", rr2.Code)
	}
	var games []any
	json.NewDecoder(rr2.Body).Decode(&games) //nolint
	if len(games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(games))
	}
}
