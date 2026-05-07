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

// PrintSummary writes a brief count summary of promotion results to os.Stdout.
func PrintSummary(results []Result) {
	FprintSummary(os.Stdout, results)
}

// FprintSummary writes a brief count summary of promotion results to the given writer.
func FprintSummary(w io.Writer, results []Result) {
	var promoted, skipped, dryRun, errored int
	for _, r := range results {
		switch {
		case r.Err != nil:
			errored++
		case r.DryRun:
			dryRun++
		case r.Skipped:
			skipped++
		default:
			promoted++
		}
	}
	fmt.Fprintf(w, "summary: %d promoted, %d skipped, %d dry-run, %d errored\n", promoted, skipped, dryRun, errored)
}
