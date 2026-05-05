package archive

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes archive results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes archive results to w.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths to archive")
		return
	}
	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "  ERROR    %s: %v\n", r.Path, r.Err)
		case r.DryRun:
			fmt.Fprintf(w, "  DRY-RUN  %s (version %d)\n", r.Path, r.Version)
		case r.Skipped:
			fmt.Fprintf(w, "  SKIPPED  %s\n", r.Path)
		default:
			fmt.Fprintf(w, "  ARCHIVED %s (version %d)\n", r.Path, r.Version)
		}
	}
}
