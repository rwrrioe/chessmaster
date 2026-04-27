//go:build ignore

// fake_stockfish is a minimal UCI-compliant stub used in engine tests.
// It responds to uci with "uciok", to isready with "readyok",
// and to "go" with "bestmove e2e4".
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch {
		case line == "uci":
			fmt.Println("id name FakeStockfish")
			fmt.Println("uciok")
		case line == "isready":
			fmt.Println("readyok")
		case strings.HasPrefix(line, "go"):
			fmt.Println("info depth 1 score cp 30")
			fmt.Println("bestmove e2e4")
		case line == "quit":
			return
		}
	}
}
