package merge

import (
	"context"
	"fmt"
)

// VaultClient defines the interface required by Merger.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
}

// Result holds the outcome of merging one path.
type Result struct {
	Path    string
	Merged  int
	Skipped int
	DryRun  bool
	Err     error
}

// Options controls merge behaviour.
type Options struct {
	Overwrite bool
	DryRun    bool
}

// Merger reads secrets from one or more source paths and merges them into a
// destination path, optionally overwriting existing keys.
type Merger struct {
	client VaultClient
	opts   Options
}

// New creates a new Merger.
func New(client VaultClient, opts Options) *Merger {
	return &Merger{client: client, opts: opts}
}

// MergePaths merges all source paths into dest and returns one Result per call.
func (m *Merger) MergePaths(ctx context.Context, dest string, sources []string) []Result {
	results := make([]Result, 0, len(sources))
	for _, src := range sources {
		results = append(results, m.MergePath(ctx, dest, src))
	}
	return results
}

// MergePath merges a single source path into dest.
func (m *Merger) MergePath(ctx context.Context, dest, src string) Result {
	r := Result{Path: src, DryRun: m.opts.DryRun}

	srcData, err := m.client.ReadSecret(ctx, src)
	if err != nil {
		r.Err = fmt.Errorf("read source %q: %w", src, err)
		return r
	}

	destData, err := m.client.ReadSecret(ctx, dest)
	if err != nil {
		// Destination may not exist yet; start with an empty map.
		destData = make(map[string]interface{})
	}

	merged := copyMap(destData)
	for k, v := range srcData {
		if _, exists := merged[k]; exists && !m.opts.Overwrite {
			r.Skipped++
			continue
		}
		merged[k] = v
		r.Merged++
	}

	if r.Merged == 0 {
		return r
	}

	if !m.opts.DryRun {
		if err := m.client.WriteSecret(ctx, dest, merged); err != nil {
			r.Err = fmt.Errorf("write dest %q: %w", dest, err)
		}
	}
	return r
}

func copyMap(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
