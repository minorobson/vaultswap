package prune

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes prune results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes prune results to the given writer.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths evaluated")
		return
	}
	for _, r := range results {
		switch {
		case r.Error != nil:
			fmt.Fprintf(w, "[error]   %s: %v\n", r.Path, r.Error)
		case r.Skipped:
			fmt.Fprintf(w, "[skipped] %s (not empty)\n", r.Path)
		case r.DryRun:
			fmt.Fprintf(w, "[dry-run] %s would be pruned\n", r.Path)
		case r.Pruned:
			fmt.Fprintf(w, "[pruned]  %s\n", r.Path)
		default:
			fmt.Fprintf(w, "[unknown] %s\n", r.Path)
		}
	}
}
