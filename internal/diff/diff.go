package diff

import (
	"fmt"
	"sort"
	"strings"
)

// ChangeType represents the type of change for a secret key.
type ChangeType string

const (
	Added    ChangeType = "added"
	Removed  ChangeType = "removed"
	Modified ChangeType = "modified"
	Unchanged ChangeType = "unchanged"
)

// Change represents a single key-level diff entry.
type Change struct {
	Key      string
	Type     ChangeType
	OldValue string
	NewValue string
}

// Result holds the full diff between two secret maps.
type Result struct {
	Changes []Change
}

// HasChanges returns true if there are any non-unchanged entries.
func (r *Result) HasChanges() bool {
	for _, c := range r.Changes {
		if c.Type != Unchanged {
			return true
		}
	}
	return false
}

// Compare computes the diff between src and dst secret maps.
func Compare(src, dst map[string]interface{}) *Result {
	result := &Result{}
	keys := unionKeys(src, dst)

	for _, k := range keys {
		srcVal, srcOk := src[k]
		dstVal, dstOk := dst[k]

		switch {
		case srcOk && !dstOk:
			result.Changes = append(result.Changes, Change{
				Key: k, Type: Added,
				NewValue: fmt.Sprintf("%v", srcVal),
			})
		case !srcOk && dstOk:
			result.Changes = append(result.Changes, Change{
				Key: k, Type: Removed,
				OldValue: fmt.Sprintf("%v", dstVal),
			})
		default:
			srcStr := fmt.Sprintf("%v", srcVal)
			dstStr := fmt.Sprintf("%v", dstVal)
			if srcStr != dstStr {
				result.Changes = append(result.Changes, Change{
					Key: k, Type: Modified,
					OldValue: dstStr, NewValue: srcStr,
				})
			} else {
				result.Changes = append(result.Changes, Change{
					Key: k, Type: Unchanged,
				})
			}
		}
	}
	return result
}

// Format renders the diff result as a human-readable string.
func Format(r *Result, maskValues bool) string {
	var sb strings.Builder
	for _, c := range r.Changes {
		switch c.Type {
		case Added:
			val := valueOrMask(c.NewValue, maskValues)
			fmt.Fprintf(&sb, "+ %s = %s\n", c.Key, val)
		case Removed:
			fmt.Fprintf(&sb, "- %s\n", c.Key)
		case Modified:
			oldVal := valueOrMask(c.OldValue, maskValues)
			newVal := valueOrMask(c.NewValue, maskValues)
			fmt.Fprintf(&sb, "~ %s: %s -> %s\n", c.Key, oldVal, newVal)
		}
	}
	return sb.String()
}

func valueOrMask(v string, mask bool) string {
	if mask {
		return "***"
	}
	return v
}

func unionKeys(a, b map[string]interface{}) []string {
	seen := make(map[string]struct{})
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
	sort.Strings(keys)
	return keys
}
