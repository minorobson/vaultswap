package namespace

import "strings"

// FilterOptions holds configuration for namespace filtering.
type FilterOptions struct {
	Prefix    string
	Substring string
	Exclude   []string
}

// ApplyFilters returns namespaces matching the given FilterOptions.
func ApplyFilters(namespaces []string, opts FilterOptions) []string {
	excludeSet := make(map[string]struct{}, len(opts.Exclude))
	for _, e := range opts.Exclude {
		excludeSet[strings.TrimSpace(e)] = struct{}{}
	}

	var result []string
	for _, ns := range namespaces {
		if _, excluded := excludeSet[ns]; excluded {
			continue
		}
		if opts.Prefix != "" && !strings.HasPrefix(ns, opts.Prefix) {
			continue
		}
		if opts.Substring != "" && !strings.Contains(ns, opts.Substring) {
			continue
		}
		result = append(result, ns)
	}
	return result
}
