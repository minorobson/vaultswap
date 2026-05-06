package redact

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// PrintResults writes a human-readable summary of redaction results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes a human-readable summary of redaction results to w.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths processed")
		return
	}
	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(w, "  ERROR   %s: %v\n", r.Path, r.Err)
		case len(r.Keys) == 0:
			fmt.Fprintf(w, "  SKIP    %s (no matching values)\n", r.Path)
		case r.DryRun:
			sort.Strings(r.Keys)
			fmt.Fprintf(w, "  DRY-RUN %s [%s]\n", r.Path, strings.Join(r.Keys, ", "))
		default:
			sort.Strings(r.Keys)
			fmt.Fprintf(w, "  REDACTED %s [%s]\n", r.Path, strings.Join(r.Keys, ", "))
		}
	}
}
