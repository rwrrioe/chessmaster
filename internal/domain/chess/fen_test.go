package chess

import (
	"strings"
	"testing"
)

func TestFENRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		fen  string
	}{
		{
			name: "starting position",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		},
		{
			name: "kiwipete midgame",
			fen:  "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		},
		{
			name: "en passant target",
			fen:  "rnbqkbnr/pppp1ppp/8/4pP2/8/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2",
		},
		{
			name: "partial castle rights - only white kingside",
			fen:  "r3k2r/8/8/8/8/8/8/R3K2R w K - 4 10",
		},
		{
			name: "no castle rights",
			fen:  "8/8/8/8/8/8/8/4K2k w - - 0 1",
		},
		{
			name: "black to move with en passant",
			fen:  "rnbqkbnr/pppp1ppp/8/8/4pP2/8/PPPPP1PP/RNBQKBNR b KQkq f3 0 2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pos, err := ParseFEN(tc.fen)
			if err != nil {
				t.Fatalf("ParseFEN(%q) error: %v", tc.fen, err)
			}
			got := pos.FEN()
			if got != tc.fen {
				t.Errorf("FEN round-trip failed:\n  want %q\n  got  %q", tc.fen, got)
			}
		})
	}
}

func TestParseFENFields(t *testing.T) {
	pos, err := ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if pos.SideToMove != White {
		t.Errorf("SideToMove: want White got %v", pos.SideToMove)
	}
	if !pos.Castle.Has(CastleWK) || !pos.Castle.Has(CastleWQ) || !pos.Castle.Has(CastleBK) || !pos.Castle.Has(CastleBQ) {
		t.Errorf("Castle rights should be full: got %v", pos.Castle)
	}
	if pos.EnPassantTarget != NoSquare {
		t.Errorf("EnPassantTarget should be NoSquare: got %v", pos.EnPassantTarget)
	}
	if pos.HalfmoveClock != 0 {
		t.Errorf("HalfmoveClock: want 0 got %d", pos.HalfmoveClock)
	}
	if pos.FullmoveNumber != 1 {
		t.Errorf("FullmoveNumber: want 1 got %d", pos.FullmoveNumber)
	}
	// Spot-check specific squares
	if p := pos.Board.At(SquareOf(4, 0)); p != (Piece{White, King}) {
		t.Errorf("e1 should be White King, got %+v", p)
	}
	if p := pos.Board.At(SquareOf(4, 7)); p != (Piece{Black, King}) {
		t.Errorf("e8 should be Black King, got %+v", p)
	}
	if p := pos.Board.At(SquareOf(0, 1)); p != (Piece{White, Pawn}) {
		t.Errorf("a2 should be White Pawn, got %+v", p)
	}
}

func TestParseFENEnPassantSquare(t *testing.T) {
	pos, err := ParseFEN("rnbqkbnr/pppp1ppp/8/4pP2/8/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2")
	if err != nil {
		t.Fatal(err)
	}
	want := SquareOf(4, 5) // e6
	if pos.EnPassantTarget != want {
		t.Errorf("EnPassantTarget: want %v got %v", want, pos.EnPassantTarget)
	}
}

func TestParseFENErrors(t *testing.T) {
	cases := []struct {
		name string
		fen  string
		want string // substring of expected error
	}{
		{"empty string", "", "fields"},
		{"too few fields", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -", "fields"},
		{"bad rank count", "rnbqkbnr/pppppppp/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "ranks"},
		{"bad side to move", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR x KQkq - 0 1", "side"},
		{"bad castle", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w XQkq - 0 1", "castle"},
		{"bad ep square", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq z9 0 1", "en passant"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseFEN(tc.fen)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.want)
			}
			if !strings.Contains(strings.ToLower(err.Error()), tc.want) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}
