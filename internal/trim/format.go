package trim

import (
	"fmt"
	"io"
	"os"
)

const (
	labelRemoved = "removed"
	labelSkipped = "skipped"
	labelDryRun  = "dry-run"
	labelError   = "error"
)

// PrintResults writes trim results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes trim results to w.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no keys targeted")
		return
	}
	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "  %-10s %s [%s]: %v\n", labelError, r.Path, r.Key, r.Err)
		case r.DryRun && r.Removed:
			fmt.Fprintf(w, "  %-10s %s [%s]\n", labelDryRun, r.Path, r.Key)
		case r.Removed:
			fmt.Fprintf(w, "  %-10s %s [%s]\n", labelRemoved, r.Path, r.Key)
		default:
			fmt.Fprintf(w, "  %-10s %s [%s]\n", labelSkipped, r.Path, r.Key)
		}
	}
}
