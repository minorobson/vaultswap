package clone

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes a human-readable summary of clone results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes clone results to the provided writer.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths to clone")
		return
	}

	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "  ERROR   %s: %v\n", r.Path, r.Err)
		case r.Skipped:
			fmt.Fprintf(w, "  SKIP    %s (no changes)\n", r.Path)
		case r.DryRun:
			fmt.Fprintf(w, "  DRY-RUN %s (would clone)\n", r.Path)
		default:
			fmt.Fprintf(w, "  CLONED  %s\n", r.Path)
		}
	}
}
