package promote

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes a human-readable summary of promotion results.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes results to the given writer.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths to promote")
		return
	}

	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "[ERROR]  %s (%s -> %s): %v\n", r.Path, r.Src, r.Dst, r.Err)
			continue
		}
		label := "promoted"
		if r.DryRun {
			label = "dry-run"
		} else if r.Skipped {
			label = "skipped"
		}
		fmt.Fprintf(w, "[%-8s] %s (%s -> %s)\n", label, r.Path, r.Src, r.Dst)
	}
}
