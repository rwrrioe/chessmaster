package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chessmaster-pro/chessmaster/internal/ports"
)

// geminiEnvelope wraps an analysis JSON string inside the Gemini REST response shape.
func geminiEnvelope(t *testing.T, text string) []byte {
	t.Helper()
	env := map[string]any{
		"candidates": []any{
			map[string]any{
				"content": map[string]any{
					"parts": []any{
						map[string]any{"text": text},
					},
				},
			},
		},
	}
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("geminiEnvelope: %v", err)
	}
	return b
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New("test-key").withBaseURL(srv.URL)
}

// ---- table-driven tests for straightforward cases ----

func TestAnalyze_TableDriven(t *testing.T) {
	t.Parallel()

	wantAnalysis := ports.Analysis{
		Summary: "A solid game with one tactical miss.",
		Mistakes: []ports.Mistake{
			{Ply: 14, Move: "d4d5", Severity: ports.SevBlunder, Better: "f2f4", Comment: "Drops the pawn."},
		},
	}
	validJSON, _ := json.Marshal(wantAnalysis)

	fencedJSON := "```json\n" + string(validJSON) + "\n```"

	tests := []struct {
		name        string
		serverText  string // text placed inside the Gemini envelope
		wantErr     string // substring expected in error; empty means no error
		wantSummary string
	}{
		{
			name:        "happy path",
			serverText:  string(validJSON),
			wantSummary: wantAnalysis.Summary,
		},
		{
			name:        "json inside code fences",
			serverText:  fencedJSON,
			wantSummary: wantAnalysis.Summary,
		},
		{
			name:       "malformed JSON in text",
			serverText: `{"summary": "oops",,}`,
			wantErr:    "parse analysis JSON",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(geminiEnvelope(t, tc.serverText)) //nolint:errcheck
			})

			got, err := c.Analyze(context.Background(), "1. e4 e5")

			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Summary != tc.wantSummary {
				t.Fatalf("summary: got %q want %q", got.Summary, tc.wantSummary)
			}
		})
	}
}

// ---- separate tests where setup differs significantly ----

func TestAnalyze_Non2xx(t *testing.T) {
	t.Parallel()
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid key"}`)) //nolint:errcheck
	})

	_, err := c.Analyze(context.Background(), "1. e4 e5")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("error %q does not mention 401", err.Error())
	}
}

func TestAnalyze_EmptyCandidates(t *testing.T) {
	t.Parallel()
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"candidates":[]}`)) //nolint:errcheck
	})

	_, err := c.Analyze(context.Background(), "1. e4 e5")
	if err == nil {
		t.Fatal("expected error for empty candidates, got nil")
	}
	if !strings.Contains(err.Error(), "no candidates") {
		t.Fatalf("error %q should mention 'no candidates'", err.Error())
	}
}

func TestAnalyze_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Slow handler that never responds within time
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block until the request context is done
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)

	c := New("test-key").withBaseURL(srv.URL)
	// Give the HTTP client a generous timeout so the test controls cancellation
	c.http = &http.Client{Timeout: 10 * time.Second}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately before the call

	_, err := c.Analyze(ctx, "1. e4 e5")
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	// The error should wrap context.Canceled
	if !strings.Contains(err.Error(), "context canceled") && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("expected ctx error, got: %v", err)
	}
}

func TestAnalyze_SeverityNormalisation(t *testing.T) {
	t.Parallel()

	analysisWithWeird := ports.Analysis{
		Summary: "test",
		Mistakes: []ports.Mistake{
			{Ply: 1, Move: "e2e4", Severity: "weird", Comment: "odd"},
			{Ply: 2, Move: "d2d4", Severity: ports.SevBlunder, Comment: "real blunder"},
		},
	}
	raw, _ := json.Marshal(analysisWithWeird)

	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(geminiEnvelope(t, string(raw))) //nolint:errcheck
	})

	got, err := c.Analyze(context.Background(), "1. e4 e5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Mistakes[0].Severity != ports.SevMistake {
		t.Fatalf("expected 'weird' normalised to %q, got %q", ports.SevMistake, got.Mistakes[0].Severity)
	}
	if got.Mistakes[1].Severity != ports.SevBlunder {
		t.Fatalf("valid severity should be unchanged, got %q", got.Mistakes[1].Severity)
	}
}
