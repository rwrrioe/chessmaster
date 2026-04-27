// Fake UCI-compliant chess engine stub used in stockfish adapter tests.
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
		case strings.HasPrefix(line, "setoption"):
			// ignore
		case line == "isready":
			fmt.Println("readyok")
		case strings.HasPrefix(line, "position"):
			// ignore
		case strings.HasPrefix(line, "go"):
			fmt.Println("info depth 1 score cp 30")
			fmt.Println("bestmove e2e4")
		case line == "quit":
			return
		}
	}
}
