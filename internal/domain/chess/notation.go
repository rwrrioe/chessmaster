package chess

import (
	"fmt"
	"strings"
)

// UCI returns the UCI representation of a move (e.g. "e2e4", "e7e8q").
// Castling is represented as a king move (e.g. "e1g1"), not "O-O".
func (m Move) UCI() string {
	s := m.From.String() + m.To.String()
	if m.Promotion != None {
		s += string(promoChar(m.Promotion))
	}
	return s
}

// ParseUCI parses a UCI move string (e.g. "e2e4", "e7e8q") into a Move.
// It validates that the squares are valid and the promotion piece (if any) is valid.
func ParseUCI(pos *Position, s string) (Move, error) {
	s = strings.TrimSpace(s)
	if len(s) < 4 || len(s) > 5 {
		return Move{}, fmt.Errorf("uci: invalid move %q", s)
	}

	from, err := parseSquare(s[0:2])
	if err != nil {
		return Move{}, fmt.Errorf("uci: invalid from square in %q: %w", s, err)
	}
	to, err := parseSquare(s[2:4])
	if err != nil {
		return Move{}, fmt.Errorf("uci: invalid to square in %q: %w", s, err)
	}

	m := Move{From: from, To: to}

	if len(s) == 5 {
		pt, err := promoCharToPieceType(rune(s[4]))
		if err != nil {
			return Move{}, fmt.Errorf("uci: invalid promotion in %q: %w", s, err)
		}
		m.Promotion = pt
	}

	// Detect castling and en passant from the position context
	if pos != nil {
		piece := pos.Board.At(from)
		if piece.Type == King {
			df := to.File() - from.File()
			if df == 2 || df == -2 {
				m.IsCastle = true
			}
		}
		if piece.Type == Pawn && pos.EnPassantTarget != NoSquare && to == pos.EnPassantTarget {
			m.IsEnPassant = true
		}
	}

	return m, nil
}

// MoveSAN returns the Standard Algebraic Notation of a move in the given position.
func MoveSAN(pos *Position, m Move) string {
	// Castling
	if m.IsCastle {
		if m.To.File() == 6 {
			return "O-O"
		}
		return "O-O-O"
	}

	piece := pos.Board.At(m.From)
	var sb strings.Builder

	// Piece letter (omit for pawns)
	if piece.Type != Pawn {
		sb.WriteByte(pieceTypeLetter(piece.Type))
	}

	// Disambiguation
	if piece.Type != Pawn && piece.Type != King {
		disambig := disambiguation(pos, m, piece)
		sb.WriteString(disambig)
	}

	// Capture mark
	captured := pos.Board.At(m.To)
	isCapture := captured.Type != None || m.IsEnPassant
	if isCapture && piece.Type == Pawn {
		sb.WriteByte(byte('a' + m.From.File()))
	}
	if isCapture {
		sb.WriteByte('x')
	}

	// Destination
	sb.WriteString(m.To.String())

	// Promotion
	if m.Promotion != None {
		sb.WriteByte('=')
		sb.WriteByte(pieceTypeLetter(m.Promotion))
	}

	// Check / mate annotation — apply the move and see
	next, err := Apply(pos, m)
	if err == nil {
		legal := LegalMoves(next)
		if len(legal) == 0 && next.InCheck() {
			sb.WriteByte('#')
		} else if next.InCheck() {
			sb.WriteByte('+')
		}
	}

	return sb.String()
}

func pieceTypeLetter(pt PieceType) byte {
	switch pt {
	case Knight:
		return 'N'
	case Bishop:
		return 'B'
	case Rook:
		return 'R'
	case Queen:
		return 'Q'
	case King:
		return 'K'
	}
	return '?'
}

// disambiguation returns the file, rank, or both needed to disambiguate a move.
func disambiguation(pos *Position, m Move, piece Piece) string {
	// Find all other pieces of the same type and color that can also move to m.To
	var ambig []Square
	for sq := Square(0); sq < 64; sq++ {
		if sq == m.From {
			continue
		}
		p := pos.Board.At(sq)
		if p != piece {
			continue
		}
		// Check if this piece has a legal move to m.To
		for _, lm := range LegalMoves(pos) {
			if lm.From == sq && lm.To == m.To {
				ambig = append(ambig, sq)
				break
			}
		}
	}
	if len(ambig) == 0 {
		return ""
	}

	// Check if file disambiguates
	sameFile := false
	sameRank := false
	for _, sq := range ambig {
		if sq.File() == m.From.File() {
			sameFile = true
		}
		if sq.Rank() == m.From.Rank() {
			sameRank = true
		}
	}

	if !sameFile {
		return string(rune('a' + m.From.File()))
	}
	if !sameRank {
		return string(rune('1' + m.From.Rank()))
	}
	// Need both file and rank
	return string(rune('a'+m.From.File())) + string(rune('1'+m.From.Rank()))
}

func promoChar(pt PieceType) rune {
	switch pt {
	case Queen:
		return 'q'
	case Rook:
		return 'r'
	case Bishop:
		return 'b'
	case Knight:
		return 'n'
	}
	return '?'
}

func promoCharToPieceType(ch rune) (PieceType, error) {
	switch ch {
	case 'q':
		return Queen, nil
	case 'r':
		return Rook, nil
	case 'b':
		return Bishop, nil
	case 'n':
		return Knight, nil
	}
	return None, fmt.Errorf("unknown promotion piece %q", ch)
}
