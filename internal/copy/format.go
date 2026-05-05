package copy

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes copy results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes copy results to the given writer.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths to copy")
		return
	}
	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "[error]  %s → %s: %v\n", r.Src, r.Dst, r.Err)
		case r.DryRun:
			fmt.Fprintf(w, "[dry-run] %s → %s\n", r.Src, r.Dst)
		default:
			fmt.Fprintf(w, "[copied]  %s → %s\n", r.Src, r.Dst)
		}
	}
}
