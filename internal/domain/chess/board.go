package chess

// Board is a 64-entry array representing the chess board.
type Board [64]Piece

// At returns the piece on a square.
func (b *Board) At(sq Square) Piece { return b[sq] }

// Set places a piece on a square.
func (b *Board) Set(sq Square, p Piece) { b[sq] = p }

// Clone returns a copy of the board.
func (b Board) Clone() Board { return b }

// isOccupied reports whether a square holds any piece.
func (b *Board) isOccupied(sq Square) bool { return b[sq].Type != None }

// attackedBySlider checks if any sliding piece of the given color attacks sq
// along a ray defined by (df, dr).
func (b *Board) attackedBySlider(sq Square, color Color, df, dr int, diagOK, straightOK bool) bool {
	f, r := sq.File(), sq.Rank()
	for {
		f += df
		r += dr
		if f < 0 || f > 7 || r < 0 || r > 7 {
			break
		}
		occ := b[SquareOf(f, r)]
		if occ.Type == None {
			continue
		}
		if occ.Color != color {
			break
		}
		switch occ.Type {
		case Queen:
			return true
		case Bishop:
			return diagOK
		case Rook:
			return straightOK
		}
		break
	}
	return false
}

// IsAttackedBy reports whether sq is attacked by any piece of the given color.
func (b *Board) IsAttackedBy(sq Square, color Color) bool {
	f, r := sq.File(), sq.Rank()

	// Pawn attacks: a white pawn on (f±1, r-1) attacks sq at (f, r).
	// So for color White we look at rank r-1; for Black at rank r+1.
	pawnRank := r - 1
	if color == Black {
		pawnRank = r + 1
	}
	for _, df := range []int{-1, 1} {
		pf := f + df
		if pf >= 0 && pf <= 7 && pawnRank >= 0 && pawnRank <= 7 {
			p := b[SquareOf(pf, pawnRank)]
			if p.Type == Pawn && p.Color == color {
				return true
			}
		}
	}

	// Knight attacks
	for _, d := range [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}} {
		nf, nr := f+d[0], r+d[1]
		if nf >= 0 && nf <= 7 && nr >= 0 && nr <= 7 {
			p := b[SquareOf(nf, nr)]
			if p.Type == Knight && p.Color == color {
				return true
			}
		}
	}

	// King attacks
	for _, d := range [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}} {
		kf, kr := f+d[0], r+d[1]
		if kf >= 0 && kf <= 7 && kr >= 0 && kr <= 7 {
			p := b[SquareOf(kf, kr)]
			if p.Type == King && p.Color == color {
				return true
			}
		}
	}

	// Diagonal sliders (Bishop/Queen)
	for _, d := range [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}} {
		if b.attackedBySlider(sq, color, d[0], d[1], true, false) {
			return true
		}
	}

	// Straight sliders (Rook/Queen)
	for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		if b.attackedBySlider(sq, color, d[0], d[1], false, true) {
			return true
		}
	}

	return false
}

// findKing returns the square of the king of the given color.
func (b *Board) findKing(color Color) Square {
	for i := Square(0); i < 64; i++ {
		if b[i].Type == King && b[i].Color == color {
			return i
		}
	}
	return NoSquare
}
