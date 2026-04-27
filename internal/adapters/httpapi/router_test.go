package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	jwtadapter "github.com/chessmaster-pro/chessmaster/internal/adapters/auth/jwt"
	"github.com/chessmaster-pro/chessmaster/internal/adapters/memrepo"
)

func newTestDeps() Deps {
	players := memrepo.NewPlayers()
	return Deps{
		Players: players,
		Games:   memrepo.NewGames(),
		Moves:   memrepo.NewMoves(),
		Ratings: memrepo.NewRatings(players),
		Signer:  jwtadapter.NewSigner("test-secret"),
	}
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"healthz ok", http.MethodGet, "/healthz", http.StatusOK, `"status":"ok"`},
		{"unknown 404", http.MethodGet, "/nope", http.StatusNotFound, ""},
		{"cors preflight", http.MethodOptions, "/healthz", http.StatusNoContent, ""},
	}

	srv := NewRouter(newTestDeps())
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()
			srv.ServeHTTP(rr, req)
			if rr.Code != tc.wantStatus {
				t.Fatalf("status: got %d want %d", rr.Code, tc.wantStatus)
			}
			if tc.wantBody != "" && !strings.Contains(rr.Body.String(), tc.wantBody) {
				t.Fatalf("body: got %q want substring %q", rr.Body.String(), tc.wantBody)
			}
		})
	}
}
