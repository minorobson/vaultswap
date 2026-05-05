// Package diff provides utilities for comparing secret maps and formatting
// human-readable change previews with optional value masking.
package diff

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// ChangeType represents the kind of change detected for a secret key.
type ChangeType string

const (
	Added    ChangeType = "added"
	Removed  ChangeType = "removed"
	Modified ChangeType = "modified"
	Unchanged ChangeType = "unchanged"
)

// Change describes a single key-level difference between two secret maps.
type Change struct {
	Key    string
	Type   ChangeType
	OldVal string
	NewVal string
}

// Compare returns the list of changes between two secret maps (old → new).
// Keys present only in old are Removed; keys only in new are Added;
// keys in both with differing values are Modified; equal values are Unchanged.
func Compare(old, new map[string]any) []Change {
	keys := unionKeys(old, new)
	sort.Strings(keys)

	var changes []Change
	for _, k := range keys {
		oldVal, inOld := old[k]
		newVal, inNew := new[k]

		switch {
		case inOld && !inNew:
			changes = append(changes, Change{
				Key:    k,
				Type:   Removed,
				OldVal: fmt.Sprintf("%v", oldVal),
			})
		case !inOld && inNew:
			changes = append(changes, Change{
				Key:    k,
				Type:   Added,
				NewVal: fmt.Sprintf("%v", newVal),
			})
		default:
			ov := fmt.Sprintf("%v", oldVal)
			nv := fmt.Sprintf("%v", newVal)
			if ov != nv {
				changes = append(changes, Change{
					Key:    k,
					Type:   Modified,
					OldVal: ov,
					NewVal: nv,
				})
			} else {
				changes = append(changes, Change{
					Key:    k,
					Type:   Unchanged,
					OldVal: ov,
					NewVal: nv,
				})
			}
		}
	}
	return changes
}

// Format writes a human-readable diff to stdout.
func Format(changes []Change, maskValues bool) {
	FprintDiff(os.Stdout, changes, maskValues)
}

// FprintDiff writes a human-readable diff to w.
func FprintDiff(w io.Writer, changes []Change, maskValues bool) {
	if len(changes) == 0 {
		fmt.Fprintln(w, "  (no changes)")
		return
	}
	for _, c := range changes {
		switch c.Type {
		case Added:
			fmt.Fprintf(w, "  + %-30s %s\n", c.Key, valueOrMask(c.NewVal, maskValues))
		case Removed:
			fmt.Fprintf(w, "  - %-30s %s\n", c.Key, valueOrMask(c.OldVal, maskValues))
		case Modified:
			fmt.Fprintf(w, "  ~ %-30s %s → %s\n",
				c.Key,
				valueOrMask(c.OldVal, maskValues),
				valueOrMask(c.NewVal, maskValues),
			)
		case Unchanged:
			fmt.Fprintf(w, "    %-30s (unchanged)\n", c.Key)
		}
	}
}

// HasChanges returns true if any of the changes are not Unchanged.
func HasChanges(changes []Change) bool {
	for _, c := range changes {
		if c.Type != Unchanged {
			return true
		}
	}
	return false
}

// valueOrMask returns the value or a masked placeholder depending on the flag.
func valueOrMask(v string, mask bool) string {
	if mask {
		return strings.Repeat("*", 8)
	}
	return v
}

// unionKeys returns the deduplicated union of keys from both maps.
func unionKeys(a, b map[string]any) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	return keys
}
