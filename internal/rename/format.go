package rename

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes rename results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes rename results to the provided writer.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths to rename")
		return
	}

	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "[error]  %s → %s: %v\n", r.Src, r.Dst, r.Err)
		case r.DryRun:
			fmt.Fprintf(w, "[dry-run] %s → %s\n", r.Src, r.Dst)
		default:
			fmt.Fprintf(w, "[renamed] %s → %s\n", r.Src, r.Dst)
		}
	}
}
