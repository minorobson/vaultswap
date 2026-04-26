package expire

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Result holds the expiry check result for a single secret path.
type Result struct {
	Path      string
	TTL       time.Duration
	CreatedAt time.Time
	ExpiresAt time.Time
	Expired   bool
	NoTTL     bool
	Err       error
}

// PrintResults writes expiry results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes expiry results to the given writer.
func FprintResults(w io.Writer, results []Result) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no paths checked")
		return
	}
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "[ERROR]   %s: %v\n", r.Path, r.Err)
			continue
		}
		if r.NoTTL {
			fmt.Fprintf(w, "[NO TTL]  %s: no expiry metadata\n", r.Path)
			continue
		}
		if r.Expired {
			fmt.Fprintf(w, "[EXPIRED] %s: expired at %s (ttl: %s)\n",
				r.Path, r.ExpiresAt.Format(time.RFC3339), r.TTL)
			continue
		}
		remaining := time.Until(r.ExpiresAt).Round(time.Second)
		fmt.Fprintf(w, "[OK]      %s: expires at %s (in %s)\n",
			r.Path, r.ExpiresAt.Format(time.RFC3339), remaining)
	}
}
