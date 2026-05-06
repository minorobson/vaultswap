package diff2

import (
	"fmt"
	"io"
	"os"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// PrintResults writes diff results to stdout.
func PrintResults(results []Result) {
	FprintResults(os.Stdout, results)
}

// FprintResults writes diff results to the given writer.
func FprintResults(w io.Writer, results []Result) {
	for _, r := range results {
		if r.Error != nil {
			fmt.Fprintf(w, "[error] %s: %v\n", r.Path, r.Error)
			continue
		}
		if !r.HasChanges() {
			fmt.Fprintf(w, "[no diff] %s\n", r.Path)
			continue
		}
		fmt.Fprintf(w, "[diff] %s\n", r.Path)
		for _, k := range SortedKeys(r.Added) {
			fmt.Fprintf(w, "  %s+ %s = %s%s\n", colorGreen, k, r.Added[k], colorReset)
		}
		for _, k := range SortedKeys(r.Removed) {
			fmt.Fprintf(w, "  %s- %s = %s%s\n", colorRed, k, r.Removed[k], colorReset)
		}
		for _, k := range SortedKeys(changedKeys(r.Changed)) {
			pair := r.Changed[k]
			fmt.Fprintf(w, "  %s~ %s: %s -> %s%s\n", colorYellow, k, pair[0], pair[1], colorReset)
		}
	}
}

func changedKeys(m map[string][2]string) map[string]string {
	out := make(map[string]string, len(m))
	for k := range m {
		out[k] = ""
	}
	return out
}
