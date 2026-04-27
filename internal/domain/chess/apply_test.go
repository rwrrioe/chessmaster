package chess

import (
	"testing"
)

func TestApplyNormalMove(t *testing.T) {
	pos := mustParseFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	m := Move{From: SquareOf(4, 1), To: SquareOf(4, 3)} // e2-e4
	next, err := Apply(pos, m)
	if err != nil {
		t.Fatal(err)
	}
	if next.Board.At(SquareOf(4, 1)).Type != None {
		t.Error("e2 should be empty after e2-e4")
	}
	if next.Board.At(SquareOf(4, 3)) != (Piece{White, Pawn}) {
		t.Error("e4 should have white pawn")
	}
	if next.SideToMove != Black {
		t.Error("side to move should switch to Black")
	}
	if next.EnPassantTarget != SquareOf(4, 2) {
		t.Errorf("en passant target should be e3, got %v", next.EnPassantTarget)
	}
	if next.HalfmoveClock != 0 {
		t.Error("halfmove clock should reset on pawn move")
	}
}

func TestApplyEnPassantCapture(t *testing.T) {
	// White pawn on e5, black pawn just moved d7-d5, ep target is d6
	pos := mustParseFEN(t, "rnbqkbnr/ppp1pppp/8/3pP3/8/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 3")
	m := Move{From: SquareOf(4, 4), To: SquareOf(3, 5), IsEnPassant: true} // exd6
	next, err := Apply(pos, m)
	if err != nil {
		t.Fatal(err)
	}
	// White pawn moves to d6
	if next.Board.At(SquareOf(3, 5)) != (Piece{White, Pawn}) {
		t.Error("white pawn should be on d6")
	}
	// e5 is empty
	if next.Board.At(SquareOf(4, 4)).Type != None {
		t.Error("e5 should be empty")
	}
	// The captured pawn on d5 should be gone
	if next.Board.At(SquareOf(3, 4)).Type != None {
		t.Error("captured pawn on d5 should be removed")
	}
	// No new ep target
	if next.EnPassantTarget != NoSquare {
		t.Error("en passant target should be cleared after capture")
	}
}

func TestApplyPromotion(t *testing.T) {
	pos := mustParseFEN(t, "8/P7/8/8/8/8/8/4K2k w - - 0 1")
	m := Move{From: SquareOf(0, 6), To: SquareOf(0, 7), Promotion: Queen}
	next, err := Apply(pos, m)
	if err != nil {
		t.Fatal(err)
	}
	if next.Board.At(SquareOf(0, 7)) != (Piece{White, Queen}) {
		t.Errorf("a8 should be white queen after promotion, got %+v", next.Board.At(SquareOf(0, 7)))
	}
	if next.Board.At(SquareOf(0, 6)).Type != None {
		t.Error("a7 should be empty after promotion")
	}
}

func TestApplyKingsideCastle(t *testing.T) {
	pos := mustParseFEN(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	m := Move{From: SquareOf(4, 0), To: SquareOf(6, 0), IsCastle: true}
	next, err := Apply(pos, m)
	if err != nil {
		t.Fatal(err)
	}
	// King on g1
	if next.Board.At(SquareOf(6, 0)) != (Piece{White, King}) {
		t.Error("king should be on g1 after kingside castling")
	}
	// Rook on f1
	if next.Board.At(SquareOf(5, 0)) != (Piece{White, Rook}) {
		t.Error("rook should be on f1 after kingside castling")
	}
	// Old squares empty
	if next.Board.At(SquareOf(4, 0)).Type != None {
		t.Error("e1 should be empty after castling")
	}
	if next.Board.At(SquareOf(7, 0)).Type != None {
		t.Error("h1 should be empty after castling")
	}
	// Castle rights revoked for white
	if next.Castle.Has(CastleWK) || next.Castle.Has(CastleWQ) {
		t.Error("white castle rights should be revoked after king move")
	}
}

func TestApplyQueensideCastle(t *testing.T) {
	pos := mustParseFEN(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	m := Move{From: SquareOf(4, 0), To: SquareOf(2, 0), IsCastle: true}
	next, err := Apply(pos, m)
	if err != nil {
		t.Fatal(err)
	}
	if next.Board.At(SquareOf(2, 0)) != (Piece{White, King}) {
		t.Error("king should be on c1 after queenside castling")
	}
	if next.Board.At(SquareOf(3, 0)) != (Piece{White, Rook}) {
		t.Error("rook should be on d1 after queenside castling")
	}
	if next.Board.At(SquareOf(4, 0)).Type != None {
		t.Error("e1 should be empty")
	}
	if next.Board.At(SquareOf(0, 0)).Type != None {
		t.Error("a1 should be empty")
	}
}

func TestApplyRookCaptureRevokesRight(t *testing.T) {
	// White queen on h2 takes h8 rook, should revoke black's kingside castle right
	pos2 := mustParseFEN(t, "r3k2r/8/8/8/8/8/7Q/R3K3 w KQkq - 0 1")
	m := Move{From: SquareOf(7, 1), To: SquareOf(7, 7)} // Qh2xh8
	next, err := Apply(pos2, m)
	if err != nil {
		t.Fatal(err)
	}
	// Black's kingside castle right should be revoked
	if next.Castle.Has(CastleBK) {
		t.Error("black kingside castle right should be revoked when h8 rook is captured")
	}
	// Black's queenside should remain
	if !next.Castle.Has(CastleBQ) {
		t.Error("black queenside castle right should remain")
	}
}

func TestApplyHalfmoveClock(t *testing.T) {
	pos := mustParseFEN(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 10 20")
	// Non-pawn, non-capture move: knight — but no knight. Use rook.
	m := Move{From: SquareOf(0, 0), To: SquareOf(0, 1)} // Ra1-Ra2
	next, err := Apply(pos, m)
	if err != nil {
		t.Fatal(err)
	}
	if next.HalfmoveClock != 11 {
		t.Errorf("halfmove clock should be 11 after quiet rook move, got %d", next.HalfmoveClock)
	}
	// Pawn move resets clock
	pos2 := mustParseFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 10 5")
	m2 := Move{From: SquareOf(4, 1), To: SquareOf(4, 2)} // e2-e3
	next2, err := Apply(pos2, m2)
	if err != nil {
		t.Fatal(err)
	}
	if next2.HalfmoveClock != 0 {
		t.Errorf("halfmove clock should reset on pawn move, got %d", next2.HalfmoveClock)
	}
}

func TestApplyFullmoveIncrement(t *testing.T) {
	pos := mustParseFEN(t, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")
	m := Move{From: SquareOf(4, 6), To: SquareOf(4, 5)} // e7-e6
	next, err := Apply(pos, m)
	if err != nil {
		t.Fatal(err)
	}
	// Fullmove increments after black's move
	if next.FullmoveNumber != 2 {
		t.Errorf("fullmove should be 2 after black moves, got %d", next.FullmoveNumber)
	}
}

func TestApplyDoublePushSetsEP(t *testing.T) {
	pos := mustParseFEN(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	m := Move{From: SquareOf(4, 1), To: SquareOf(4, 3)} // e2-e4
	next, err := Apply(pos, m)
	if err != nil {
		t.Fatal(err)
	}
	if next.EnPassantTarget != SquareOf(4, 2) {
		t.Errorf("double push should set ep to e3, got %v", next.EnPassantTarget)
	}
	// Subsequent non-double-push clears ep
	m2 := Move{From: SquareOf(4, 6), To: SquareOf(4, 5)} // e7-e6
	next2, err := Apply(next, m2)
	if err != nil {
		t.Fatal(err)
	}
	if next2.EnPassantTarget != NoSquare {
		t.Errorf("ep target should be cleared after non-double-push, got %v", next2.EnPassantTarget)
	}
}

func TestApplyRookMoveLosesCastleRight(t *testing.T) {
	pos := mustParseFEN(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	// Move white h-rook
	m := Move{From: SquareOf(7, 0), To: SquareOf(7, 1)} // Rh1-Rh2
	next, err := Apply(pos, m)
	if err != nil {
		t.Fatal(err)
	}
	if next.Castle.Has(CastleWK) {
		t.Error("white kingside castle right should be revoked after h1 rook moves")
	}
	if !next.Castle.Has(CastleWQ) {
		t.Error("white queenside right should remain")
	}
}
