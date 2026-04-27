// Package stockfish provides a chess engine adapter that communicates via UCI protocol.
package stockfish

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// levelConfig maps level to Stockfish skill level and movetime.
type levelConfig struct {
	skill    int
	movetime time.Duration
}

var levels = map[int]levelConfig{
	1: {skill: 3, movetime: 200 * time.Millisecond},
	2: {skill: 10, movetime: 500 * time.Millisecond},
	3: {skill: 20, movetime: 1500 * time.Millisecond},
}

// Engine calls Stockfish per move via UCI. Thread-safe; each call spawns a fresh process.
type Engine struct {
	Path string
}

// New creates an Engine, resolving the binary path from STOCKFISH_PATH env or defaults.
func New() *Engine {
	path := os.Getenv("STOCKFISH_PATH")
	if path == "" {
		path = "stockfish"
	}
	return &Engine{Path: path}
}

// BestMove asks Stockfish for the best move in the given position. Level must be 1, 2, or 3.
func (e *Engine) BestMove(ctx context.Context, fen string, level int) (string, error) {
	cfg, ok := levels[level]
	if !ok {
		cfg = levels[2]
	}

	timeout := cfg.movetime*2 + 3*time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.Path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("stockfish stdin: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stockfish stdout: %w", err)
	}
	if err = cmd.Start(); err != nil {
		return "", fmt.Errorf("stockfish start: %w", err)
	}
	defer func() {
		stdin.Close()
		cmd.Wait() //nolint
	}()

	commands := fmt.Sprintf(
		"uci\nsetoption name Skill Level value %d\nisready\nposition fen %s\ngo movetime %d\n",
		cfg.skill, fen, cfg.movetime.Milliseconds(),
	)
	if _, err = io.WriteString(stdin, commands); err != nil {
		return "", fmt.Errorf("stockfish write: %w", err)
	}

	return parseBestMove(ctx, stdout)
}

func parseBestMove(ctx context.Context, r io.Reader) (string, error) {
	type result struct {
		move string
		err  error
	}
	ch := make(chan result, 1)

	go func() {
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "bestmove") {
				parts := strings.Fields(line)
				if len(parts) >= 2 && parts[1] != "(none)" {
					ch <- result{move: parts[1]}
					return
				}
				ch <- result{err: fmt.Errorf("stockfish: no move available")}
				return
			}
		}
		if err := sc.Err(); err != nil {
			ch <- result{err: fmt.Errorf("stockfish scan: %w", err)}
			return
		}
		ch <- result{err: fmt.Errorf("stockfish: stdout closed without bestmove")}
	}()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("stockfish: %w", ctx.Err())
	case res := <-ch:
		return res.move, res.err
	}
}
