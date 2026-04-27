package stockfish_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/chessmaster-pro/chessmaster/internal/adapters/stockfish"
)

func buildFake(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	binName := "fake_stockfish"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	out := filepath.Join(dir, binName)

	// Build from the package directory (not a single file) to avoid Go 1.25 issues
	// with command-line-arguments mode.
	srcDir, err := filepath.Abs(filepath.Join("testdata", "fakesf"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "-o", out, srcDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		t.Fatalf("build fake stockfish: %v", err)
	}
	return out
}

func TestBestMove(t *testing.T) {
	t.Parallel()
	bin := buildFake(t)

	e := &stockfish.Engine{Path: bin}
	move, err := e.BestMove(context.Background(), "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 1)
	if err != nil {
		t.Fatal(err)
	}
	if move != "e2e4" {
		t.Fatalf("expected e2e4, got %s", move)
	}
}

func TestBestMoveContextCancel(t *testing.T) {
	t.Parallel()
	bin := buildFake(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	e := &stockfish.Engine{Path: bin}
	_, err := e.BestMove(ctx, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 1)
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}

func TestBestMoveInvalidBinary(t *testing.T) {
	t.Parallel()
	e := &stockfish.Engine{Path: "/nonexistent/stockfish"}
	_, err := e.BestMove(context.Background(), "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", 2)
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}
