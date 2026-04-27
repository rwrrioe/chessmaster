package chess

import (
	"fmt"
	"sort"
	"testing"
)

func mustParseFEN(t *testing.T, fen string) *Position {
	t.Helper()
	pos, err := ParseFEN(fen)
	if err != nil {
		t.Fatalf("ParseFEN(%q): %v", fen, err)
	}
	return pos
}

func moveSet(moves []Move) []string {
	s := make([]string, len(moves))
	for i, m := range moves {
		s[i] = m.UCI()
	}
	sort.Strings(s)
	return s
}

func TestLegalMoveCount(t *testing.T) {
	cases := []struct {
		name  string
		fen   string
		count int
	}{
		{
			name:  "starting position",
			fen:   "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			count: 20,
		},
		{
			name:  "kiwipete",
			fen:   "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
			count: 48,
		},
		{
			name:  "en passant available",
			fen:   "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3",
			count: 31,
		},
		{
			name:  "king in check - only evasions",
			fen:   "8/8/8/8/8/8/6K1/4k1r1 w - - 0 1",
			count: 4, // king can go to f2 h2 h1 f1 - but must verify each doesn't leave in check
		},
		{
			name:  "only king vs king",
			fen:   "8/8/8/8/8/8/8/4K2k w - - 0 1",
			count: 5,
		},
		{
			name:  "castling both sides allowed",
			fen:   "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1",
			count: 26,
		},
		{
			name:  "castling blocked by piece in between",
			fen:   "r3k2r/8/8/8/8/8/8/RN2K2R w KQkq - 0 1",
			count: 25, // no queenside castling for white (b1 has a knight)
		},
		{
			name:  "no legal moves for pinned piece",
			fen:   "8/8/8/8/8/b7/3P4/3K4 w - - 0 1",
			count: 5, // king moves only; d2 pawn is pinned by bishop on a3
		},
		{
			name:  "promotion position",
			fen:   "8/P7/8/8/8/8/8/4K2k w - - 0 1",
			count: 9, // 4 promotions on a8 + 5 king moves
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pos := mustParseFEN(t, tc.fen)
			moves := LegalMoves(pos)
			if len(moves) != tc.count {
				uci := moveSet(moves)
				t.Errorf("LegalMoves(%q): want %d, got %d\nMoves: %v", tc.fen, tc.count, len(moves), uci)
			}
		})
	}
}

func TestCastlingBlocked(t *testing.T) {
	// King passes through attacked square — queenside castling blocked
	pos := mustParseFEN(t, "r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1")
	moves := LegalMoves(pos)
	// Verify kingside and queenside castles are both present when unattacked
	var hasKS, hasQS bool
	for _, m := range moves {
		if m.IsCastle && m.From == SquareOf(4, 7) {
			if m.To == SquareOf(6, 7) {
				hasKS = true
			}
			if m.To == SquareOf(2, 7) {
				hasQS = true
			}
		}
	}
	if !hasKS {
		t.Error("expected kingside castling for black")
	}
	if !hasQS {
		t.Error("expected queenside castling for black")
	}

	// Now attack the transit square d8 (c8 is transit for QS) — block QS
	// White rook on c1 attacks c8 (transit square for black QS)
	pos2 := mustParseFEN(t, "r3k2r/8/8/8/8/8/8/2R1K2R b Kkq - 0 1")
	moves2 := LegalMoves(pos2)
	for _, m := range moves2 {
		if m.IsCastle && m.From == SquareOf(4, 7) && m.To == SquareOf(2, 7) {
			t.Error("queenside castling should be blocked when c8 is attacked")
		}
	}
}

func TestCastlingNotAllowedInCheck(t *testing.T) {
	// King in check must not castle
	pos := mustParseFEN(t, "4k2r/8/8/8/8/8/8/4K2R b kq - 0 1")
	// Put white rook on e1 attacking e8 to put black king in check?
	// Use a position where black king is in check from a rook on e1
	pos2 := mustParseFEN(t, "r3k2r/8/8/8/8/8/8/R3K1NR b KQkq - 0 1")
	_ = pos
	_ = pos2
	// King in check from white rook on e-file:
	checkPos := mustParseFEN(t, "r3k2r/8/8/8/8/8/8/4R3 b kq - 0 1")
	moves := LegalMoves(checkPos)
	for _, m := range moves {
		if m.IsCastle {
			t.Errorf("castling not allowed when in check: got castle move %v", m.UCI())
		}
	}
}

func TestEnPassantCapture(t *testing.T) {
	// White pawn on e5, black pawn just moved d7-d5
	pos := mustParseFEN(t, "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3")
	moves := LegalMoves(pos)
	var hasEP bool
	for _, m := range moves {
		if m.IsEnPassant {
			hasEP = true
			if m.From != SquareOf(4, 4) || m.To != SquareOf(3, 5) {
				t.Errorf("unexpected EP move: %v", m.UCI())
			}
		}
	}
	if !hasEP {
		t.Error("expected en passant capture")
	}
}

func TestPromotionMoves(t *testing.T) {
	pos := mustParseFEN(t, "8/P7/8/8/8/8/8/4K2k w - - 0 1")
	moves := LegalMoves(pos)
	promos := map[PieceType]bool{}
	for _, m := range moves {
		if m.Promotion != None {
			promos[m.Promotion] = true
		}
	}
	for _, pt := range []PieceType{Queen, Rook, Bishop, Knight} {
		if !promos[pt] {
			t.Errorf("missing promotion to %v", pt)
		}
	}
}

func TestKingCannotMoveIntoCheck(t *testing.T) {
	// King on e1, black queen on e8 — king cannot move to e2
	pos := mustParseFEN(t, "4q3/8/8/8/8/8/8/4K2k w - - 0 1")
	moves := LegalMoves(pos)
	for _, m := range moves {
		if m.From == SquareOf(4, 0) && m.To == SquareOf(4, 1) {
			t.Errorf("king should not be able to move to e2 (attacked by queen on e8)")
		}
	}
	// Also must have some legal moves (f1, d1, d2, f2 if safe)
	if len(moves) == 0 {
		t.Error("expected some legal moves for king")
	}
}

func TestPerft1StartingPosition(t *testing.T) {
	// perft(1) from start = 20
	pos := mustParseFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	got := len(LegalMoves(pos))
	if got != 20 {
		t.Errorf("perft(1) from start: want 20, got %d", got)
	}
}

// perft counts leaf nodes at a given depth.
func perft(pos *Position, depth int) int {
	if depth == 0 {
		return 1
	}
	moves := LegalMoves(pos)
	if depth == 1 {
		return len(moves)
	}
	total := 0
	for _, m := range moves {
		next, err := Apply(pos, m)
		if err != nil {
			continue
		}
		total += perft(next, depth-1)
	}
	return total
}

func TestPerft2StartingPosition(t *testing.T) {
	pos := mustParseFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	got := perft(pos, 2)
	if got != 400 {
		t.Errorf("perft(2) from start: want 400, got %d", got)
	}
}

func TestPerft3StartingPosition(t *testing.T) {
	pos := mustParseFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	got := perft(pos, 3)
	if got != 8902 {
		t.Errorf("perft(3) from start: want 8902, got %d", got)
	}
}

func TestPerftKiwipete(t *testing.T) {
	pos := mustParseFEN(t, "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	got := perft(pos, 1)
	if got != 48 {
		t.Errorf("perft(1) kiwipete: want 48, got %d", got)
	}
}

func TestCastlingRightsLost(t *testing.T) {
	// After white king moves, castle rights should be gone
	pos := mustParseFEN(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	// Move king e1-f1
	kingMove := Move{From: SquareOf(4, 0), To: SquareOf(5, 0)}
	next, err := Apply(pos, kingMove)
	if err != nil {
		t.Fatal(err)
	}
	if next.Castle.Has(CastleWK) || next.Castle.Has(CastleWQ) {
		t.Error("white castle rights should be gone after king move")
	}
	if !next.Castle.Has(CastleBK) || !next.Castle.Has(CastleBQ) {
		t.Error("black castle rights should remain")
	}
}

func TestLegalMovesKingInCheck(t *testing.T) {
	// After e4 e5 Qh5, black must respond to check threat — but here use a direct check
	// White rook gives check on e8 — black must block or move king
	pos := mustParseFEN(t, "4k3/8/8/8/8/8/8/4R2K b - - 0 1")
	moves := LegalMoves(pos)
	// King can go to d8, f8, d7, f7 - all 4 directions not blocked
	// e8 is attacked so can't stay; e7 is attacked by rook
	// valid: d8, f8, d7, f7
	if len(moves) == 0 {
		t.Error("expected evasion moves")
	}
	// Verify all moves result in king not being in check
	for _, m := range moves {
		next, err := Apply(pos, m)
		if err != nil {
			t.Errorf("Apply failed for move %v: %v", m.UCI(), err)
			continue
		}
		// The side that just moved is now opponent
		_ = fmt.Sprintf("applied %v", m.UCI())
		// After black moves, it's white's turn. Check if black king is still in check
		// by temporarily swapping
		if next.Board.IsAttackedBy(next.Board.findKing(Black), White) {
			t.Errorf("move %v leaves black king in check", m.UCI())
		}
	}
}
