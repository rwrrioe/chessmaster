package chess

// LegalMoves returns all legal moves for the side to move in the given position.
func LegalMoves(pos *Position) []Move {
	pseudo := pseudoLegal(pos)
	legal := pseudo[:0:0]
	for _, m := range pseudo {
		next, err := Apply(pos, m)
		if err != nil {
			continue
		}
		// After applying the move it is the opponent's turn.
		// Check whether our king is now attacked by the opponent.
		kSq := next.Board.findKing(pos.SideToMove)
		if kSq == NoSquare {
			continue
		}
		if !next.Board.IsAttackedBy(kSq, next.SideToMove) {
			legal = append(legal, m)
		}
	}
	return legal
}

// pseudoLegal generates all candidate moves without checking for leaving the king in check.
func pseudoLegal(pos *Position) []Move {
	var moves []Move
	us := pos.SideToMove

	for sq := Square(0); sq < 64; sq++ {
		p := pos.Board.At(sq)
		if p.Color != us || p.Type == None {
			continue
		}
		switch p.Type {
		case Pawn:
			moves = append(moves, pawnMoves(pos, sq)...)
		case Knight:
			moves = append(moves, knightMoves(pos, sq)...)
		case Bishop:
			moves = append(moves, slidingMoves(pos, sq, [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}})...)
		case Rook:
			moves = append(moves, slidingMoves(pos, sq, [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}})...)
		case Queen:
			moves = append(moves, slidingMoves(pos, sq, [][2]int{
				{-1, -1}, {-1, 1}, {1, -1}, {1, 1},
				{-1, 0}, {1, 0}, {0, -1}, {0, 1},
			})...)
		case King:
			moves = append(moves, kingMoves(pos, sq)...)
		}
	}
	return moves
}

func pawnMoves(pos *Position, sq Square) []Move {
	var moves []Move
	f, r := sq.File(), sq.Rank()
	us := pos.Board.At(sq).Color

	dir := 1
	startRank := 1
	promoRank := 6
	if us == Black {
		dir = -1
		startRank = 6
		promoRank = 1
	}

	// Single push
	nr := r + dir
	if nr >= 0 && nr <= 7 {
		dest := SquareOf(f, nr)
		if pos.Board.At(dest).Type == None {
			if r == promoRank {
				moves = append(moves, promoMoves(sq, dest)...)
			} else {
				moves = append(moves, Move{From: sq, To: dest})
			}
			// Double push from starting rank
			if r == startRank {
				nr2 := r + 2*dir
				dest2 := SquareOf(f, nr2)
				if pos.Board.At(dest2).Type == None {
					moves = append(moves, Move{From: sq, To: dest2})
				}
			}
		}
	}

	// Captures
	for _, df := range []int{-1, 1} {
		nf := f + df
		if nf < 0 || nf > 7 || nr < 0 || nr > 7 {
			continue
		}
		dest := SquareOf(nf, nr)
		target := pos.Board.At(dest)
		if target.Type != None && target.Color != us {
			if r == promoRank {
				moves = append(moves, promoMoves(sq, dest)...)
			} else {
				moves = append(moves, Move{From: sq, To: dest})
			}
		}
		// En passant
		if pos.EnPassantTarget != NoSquare && dest == pos.EnPassantTarget {
			moves = append(moves, Move{From: sq, To: dest, IsEnPassant: true})
		}
	}

	return moves
}

func promoMoves(from, to Square) []Move {
	return []Move{
		{From: from, To: to, Promotion: Queen},
		{From: from, To: to, Promotion: Rook},
		{From: from, To: to, Promotion: Bishop},
		{From: from, To: to, Promotion: Knight},
	}
}

func knightMoves(pos *Position, sq Square) []Move {
	var moves []Move
	f, r := sq.File(), sq.Rank()
	us := pos.Board.At(sq).Color
	for _, d := range [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}} {
		nf, nr := f+d[0], r+d[1]
		if nf < 0 || nf > 7 || nr < 0 || nr > 7 {
			continue
		}
		dest := SquareOf(nf, nr)
		target := pos.Board.At(dest)
		if target.Type != None && target.Color == us {
			continue
		}
		moves = append(moves, Move{From: sq, To: dest})
	}
	return moves
}

func slidingMoves(pos *Position, sq Square, dirs [][2]int) []Move {
	var moves []Move
	f, r := sq.File(), sq.Rank()
	us := pos.Board.At(sq).Color
	for _, d := range dirs {
		cf, cr := f, r
		for {
			cf += d[0]
			cr += d[1]
			if cf < 0 || cf > 7 || cr < 0 || cr > 7 {
				break
			}
			dest := SquareOf(cf, cr)
			target := pos.Board.At(dest)
			if target.Type == None {
				moves = append(moves, Move{From: sq, To: dest})
				continue
			}
			// Occupied square: capture if enemy, always stop
			if target.Color != us {
				moves = append(moves, Move{From: sq, To: dest})
			}
			break
		}
	}
	return moves
}

func kingMoves(pos *Position, sq Square) []Move {
	var moves []Move
	f, r := sq.File(), sq.Rank()
	us := pos.Board.At(sq).Color
	for _, d := range [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}} {
		nf, nr := f+d[0], r+d[1]
		if nf < 0 || nf > 7 || nr < 0 || nr > 7 {
			continue
		}
		dest := SquareOf(nf, nr)
		target := pos.Board.At(dest)
		if target.Type != None && target.Color == us {
			continue
		}
		moves = append(moves, Move{From: sq, To: dest})
	}

	// Castling
	moves = append(moves, castleMoves(pos, sq)...)
	return moves
}

func castleMoves(pos *Position, kingSq Square) []Move {
	var moves []Move
	us := pos.SideToMove
	opp := pos.opponent()

	// King must not currently be in check
	if pos.Board.IsAttackedBy(kingSq, opp) {
		return nil
	}

	rank := 0
	wkRight := CastleWK
	wqRight := CastleWQ
	if us == Black {
		rank = 7
		wkRight = CastleBK
		wqRight = CastleBQ
	}

	// Kingside
	if pos.Castle.Has(wkRight) {
		f1 := SquareOf(5, rank)
		g1 := SquareOf(6, rank)
		// Squares between king and rook must be empty
		if pos.Board.At(f1).Type == None && pos.Board.At(g1).Type == None {
			// King must not pass through or land on attacked square
			if !pos.Board.IsAttackedBy(f1, opp) && !pos.Board.IsAttackedBy(g1, opp) {
				moves = append(moves, Move{From: kingSq, To: g1, IsCastle: true})
			}
		}
	}

	// Queenside
	if pos.Castle.Has(wqRight) {
		b1 := SquareOf(1, rank)
		c1 := SquareOf(2, rank)
		d1 := SquareOf(3, rank)
		// Squares between king and rook must be empty (b1, c1, d1)
		if pos.Board.At(b1).Type == None && pos.Board.At(c1).Type == None && pos.Board.At(d1).Type == None {
			// King passes through d1 and c1; both must be safe
			if !pos.Board.IsAttackedBy(d1, opp) && !pos.Board.IsAttackedBy(c1, opp) {
				moves = append(moves, Move{From: kingSq, To: c1, IsCastle: true})
			}
		}
	}

	return moves
}
