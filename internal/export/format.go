package export

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes a human-readable summary of export results to stdout.
func PrintResults(results []Result, dest, format string, dryRun bool) {
	FprintResults(os.Stdout, results, dest, format, dryRun)
}

// FprintResults writes a human-readable summary to the given writer.
func FprintResults(w io.Writer, results []Result, dest, format string, dryRun bool) {
	label := "exported"
	if dryRun {
		label = "[dry-run] would export"
	}

	ok := 0
	for _, r := range results {
		if r.Error != nil {
			fmt.Fprintf(w, "  ERROR   %s: %v\n", r.Path, r.Error)
		} else {
			ok++
			fmt.Fprintf(w, "  OK      %s\n", r.Path)
		}
	}

	if dryRun {
		fmt.Fprintf(w, "\n%s %d path(s) to %s (format: %s)\n", label, ok, dest, format)
	} else {
		fmt.Fprintf(w, "\n%s %d path(s) → %s (format: %s)\n", label, ok, dest, format)
	}
}
