package chess

import (
	"fmt"
	"strings"
)

// Game wraps a sequence of positions and moves to represent a complete chess game.
type Game struct {
	positions []*Position // positions[i] is the position before moves[i]
	moves     []Move
	sans      []string // SAN of each move for PGN
}

// NewGame returns a game starting from the standard initial position.
func NewGame() *Game {
	start, _ := ParseFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	return &Game{positions: []*Position{start}}
}

// currentPosition returns the most recent position.
func (g *Game) currentPosition() *Position {
	return g.positions[len(g.positions)-1]
}

// Move attempts to make a move. Returns an error if the move is illegal.
func (g *Game) Move(m Move) error {
	pos := g.currentPosition()
	legal := LegalMoves(pos)
	found := false
	for _, lm := range legal {
		if lm == m {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("game: illegal move %v", m.UCI())
	}

	san := MoveSAN(pos, m)
	next, err := Apply(pos, m)
	if err != nil {
		return err
	}

	g.positions = append(g.positions, next)
	g.moves = append(g.moves, m)
	g.sans = append(g.sans, san)
	return nil
}

// LegalMoves returns the legal moves in the current position.
func (g *Game) LegalMoves() []Move {
	return LegalMoves(g.currentPosition())
}

// Status returns the current game status, including threefold repetition detection.
func (g *Game) Status() GameStatusCode {
	history := make([]string, len(g.positions))
	for i, p := range g.positions {
		history[i] = positionKey(p)
	}
	return GameStatus(g.currentPosition(), history)
}

// PositionFEN returns the FEN of the current position.
func (g *Game) PositionFEN() string {
	return g.currentPosition().FEN()
}

// PGN generates a minimal PGN string with the seven-tag roster and movetext.
func (g *Game) PGN() string {
	var sb strings.Builder

	result := g.resultString()

	// Seven-tag roster
	tags := [][2]string{
		{"Event", "?"},
		{"Site", "?"},
		{"Date", "????.??.??"},
		{"Round", "?"},
		{"White", "?"},
		{"Black", "?"},
		{"Result", result},
	}
	for _, tag := range tags {
		sb.WriteString(fmt.Sprintf("[%s \"%s\"]\n", tag[0], tag[1]))
	}
	sb.WriteByte('\n')

	// Movetext
	for i, san := range g.sans {
		pos := g.positions[i]
		if pos.SideToMove == White {
			sb.WriteString(fmt.Sprintf("%d. ", pos.FullmoveNumber))
		}
		sb.WriteString(san)
		sb.WriteByte(' ')
	}
	sb.WriteString(result)
	sb.WriteByte('\n')

	return sb.String()
}

func (g *Game) resultString() string {
	status := g.Status()
	switch status {
	case Checkmate:
		if g.currentPosition().SideToMove == Black {
			return "1-0"
		}
		return "0-1"
	case Stalemate, DrawFiftyMove, DrawInsufficientMaterial, DrawByRepetition:
		return "1/2-1/2"
	}
	return "*"
}
