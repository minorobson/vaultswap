package protect

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes a human-readable summary of protect results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes a human-readable summary of protect results to w.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths to protect")
		return
	}
	for _, r := range results {
		switch {
		case r.Error != nil:
			fmt.Fprintf(w, "  [error]     %s — %v\n", r.Path, r.Error)
		case r.Skipped:
			fmt.Fprintf(w, "  [skipped]   %s (already protected)\n", r.Path)
		case r.DryRun:
			fmt.Fprintf(w, "  [dry-run]   %s\n", r.Path)
		default:
			fmt.Fprintf(w, "  [protected] %s\n", r.Path)
		}
	}
}
