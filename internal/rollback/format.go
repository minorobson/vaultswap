package rollback

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes a human-readable summary of rollback results.
func PrintResults(results []Result, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}

	if len(results) == 0 {
		fmt.Fprintln(w, "No secrets to restore.")
		return
	}

	label := "LIVE"
	if len(results) > 0 && results[0].DryRun {
		label = "DRY-RUN"
	}

	fmt.Fprintf(w, "Rollback [%s]\n", label)
	fmt.Fprintf(w, "%-6s  %s\n", "STATUS", "PATH")
	fmt.Fprintln(w, "------  ----")

	for _, r := range results {
		status := "skipped"
		if r.Restored {
			status = "restored"
		} else if r.DryRun {
			status = "would restore"
		}
		fmt.Fprintf(w, "%-13s  %s\n", status, r.Path)
	}
}
