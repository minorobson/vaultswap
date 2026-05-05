package patch

import (
	"fmt"
	"io"
	"os"
)

// PrintResults writes patch results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes patch results to w.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no patches applied")
		return
	}

	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "  [error]   %s#%s: %v\n", r.Path, r.Key, r.Err)
			continue
		}

		label := "patched"
		if r.DryRun {
			label = "dry-run"
		}

		if r.OldValue == "" {
			fmt.Fprintf(w, "  [%s]  %s#%s  (new key)\n", label, r.Path, r.Key)
		} else {
			fmt.Fprintf(w, "  [%s]  %s#%s  %q -> %q\n", label, r.Path, r.Key, r.OldValue, r.NewValue)
		}
	}
}
