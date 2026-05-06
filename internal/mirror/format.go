package mirror

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes mirror results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes mirror results to w.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths to mirror")
		return
	}
	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "[error]   %s -> %s: %v\n", r.SourcePath, r.DestPath, r.Err)
		case r.Skipped:
			fmt.Fprintf(w, "[skipped] %s -> %s (no data)\n", r.SourcePath, r.DestPath)
		case r.DryRun:
			fmt.Fprintf(w, "[dry-run] %s -> %s\n", r.SourcePath, r.DestPath)
		default:
			fmt.Fprintf(w, "[mirrored] %s -> %s\n", r.SourcePath, r.DestPath)
		}
	}
}
