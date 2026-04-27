package chess

// CastleRights encodes the four castling permissions as a bitmask.
// Bit 0 = white kingside, bit 1 = white queenside, bit 2 = black kingside, bit 3 = black queenside.
type CastleRights uint8

const (
	CastleWK CastleRights = 1 << iota
	CastleWQ
	CastleBK
	CastleBQ
)

// Has reports whether a particular right is present.
func (c CastleRights) Has(r CastleRights) bool { return c&r != 0 }

// Position is the complete game state at a single moment, equivalent to a FEN record.
type Position struct {
	Board           Board
	SideToMove      Color
	Castle          CastleRights
	EnPassantTarget Square // NoSquare when no en-passant is possible
	HalfmoveClock   int
	FullmoveNumber  int
}

// opponent returns the side not to move.
func (p *Position) opponent() Color {
	if p.SideToMove == White {
		return Black
	}
	return White
}

// Clone returns a deep copy of the position (Board is a value type so copying is cheap).
func (p *Position) Clone() *Position {
	cp := *p
	return &cp
}

// InCheck reports whether the side to move is currently in check.
func (p *Position) InCheck() bool {
	kSq := p.Board.findKing(p.SideToMove)
	if kSq == NoSquare {
		return false
	}
	return p.Board.IsAttackedBy(kSq, p.opponent())
}
