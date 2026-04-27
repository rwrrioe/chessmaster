package chess

import (
	"testing"
)

func TestUCIRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		fen  string
		uci  string
	}{
		{
			name: "simple pawn move",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			uci:  "e2e4",
		},
		{
			name: "knight move",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			uci:  "g1f3",
		},
		{
			name: "queen promotion",
			fen:  "8/P7/8/8/8/8/8/4K2k w - - 0 1",
			uci:  "a7a8q",
		},
		{
			name: "knight promotion",
			fen:  "8/P7/8/8/8/8/8/4K2k w - - 0 1",
			uci:  "a7a8n",
		},
		{
			name: "kingside castle",
			fen:  "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			uci:  "e1g1",
		},
		{
			name: "queenside castle",
			fen:  "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			uci:  "e1c1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pos := mustParseFEN(t, tc.fen)
			m, err := ParseUCI(pos, tc.uci)
			if err != nil {
				t.Fatalf("ParseUCI(%q): %v", tc.uci, err)
			}
			got := m.UCI()
			if got != tc.uci {
				t.Errorf("UCI round-trip: want %q got %q", tc.uci, got)
			}
		})
	}
}

func TestParseUCIErrors(t *testing.T) {
	pos := mustParseFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	cases := []string{"", "e2", "e9e4", "e2e9", "e2e4x"}
	for _, uci := range cases {
		if _, err := ParseUCI(pos, uci); err == nil {
			t.Errorf("ParseUCI(%q) expected error, got nil", uci)
		}
	}
}

func TestMoveSAN(t *testing.T) {
	cases := []struct {
		name string
		fen  string
		uci  string
		want string
	}{
		{
			name: "simple knight move",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			uci:  "g1f3",
			want: "Nf3",
		},
		{
			name: "pawn push",
			fen:  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			uci:  "e2e4",
			want: "e4",
		},
		{
			name: "pawn capture",
			fen:  "rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq - 0 2",
			uci:  "e4d5",
			want: "exd5",
		},
		{
			name: "kingside castling",
			fen:  "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			uci:  "e1g1",
			want: "O-O",
		},
		{
			name: "queenside castling",
			fen:  "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			uci:  "e1c1",
			want: "O-O-O",
		},
		{
			name: "promotion with check",
			fen:  "4k3/P7/8/8/8/8/8/4K3 w - - 0 1",
			uci:  "a7a8q",
			want: "a8=Q+",
		},
		{
			name: "fool's mate final move (checkmate)",
			fen:  "rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3",
			uci:  "", // already in checkmate for white, test status instead
			want: "",
		},
	}

	for _, tc := range cases {
		if tc.uci == "" {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			pos := mustParseFEN(t, tc.fen)
			m, err := ParseUCI(pos, tc.uci)
			if err != nil {
				t.Fatalf("ParseUCI(%q): %v", tc.uci, err)
			}
			got := MoveSAN(pos, m)
			if got != tc.want {
				t.Errorf("MoveSAN: want %q got %q", tc.want, got)
			}
		})
	}
}

func TestMoveSANDisambiguationByFile(t *testing.T) {
	// Two knights that can both go to d2: knights on b1 and f3
	// Set up: white knights on b1 and f3, both can go to d2
	pos := mustParseFEN(t, "8/8/8/8/8/5N2/8/1N2K2k w - - 0 1")
	// Both Nb1 and Nf3 can go to d2
	// Find the move Nb1-d2
	m := Move{From: SquareOf(1, 0), To: SquareOf(3, 1)} // Nb1d2
	san := MoveSAN(pos, m)
	if san != "Nbd2" {
		t.Errorf("disambiguation by file: want Nbd2, got %q", san)
	}

	// And Nf3-d2
	m2 := Move{From: SquareOf(5, 2), To: SquareOf(3, 1)} // Nf3d2
	san2 := MoveSAN(pos, m2)
	if san2 != "Nfd2" {
		t.Errorf("disambiguation by file: want Nfd2, got %q", san2)
	}
}

func TestMoveSANDisambiguationByRank(t *testing.T) {
	// Two rooks on a1 and a5, both can go to a3
	pos := mustParseFEN(t, "8/8/8/R7/8/8/8/R3K2k w - - 0 1")
	// Ra1-a3 and Ra5-a3
	m := Move{From: SquareOf(0, 0), To: SquareOf(0, 2)} // Ra1a3
	san := MoveSAN(pos, m)
	if san != "R1a3" {
		t.Errorf("disambiguation by rank: want R1a3, got %q", san)
	}

	m2 := Move{From: SquareOf(0, 4), To: SquareOf(0, 2)} // Ra5a3
	san2 := MoveSAN(pos, m2)
	if san2 != "R5a3" {
		t.Errorf("disambiguation by rank: want R5a3, got %q", san2)
	}
}

func TestMoveSANCheckmateAnnotation(t *testing.T) {
	// Fool's mate last move: Qh4 delivers checkmate
	// Position: 1. f3 e5 2. g4 — white just played g2g4, now black to play Qd8h4#
	pos := mustParseFEN(t, "rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3")
	// The position is already after Qh4, white to move — white is in checkmate
	// Instead test at the position just before Qh4:
	pos2 := mustParseFEN(t, "rnbqkbnr/pppp1ppp/8/4p3/6P1/5P2/PPPPP2P/RNBQKBNR b KQkq - 0 2")
	m, err := ParseUCI(pos2, "d8h4")
	if err != nil {
		t.Fatal(err)
	}
	san := MoveSAN(pos2, m)
	if san != "Qh4#" {
		t.Errorf("checkmate annotation: want Qh4#, got %q", san)
	}
	_ = pos
}
