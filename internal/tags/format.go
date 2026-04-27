package tags

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Result holds the outcome of a tag operation on a single path.
type Result struct {
	Path    string
	Tags    map[string]string
	DryRun  bool
	Skipped bool
	Err     error
}

// PrintResults writes tag results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes tag results to the given writer.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths processed")
		return
	}

	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "[error]   %s — %v\n", r.Path, r.Err)
			continue
		}
		if r.Skipped {
			fmt.Fprintf(w, "[skipped] %s\n", r.Path)
			continue
		}

		label := "[tagged]"
		if r.DryRun {
			label = "[dry-run]"
		}

		pairs := make([]string, 0, len(r.Tags))
		for k, v := range r.Tags {
			pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
		}
		fmt.Fprintf(w, "%s %s — %s\n", label, r.Path, strings.Join(pairs, ", "))
	}
}
