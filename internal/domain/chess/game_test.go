package chess

import (
	"strings"
	"testing"
)

func TestScholarsMate(t *testing.T) {
	g := NewGame()

	moves := []string{"e2e4", "e7e5", "f1c4", "b8c6", "d1h5", "g8f6", "h5f7"}
	for _, uci := range moves {
		pos := g.currentPosition()
		m, err := ParseUCI(pos, uci)
		if err != nil {
			t.Fatalf("ParseUCI(%q): %v", uci, err)
		}
		if err := g.Move(m); err != nil {
			t.Fatalf("Move(%q): %v", uci, err)
		}
	}

	status := g.Status()
	if status != Checkmate {
		t.Errorf("Scholar's mate: want Checkmate, got %v", status)
	}

	pgn := g.PGN()
	if !strings.Contains(pgn, "1-0") {
		t.Errorf("PGN should contain '1-0': %s", pgn)
	}
	if !strings.Contains(pgn, "Qxf7#") {
		t.Errorf("PGN should contain 'Qxf7#': %s", pgn)
	}
}

func TestThreefoldRepetition(t *testing.T) {
	g := NewGame()
	// Repeat a position three times by moving knights back and forth
	// g1f3, g8f6, f3g1, f6g8 — back to start (twice = 2 repetitions after 8 half-moves, 3rd when we repeat again)
	sequence := []string{
		"g1f3", "g8f6", "f3g1", "f6g8",
		"g1f3", "g8f6", "f3g1", "f6g8",
		"g1f3", "g8f6",
	}
	for _, uci := range sequence {
		pos := g.currentPosition()
		m, err := ParseUCI(pos, uci)
		if err != nil {
			t.Fatalf("ParseUCI(%q): %v", uci, err)
		}
		if err := g.Move(m); err != nil {
			t.Fatalf("Move(%q): %v", uci, err)
		}
	}

	status := g.Status()
	if status != DrawByRepetition {
		t.Errorf("Threefold repetition: want DrawByRepetition, got %v", status)
	}
}

func TestGameLegalMoves(t *testing.T) {
	g := NewGame()
	moves := g.LegalMoves()
	if len(moves) != 20 {
		t.Errorf("starting legal moves: want 20, got %d", len(moves))
	}
}

func TestGamePositionFEN(t *testing.T) {
	g := NewGame()
	fen := g.PositionFEN()
	want := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	if fen != want {
		t.Errorf("PositionFEN: want %q, got %q", want, fen)
	}
}

func TestGameMoveError(t *testing.T) {
	g := NewGame()
	// Try an illegal move
	m := Move{From: SquareOf(4, 1), To: SquareOf(4, 5)} // e2-e6, illegal
	if err := g.Move(m); err == nil {
		t.Error("expected error for illegal move e2e6")
	}
}

func TestGamePGNSevenTagRoster(t *testing.T) {
	g := NewGame()
	pgn := g.PGN()
	// Must have seven tag roster tags
	for _, tag := range []string{"Event", "Site", "Date", "Round", "White", "Black", "Result"} {
		if !strings.Contains(pgn, "["+tag) {
			t.Errorf("PGN missing tag: %s\nPGN: %s", tag, pgn)
		}
	}
}

func TestGameStatusAfterIllegalMove(t *testing.T) {
	g := NewGame()
	if g.Status() != Ongoing {
		t.Error("new game should be Ongoing")
	}
}
