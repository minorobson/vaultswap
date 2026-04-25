package compare

import (
	"fmt"
	"io"
	"os"

	"github.com/vaultswap/internal/diff"
)

// PrintResults writes a human-readable diff summary to stdout.
func PrintResults(results []Result, maskValues bool) {
	FprintResults(os.Stdout, results, maskValues)
}

// FprintResults writes a human-readable diff summary to w.
func FprintResults(w io.Writer, results []Result, maskValues bool) {
	for _, r := range results {
		if !r.HasDiff {
			fmt.Fprintf(w, "[=] %s — no changes\n", r.Path)
			continue
		}
		fmt.Fprintf(w, "[~] %s\n", r.Path)
		formatted := diff.Format(r.Diff, maskValues)
		for _, line := range formatted {
			fmt.Fprintf(w, "    %s\n", line)
		}
	}
}
