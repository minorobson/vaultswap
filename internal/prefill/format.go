package prefill

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes a human-readable summary of prefill results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes a human-readable summary of prefill results to w.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no keys to prefill")
		return
	}

	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "  ERROR    %s [%s]: %v\n", r.Path, r.Key, r.Err)
		case r.Skipped:
			fmt.Fprintf(w, "  SKIPPED  %s [%s] (key already set)\n", r.Path, r.Key)
		case r.DryRun:
			fmt.Fprintf(w, "  DRY-RUN  %s [%s] would be prefilled\n", r.Path, r.Key)
		default:
			fmt.Fprintf(w, "  PREFILLED %s [%s]\n", r.Path, r.Key)
		}
	}
}
