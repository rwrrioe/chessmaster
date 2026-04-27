package ports

import "context"

// Severity classifies how bad a chess mistake is.
type Severity string

const (
	SevInaccuracy Severity = "inaccuracy"
	SevMistake    Severity = "mistake"
	SevBlunder    Severity = "blunder"
)

// Mistake describes a single error found during game analysis.
type Mistake struct {
	Ply      int      `json:"ply"`
	Move     string   `json:"move"`
	Severity Severity `json:"severity"`
	Better   string   `json:"better,omitempty"`
	Comment  string   `json:"comment"`
}

// Analysis is the result of analysing a complete game.
type Analysis struct {
	Summary  string    `json:"summary"`
	Mistakes []Mistake `json:"mistakes"`
}

// Coach analyses a chess game given its PGN and returns coaching feedback.
type Coach interface {
	Analyze(ctx context.Context, pgn string) (Analysis, error)
}
