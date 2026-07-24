//go:build !linux

package buttons

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// Start listens on stdin: p/n/a short, P/N/A long (+ Enter).
func Start(ctx context.Context, on Handler) error {
	fmt.Println("buttons: keyboard mode (p/n/a short, P/N/A long + Enter)")
	go func() {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}
			switch strings.TrimSpace(sc.Text()) {
			case "p":
				on(Prev)
			case "n":
				on(Next)
			case "a":
				on(Action)
			case "P":
				on(PrevLong)
			case "N":
				on(NextLong)
			case "A":
				on(ActionLong)
			}
		}
	}()
	return nil
}
