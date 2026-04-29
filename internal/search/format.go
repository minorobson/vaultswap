package search

import (
	"fmt"
	"io"
	"os"
	"sort"
)

// PrintResults writes search results to stdout.
func PrintResults(results []Result, maskValues bool) {
	FprintResults(os.Stdout, results, maskValues)
}

// FprintResults writes search results to the given writer.
func FprintResults(w io.Writer, results []Result, maskValues bool) {
	if len(results) == 0 {
		fmt.Fprintln(w, "no matches found")
		return
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Path != results[j].Path {
			return results[i].Path < results[j].Path
		}
		return results[i].Key < results[j].Key
	})

	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(w, "  [error] %s: %v\n", r.Path, r.Err)
			continue
		}
		val := r.Value
		if maskValues {
			val = "***"
		}
		fmt.Fprintf(w, "  [match] %s  %s=%s\n", r.Path, r.Key, val)
	}
}
