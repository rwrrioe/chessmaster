package chess

import "fmt"

// Apply executes a move on a position and returns the new position.
// It does NOT verify legality — call LegalMoves first to obtain valid moves.
func Apply(pos *Position, m Move) (*Position, error) {
	next := pos.Clone()
	us := pos.SideToMove
	opp := pos.opponent()

	piece := pos.Board.At(m.From)
	if piece.Type == None || piece.Color != us {
		return nil, fmt.Errorf("apply: no friendly piece on %v", m.From)
	}

	captured := pos.Board.At(m.To)

	// Clear en passant target by default; set below only for double pawn push.
	next.EnPassantTarget = NoSquare

	// Update halfmove clock.
	if piece.Type == Pawn || captured.Type != None {
		next.HalfmoveClock = 0
	} else {
		next.HalfmoveClock++
	}

	if m.IsCastle {
		applyCastle(next, m, us)
	} else if m.IsEnPassant {
		applyEnPassant(next, m, us)
	} else {
		// Normal move (possibly a capture or promotion)
		next.Board.Set(m.From, NoPiece)
		if m.Promotion != None {
			next.Board.Set(m.To, Piece{us, m.Promotion})
		} else {
			next.Board.Set(m.To, piece)
		}

		// Double pawn push: set en passant target
		if piece.Type == Pawn {
			fromRank := m.From.Rank()
			toRank := m.To.Rank()
			if abs(fromRank-toRank) == 2 {
				epRank := (fromRank + toRank) / 2
				next.EnPassantTarget = SquareOf(m.From.File(), epRank)
			}
		}

		// Update castle rights when king or rook moves, or rook is captured
		next.Castle = updateCastleRights(next.Castle, m.From, m.To, us, opp)
	}

	// Switch side to move and update fullmove number
	if us == Black {
		next.FullmoveNumber++
	}
	next.SideToMove = opp

	return next, nil
}

func applyCastle(pos *Position, m Move, us Color) {
	rank := m.From.Rank()
	pos.Board.Set(m.From, NoPiece)
	pos.Board.Set(m.To, Piece{us, King})

	// Move the rook
	if m.To.File() == 6 { // Kingside
		rookFrom := SquareOf(7, rank)
		rookTo := SquareOf(5, rank)
		pos.Board.Set(rookFrom, NoPiece)
		pos.Board.Set(rookTo, Piece{us, Rook})
	} else { // Queenside
		rookFrom := SquareOf(0, rank)
		rookTo := SquareOf(3, rank)
		pos.Board.Set(rookFrom, NoPiece)
		pos.Board.Set(rookTo, Piece{us, Rook})
	}

	// Revoke all castle rights for the side that castled
	if us == White {
		pos.Castle &^= CastleWK | CastleWQ
	} else {
		pos.Castle &^= CastleBK | CastleBQ
	}
}

func applyEnPassant(pos *Position, m Move, us Color) {
	pos.Board.Set(m.From, NoPiece)
	pos.Board.Set(m.To, Piece{us, Pawn})
	// Remove the captured pawn (one rank behind the ep target, same file as destination)
	capturedRank := m.From.Rank() // the captured pawn is on the same rank as the moving pawn
	capturedSq := SquareOf(m.To.File(), capturedRank)
	pos.Board.Set(capturedSq, NoPiece)
}

// updateCastleRights revokes rights when kings or rooks move or are captured.
func updateCastleRights(cr CastleRights, from, to Square, us, opp Color) CastleRights {
	// King moves: revoke both rights for that side
	if from == SquareOf(4, 0) { // e1 white king
		cr &^= CastleWK | CastleWQ
	}
	if from == SquareOf(4, 7) { // e8 black king
		cr &^= CastleBK | CastleBQ
	}

	// Rook moves from home square
	if from == SquareOf(7, 0) {
		cr &^= CastleWK
	}
	if from == SquareOf(0, 0) {
		cr &^= CastleWQ
	}
	if from == SquareOf(7, 7) {
		cr &^= CastleBK
	}
	if from == SquareOf(0, 7) {
		cr &^= CastleBQ
	}

	// Rook captured on home square
	if to == SquareOf(7, 0) {
		cr &^= CastleWK
	}
	if to == SquareOf(0, 0) {
		cr &^= CastleWQ
	}
	if to == SquareOf(7, 7) {
		cr &^= CastleBK
	}
	if to == SquareOf(0, 7) {
		cr &^= CastleBQ
	}

	return cr
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
