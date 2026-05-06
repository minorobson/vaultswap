package pin

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes pin results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes pin results to w.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths to pin")
		return
	}
	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "  ERROR   %s: %v\n", r.Path, r.Err)
		case r.DryRun:
			fmt.Fprintf(w, "  DRY-RUN %s → version %d\n", r.Path, r.Version)
		default:
			fmt.Fprintf(w, "  PINNED  %s → version %d\n", r.Path, r.Version)
		}
	}
}
