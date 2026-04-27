// Package jwt provides JWT token signing/parsing and bcrypt password helpers.
package jwt

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type contextKey struct{}

// Signer issues and validates HS256 JWT tokens.
type Signer struct {
	secret []byte
}

// NewSigner creates a Signer with the given HMAC secret.
func NewSigner(secret string) *Signer {
	return &Signer{secret: []byte(secret)}
}

// Issue creates a signed JWT for the given player ID, valid for 7 days.
func (s *Signer) Issue(playerID uuid.UUID) (string, error) {
	claims := gojwt.RegisteredClaims{
		Subject:   playerID.String(),
		ExpiresAt: gojwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  gojwt.NewNumericDate(time.Now()),
	}
	tok := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return tok.SignedString(s.secret)
}

// Parse validates a JWT and returns the player ID embedded in it.
func (s *Signer) Parse(token string) (uuid.UUID, error) {
	tok, err := gojwt.ParseWithClaims(token, &gojwt.RegisteredClaims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	claims, ok := tok.Claims.(*gojwt.RegisteredClaims)
	if !ok || !tok.Valid {
		return uuid.UUID{}, errors.New("invalid token")
	}
	return uuid.Parse(claims.Subject)
}

// Middleware extracts the Bearer token from Authorization header and sets the player ID on the context.
func Middleware(signer *Signer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hdr := r.Header.Get("Authorization")
			if !strings.HasPrefix(hdr, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			id, err := signer.Parse(strings.TrimPrefix(hdr, "Bearer "))
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), contextKey{}, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// PlayerIDFromContext retrieves the player UUID set by Middleware.
func PlayerIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(contextKey{}).(uuid.UUID)
	return id, ok
}

// HashPassword hashes a plaintext password with bcrypt.
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
