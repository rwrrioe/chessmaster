// Package chess implements a complete chess rules engine.
package chess

// Color represents the side of a chess piece.
type Color uint8

const (
	White Color = iota
	Black
	NoColor
)

// PieceType identifies the kind of chess piece.
type PieceType uint8

const (
	None PieceType = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

// Piece combines a Color and a PieceType.
type Piece struct {
	Color Color
	Type  PieceType
}

// NoPiece is the zero value representing an empty square.
var NoPiece = Piece{NoColor, None}

// Square is an index 0..63 in rank-major order: a1=0, b1=1, …, h8=63.
type Square uint8

const NoSquare Square = 64

// File returns the file index 0..7 (a=0, h=7).
func (s Square) File() int { return int(s) % 8 }

// Rank returns the rank index 0..7 (rank1=0, rank8=7).
func (s Square) Rank() int { return int(s) / 8 }

// SquareOf constructs a Square from file and rank indices.
func SquareOf(file, rank int) Square { return Square(rank*8 + file) }

// String returns algebraic notation (e.g. "e4").
func (s Square) String() string {
	if s >= 64 {
		return "-"
	}
	return string(rune('a'+s.File())) + string(rune('1'+s.Rank()))
}

// Move represents a single chess move in a flat, interface-free struct.
type Move struct {
	From        Square
	To          Square
	Promotion   PieceType // None unless this is a promotion
	IsCastle    bool
	IsEnPassant bool
}
