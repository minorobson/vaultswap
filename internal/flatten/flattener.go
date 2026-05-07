package flatten

import (
	"context"
	"fmt"
	"strings"
)

// VaultClient is the interface required by Flattener.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
}

// Result holds the outcome of a flatten operation on a single path.
type Result struct {
	Path    string
	Keys    []string
	Skipped bool
	DryRun  bool
	Err     error
}

// Flattener collapses nested map keys into dot-separated top-level keys.
type Flattener struct {
	client VaultClient
	dryRun bool
}

// New returns a new Flattener.
func New(client VaultClient, dryRun bool) *Flattener {
	return &Flattener{client: client, dryRun: dryRun}
}

// FlattenPath reads the secret at path, flattens any nested map values into
// dot-separated keys, and writes the result back (unless dry-run).
func (f *Flattener) FlattenPath(ctx context.Context, path string) Result {
	data, err := f.client.ReadSecret(ctx, path)
	if err != nil {
		return Result{Path: path, Err: fmt.Errorf("read: %w", err)}
	}

	flat := make(map[string]interface{})
	flattenMap("", data, flat)

	if mapsEqual(data, flat) {
		return Result{Path: path, Skipped: true}
	}

	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}

	if f.dryRun {
		return Result{Path: path, Keys: keys, DryRun: true}
	}

	if err := f.client.WriteSecret(ctx, path, flat); err != nil {
		return Result{Path: path, Err: fmt.Errorf("write: %w", err)}
	}

	return Result{Path: path, Keys: keys}
}

// FlattenPaths runs FlattenPath over each provided path.
func (f *Flattener) FlattenPaths(ctx context.Context, paths []string) []Result {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		results = append(results, f.FlattenPath(ctx, p))
	}
	return results
}

func flattenMap(prefix string, src map[string]interface{}, dst map[string]interface{}) {
	for k, v := range src {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if nested, ok := v.(map[string]interface{}); ok {
			flattenMap(key, nested, dst)
		} else {
			dst[key] = v
		}
	}
}

func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		if fmt.Sprintf("%v", av) != fmt.Sprintf("%v", bv) {
			return false
		}
	}
	return true
}

// keysString is a helper used in format output.
func keysString(keys []string) string {
	return strings.Join(keys, ", ")
}
