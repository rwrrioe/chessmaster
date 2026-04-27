package chess

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseFEN parses a FEN string and returns the corresponding Position.
func ParseFEN(fen string) (*Position, error) {
	fields := strings.Fields(fen)
	if len(fields) != 6 {
		return nil, fmt.Errorf("fen: need 6 fields, got %d", len(fields))
	}

	pos := &Position{}

	// 1. Piece placement
	ranks := strings.Split(fields[0], "/")
	if len(ranks) != 8 {
		return nil, fmt.Errorf("fen: piece placement must have 8 ranks, got %d", len(ranks))
	}
	// FEN rank 8 is index 0 in the string, which is rank index 7 in our Square layout
	for rankIdx, rankStr := range ranks {
		rank := 7 - rankIdx
		file := 0
		for _, ch := range rankStr {
			if ch >= '1' && ch <= '8' {
				file += int(ch - '0')
				continue
			}
			if file > 7 {
				return nil, fmt.Errorf("fen: rank %d overflows", rank+1)
			}
			p, err := fenCharToPiece(ch)
			if err != nil {
				return nil, err
			}
			pos.Board.Set(SquareOf(file, rank), p)
			file++
		}
		if file != 8 {
			return nil, fmt.Errorf("fen: rank %d has wrong number of squares (%d)", rank+1, file)
		}
	}

	// 2. Side to move
	switch fields[1] {
	case "w":
		pos.SideToMove = White
	case "b":
		pos.SideToMove = Black
	default:
		return nil, fmt.Errorf("fen: invalid side to move %q", fields[1])
	}

	// 3. Castling rights
	cr, err := parseCastleRights(fields[2])
	if err != nil {
		return nil, err
	}
	pos.Castle = cr

	// 4. En passant target
	if fields[3] == "-" {
		pos.EnPassantTarget = NoSquare
	} else {
		sq, err := parseSquare(fields[3])
		if err != nil {
			return nil, fmt.Errorf("fen: invalid en passant square: %w", err)
		}
		pos.EnPassantTarget = sq
	}

	// 5. Halfmove clock
	hm, err := strconv.Atoi(fields[4])
	if err != nil {
		return nil, fmt.Errorf("fen: invalid halfmove clock %q", fields[4])
	}
	pos.HalfmoveClock = hm

	// 6. Fullmove number
	fm, err := strconv.Atoi(fields[5])
	if err != nil {
		return nil, fmt.Errorf("fen: invalid fullmove number %q", fields[5])
	}
	pos.FullmoveNumber = fm

	return pos, nil
}

// FEN serializes the position back to a FEN string.
func (p *Position) FEN() string {
	var sb strings.Builder

	// 1. Piece placement
	for rankIdx := 0; rankIdx < 8; rankIdx++ {
		rank := 7 - rankIdx
		if rankIdx > 0 {
			sb.WriteByte('/')
		}
		empty := 0
		for file := 0; file < 8; file++ {
			piece := p.Board.At(SquareOf(file, rank))
			if piece.Type == None {
				empty++
				continue
			}
			if empty > 0 {
				sb.WriteByte(byte('0' + empty))
				empty = 0
			}
			sb.WriteByte(pieceToFenChar(piece))
		}
		if empty > 0 {
			sb.WriteByte(byte('0' + empty))
		}
	}

	// 2. Side to move
	if p.SideToMove == White {
		sb.WriteString(" w ")
	} else {
		sb.WriteString(" b ")
	}

	// 3. Castle rights
	sb.WriteString(castleRightsString(p.Castle))

	// 4. En passant
	sb.WriteByte(' ')
	if p.EnPassantTarget == NoSquare {
		sb.WriteByte('-')
	} else {
		sb.WriteString(p.EnPassantTarget.String())
	}

	// 5. Halfmove clock
	sb.WriteByte(' ')
	sb.WriteString(strconv.Itoa(p.HalfmoveClock))

	// 6. Fullmove number
	sb.WriteByte(' ')
	sb.WriteString(strconv.Itoa(p.FullmoveNumber))

	return sb.String()
}

func fenCharToPiece(ch rune) (Piece, error) {
	switch ch {
	case 'P':
		return Piece{White, Pawn}, nil
	case 'N':
		return Piece{White, Knight}, nil
	case 'B':
		return Piece{White, Bishop}, nil
	case 'R':
		return Piece{White, Rook}, nil
	case 'Q':
		return Piece{White, Queen}, nil
	case 'K':
		return Piece{White, King}, nil
	case 'p':
		return Piece{Black, Pawn}, nil
	case 'n':
		return Piece{Black, Knight}, nil
	case 'b':
		return Piece{Black, Bishop}, nil
	case 'r':
		return Piece{Black, Rook}, nil
	case 'q':
		return Piece{Black, Queen}, nil
	case 'k':
		return Piece{Black, King}, nil
	}
	return NoPiece, fmt.Errorf("fen: unknown piece character %q", ch)
}

func pieceToFenChar(p Piece) byte {
	chars := map[PieceType]byte{
		Pawn: 'p', Knight: 'n', Bishop: 'b', Rook: 'r', Queen: 'q', King: 'k',
	}
	ch := chars[p.Type]
	if p.Color == White {
		ch -= 32 // to uppercase
	}
	return ch
}

func parseCastleRights(s string) (CastleRights, error) {
	if s == "-" {
		return 0, nil
	}
	var cr CastleRights
	for _, ch := range s {
		switch ch {
		case 'K':
			cr |= CastleWK
		case 'Q':
			cr |= CastleWQ
		case 'k':
			cr |= CastleBK
		case 'q':
			cr |= CastleBQ
		default:
			return 0, fmt.Errorf("fen: invalid castle rights character %q", ch)
		}
	}
	return cr, nil
}

func castleRightsString(cr CastleRights) string {
	if cr == 0 {
		return "-"
	}
	var s string
	if cr.Has(CastleWK) {
		s += "K"
	}
	if cr.Has(CastleWQ) {
		s += "Q"
	}
	if cr.Has(CastleBK) {
		s += "k"
	}
	if cr.Has(CastleBQ) {
		s += "q"
	}
	return s
}

func parseSquare(s string) (Square, error) {
	if len(s) != 2 {
		return NoSquare, fmt.Errorf("en passant square %q must be 2 characters", s)
	}
	file := int(s[0] - 'a')
	rank := int(s[1] - '1')
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return NoSquare, fmt.Errorf("en passant square %q out of range", s)
	}
	return SquareOf(file, rank), nil
}
