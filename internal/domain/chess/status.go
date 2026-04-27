package chess

// GameStatusCode represents the current state of a chess game.
type GameStatusCode int

const (
	Ongoing GameStatusCode = iota
	Check
	Checkmate
	Stalemate
	DrawFiftyMove
	DrawInsufficientMaterial
	DrawByRepetition
)

func (g GameStatusCode) String() string {
	switch g {
	case Ongoing:
		return "Ongoing"
	case Check:
		return "Check"
	case Checkmate:
		return "Checkmate"
	case Stalemate:
		return "Stalemate"
	case DrawFiftyMove:
		return "DrawFiftyMove"
	case DrawInsufficientMaterial:
		return "DrawInsufficientMaterial"
	case DrawByRepetition:
		return "DrawByRepetition"
	}
	return "Unknown"
}

// GameStatus returns the status of the current position.
// positionHistory, if provided, is used to detect threefold repetition.
func GameStatus(pos *Position, positionHistory []string) GameStatusCode {
	// Fifty-move rule (halfmove clock >= 100 means 50 full moves without capture/pawn)
	if pos.HalfmoveClock >= 100 {
		return DrawFiftyMove
	}

	// Insufficient material
	if isInsufficientMaterial(&pos.Board) {
		return DrawInsufficientMaterial
	}

	// Threefold repetition (checked externally via history)
	if positionHistory != nil {
		key := positionKey(pos)
		count := 0
		for _, k := range positionHistory {
			if k == key {
				count++
			}
		}
		if count >= 3 {
			return DrawByRepetition
		}
	}

	legal := LegalMoves(pos)
	if len(legal) == 0 {
		if pos.InCheck() {
			return Checkmate
		}
		return Stalemate
	}

	if pos.InCheck() {
		return Check
	}

	return Ongoing
}

// positionKey returns a string that uniquely identifies a position for repetition detection.
// It uses FEN without halfmove clock and fullmove number.
func positionKey(pos *Position) string {
	// Build just the piece placement + side + castle + ep fields
	p := pos.Clone()
	p.HalfmoveClock = 0
	p.FullmoveNumber = 1
	return p.FEN()
}

// isInsufficientMaterial returns true when neither side can force checkmate.
func isInsufficientMaterial(pos *Board) bool {
	var whitePieces, blackPieces []Piece
	for sq := Square(0); sq < 64; sq++ {
		p := pos[sq]
		if p.Type == None || p.Type == King {
			continue
		}
		if p.Color == White {
			whitePieces = append(whitePieces, p)
		} else {
			blackPieces = append(blackPieces, p)
		}
	}

	// If either side has pawns, rooks, or queens, not insufficient
	for _, pieces := range [][]Piece{whitePieces, blackPieces} {
		for _, p := range pieces {
			if p.Type == Pawn || p.Type == Rook || p.Type == Queen {
				return false
			}
		}
	}

	// K vs K
	if len(whitePieces) == 0 && len(blackPieces) == 0 {
		return true
	}

	// KB vs K or KN vs K (either side)
	if len(whitePieces) == 0 && len(blackPieces) == 1 {
		return true
	}
	if len(whitePieces) == 1 && len(blackPieces) == 0 {
		return true
	}

	// KB vs KB — only if bishops are on same color squares
	if len(whitePieces) == 1 && whitePieces[0].Type == Bishop &&
		len(blackPieces) == 1 && blackPieces[0].Type == Bishop {
		wSq := bishopSquare(pos, White)
		bSq := bishopSquare(pos, Black)
		if wSq != NoSquare && bSq != NoSquare {
			wColor := (wSq.File() + wSq.Rank()) % 2
			bColor := (bSq.File() + bSq.Rank()) % 2
			return wColor == bColor
		}
	}

	return false
}

func bishopSquare(b *Board, color Color) Square {
	for sq := Square(0); sq < 64; sq++ {
		if b[sq].Type == Bishop && b[sq].Color == color {
			return sq
		}
	}
	return NoSquare
}
