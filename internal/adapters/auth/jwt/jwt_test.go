package jwt_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtadapter "github.com/chessmaster-pro/chessmaster/internal/adapters/auth/jwt"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestSignerRoundTrip(t *testing.T) {
	t.Parallel()
	s := jwtadapter.NewSigner("test-secret")
	id := uuid.New()
	tok, err := s.Issue(id)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Parse(tok)
	if err != nil {
		t.Fatal(err)
	}
	if got != id {
		t.Fatalf("got %s want %s", got, id)
	}
}

func TestSignerExpiredToken(t *testing.T) {
	t.Parallel()
	s := jwtadapter.NewSigner("test-secret")
	claims := gojwt.RegisteredClaims{
		Subject:   uuid.New().String(),
		ExpiresAt: gojwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}
	tok := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString([]byte("test-secret"))
	_, err := s.Parse(signed)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestSignerMalformedToken(t *testing.T) {
	t.Parallel()
	s := jwtadapter.NewSigner("test-secret")
	_, err := s.Parse("not.a.token")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestMiddlewareMissingHeader(t *testing.T) {
	t.Parallel()
	s := jwtadapter.NewSigner("test-secret")
	mw := jwtadapter.Middleware(s)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
}

func TestMiddlewareMalformedHeader(t *testing.T) {
	t.Parallel()
	s := jwtadapter.NewSigner("test-secret")
	mw := jwtadapter.Middleware(s)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token-value")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rr.Code)
	}
}

func TestMiddlewareValidToken(t *testing.T) {
	t.Parallel()
	s := jwtadapter.NewSigner("test-secret")
	mw := jwtadapter.Middleware(s)
	id := uuid.New()
	tok, _ := s.Issue(id)

	var gotID uuid.UUID
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, _ = jwtadapter.PlayerIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	if gotID != id {
		t.Fatalf("wrong player id in context")
	}
}

func TestPasswordHashing(t *testing.T) {
	t.Parallel()
	hash, err := jwtadapter.HashPassword("secret123")
	if err != nil {
		t.Fatal(err)
	}
	if err = jwtadapter.CheckPassword(hash, "secret123"); err != nil {
		t.Fatal(err)
	}
	if err = jwtadapter.CheckPassword(hash, "wrong"); err == nil {
		t.Fatal("expected error for wrong password")
	}
}
