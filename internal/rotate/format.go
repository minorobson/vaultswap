package rotate

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
)

// PrintResult writes a human-readable summary of a rotation Result to w.
func PrintResult(w io.Writer, r *Result) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	mode := "applied"
	if r.DryRun {
		mode = "dry-run"
	}

	fmt.Fprintf(tw, "Path:\t%s\n", r.Path)
	fmt.Fprintf(tw, "Mode:\t%s\n", mode)
	fmt.Fprintf(tw, "Rotated at:\t%s\n", r.RotatedAt.Format(time.RFC3339))
	fmt.Fprintf(tw, "Keys rotated:\t%s\n", strings.Join(r.Keys, ", "))
}
