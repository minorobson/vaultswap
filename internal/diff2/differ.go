package diff2

import (
	"fmt"
	"sort"
	"strings"
)

// Result holds the diff output for a single secret path.
type Result struct {
	Path    string
	Added   map[string]string
	Removed map[string]string
	Changed map[string][2]string // key -> [old, new]
	Error   error
}

// HasChanges returns true if there are any differences.
func (r Result) HasChanges() bool {
	return len(r.Added) > 0 || len(r.Removed) > 0 || len(r.Changed) > 0
}

// Comparer diffs two maps of secret values.
type Comparer struct {
	MaskValues bool
}

// New returns a new Comparer.
func New(maskValues bool) *Comparer {
	return &Comparer{MaskValues: maskValues}
}

// Compare diffs src against dst for the given path.
func (c *Comparer) Compare(path string, src, dst map[string]any) Result {
	res := Result{
		Path:    path,
		Added:   make(map[string]string),
		Removed: make(map[string]string),
		Changed: make(map[string][2]string),
	}

	for k, sv := range src {
		svStr := fmt.Sprintf("%v", sv)
		if dv, ok := dst[k]; !ok {
			res.Removed[k] = c.mask(svStr)
		} else {
			dvStr := fmt.Sprintf("%v", dv)
			if svStr != dvStr {
				res.Changed[k] = [2]string{c.mask(svStr), c.mask(dvStr)}
			}
		}
	}
	for k, dv := range dst {
		if _, ok := src[k]; !ok {
			res.Added[k] = c.mask(fmt.Sprintf("%v", dv))
		}
	}
	return res
}

func (c *Comparer) mask(v string) string {
	if c.MaskValues {
		return strings.Repeat("*", len(v))
	}
	return v
}

// SortedKeys returns the sorted keys of a string map.
func SortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
