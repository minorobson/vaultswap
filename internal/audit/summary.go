package audit

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

// OperationSummary holds aggregated stats for a vaultswap operation.
type OperationSummary struct {
	Operation string
	Namespace string
	Total     int
	Succeeded int
	Failed    int
	Skipped   int
	DryRun    bool
	StartedAt time.Time
	Duration  time.Duration
}

// PrintSummary writes a formatted summary table to stdout.
func PrintSummary(s OperationSummary) {
	FprintSummary(os.Stdout, s)
}

// FprintSummary writes a formatted summary table to the given writer.
func FprintSummary(w io.Writer, s OperationSummary) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	label := s.Operation
	if s.DryRun {
		label += " (dry-run)"
	}

	fmt.Fprintf(tw, "Operation:\t%s\n", label)
	if s.Namespace != "" {
		fmt.Fprintf(tw, "Namespace:\t%s\n", s.Namespace)
	}
	fmt.Fprintf(tw, "Started:\t%s\n", s.StartedAt.Format(time.RFC3339))
	fmt.Fprintf(tw, "Duration:\t%s\n", s.Duration.Round(time.Millisecond))
	fmt.Fprintf(tw, "Total:\t%d\n", s.Total)
	fmt.Fprintf(tw, "Succeeded:\t%d\n", s.Succeeded)
	fmt.Fprintf(tw, "Failed:\t%d\n", s.Failed)
	fmt.Fprintf(tw, "Skipped:\t%d\n", s.Skipped)

	tw.Flush()
}
