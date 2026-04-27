package chess

import "testing"

func TestStatusFoolsMate(t *testing.T) {
	// Fool's mate: 1. f3 e5 2. g4 Qh4#
	pos := mustParseFEN(t, "rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3")
	s := GameStatus(pos, nil)
	if s != Checkmate {
		t.Errorf("Fool's mate: want Checkmate, got %v", s)
	}
}

func TestStatusStalemate(t *testing.T) {
	// Black king on a8, white queen on b6, white king on a6 — black to move, stalemate
	pos := mustParseFEN(t, "k7/8/KQ6/8/8/8/8/8 b - - 0 1")
	s := GameStatus(pos, nil)
	if s != Stalemate {
		t.Errorf("Stalemate position: want Stalemate, got %v", s)
	}
}

func TestStatusCheck(t *testing.T) {
	// King in check but has evasions
	pos := mustParseFEN(t, "4k3/8/8/8/8/8/8/4R2K b - - 0 1")
	s := GameStatus(pos, nil)
	if s != Check {
		t.Errorf("Check position: want Check, got %v", s)
	}
}

func TestStatusOngoing(t *testing.T) {
	pos := mustParseFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	s := GameStatus(pos, nil)
	if s != Ongoing {
		t.Errorf("Starting position: want Ongoing, got %v", s)
	}
}

func TestStatusKvK(t *testing.T) {
	pos := mustParseFEN(t, "8/8/8/8/8/8/8/4K2k w - - 0 1")
	s := GameStatus(pos, nil)
	if s != DrawInsufficientMaterial {
		t.Errorf("KvK: want DrawInsufficientMaterial, got %v", s)
	}
}

func TestStatusKBvK(t *testing.T) {
	pos := mustParseFEN(t, "8/8/8/8/8/8/8/2B1K2k w - - 0 1")
	s := GameStatus(pos, nil)
	if s != DrawInsufficientMaterial {
		t.Errorf("KBvK: want DrawInsufficientMaterial, got %v", s)
	}
}

func TestStatusKNvK(t *testing.T) {
	pos := mustParseFEN(t, "8/8/8/8/8/8/8/2N1K2k w - - 0 1")
	s := GameStatus(pos, nil)
	if s != DrawInsufficientMaterial {
		t.Errorf("KNvK: want DrawInsufficientMaterial, got %v", s)
	}
}

func TestStatusKBvKBSameColor(t *testing.T) {
	// Bishops on same color squares → insufficient material
	// White bishop on c1 (dark), black bishop on f8 (dark)
	pos := mustParseFEN(t, "5b2/8/8/8/8/8/8/2B1K2k w - - 0 1")
	s := GameStatus(pos, nil)
	if s != DrawInsufficientMaterial {
		t.Errorf("KBvKB same color: want DrawInsufficientMaterial, got %v", s)
	}
}

func TestStatusKBvKBDifferentColor(t *testing.T) {
	// Bishops on different color squares → not insufficient
	// White bishop on c1 (dark), black bishop on g8 (light)
	pos := mustParseFEN(t, "6b1/8/8/8/8/8/8/2B1K2k w - - 0 1")
	s := GameStatus(pos, nil)
	// Not insufficient material (bishops on different colors can force mate)
	if s == DrawInsufficientMaterial {
		t.Errorf("KBvKB different color: should NOT be DrawInsufficientMaterial")
	}
}

func TestStatusFiftyMoveRule(t *testing.T) {
	pos := mustParseFEN(t, "8/8/8/8/8/8/8/4K2k w - - 100 50")
	s := GameStatus(pos, nil)
	if s != DrawFiftyMove {
		t.Errorf("halfmove=100: want DrawFiftyMove, got %v", s)
	}
}

func TestStatusFiftyMoveNotYet(t *testing.T) {
	pos := mustParseFEN(t, "8/8/8/8/8/8/8/4K2k w - - 99 50")
	s := GameStatus(pos, nil)
	// KvK is insufficient material — but 99 is not yet 50 move draw
	// Insufficient material takes priority in our impl
	// Just verify it's not DrawFiftyMove
	if s == DrawFiftyMove {
		t.Errorf("halfmove=99: should not yet be DrawFiftyMove")
	}
}
