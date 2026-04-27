package ports

import "context"

// Engine represents a chess AI that can suggest the best move for a position.
type Engine interface {
	// BestMove returns the best UCI move for the given FEN at the specified level (1=easy, 2=medium, 3=hard).
	BestMove(ctx context.Context, fen string, level int) (string, error)
}
