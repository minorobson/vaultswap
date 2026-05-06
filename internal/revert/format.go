package revert

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes revert results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes revert results to the given writer.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths to revert")
		return
	}
	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "  [error]   %s (v%d): %v\n", r.Path, r.Version, r.Err)
		case r.DryRun:
			fmt.Fprintf(w, "  [dry-run] %s → v%d\n", r.Path, r.Version)
		case r.Skipped:
			fmt.Fprintf(w, "  [skipped] %s\n", r.Path)
		default:
			fmt.Fprintf(w, "  [reverted] %s → v%d\n", r.Path, r.Version)
		}
	}
}
