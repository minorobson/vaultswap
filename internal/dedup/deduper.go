package dedup

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Result holds the outcome of a deduplication check for a single path.
type Result struct {
	Path        string
	Duplicates  []string
	Removed     int
	Skipped     bool
	DryRun      bool
	Err         error
}

// Deduper checks for and removes duplicate key values within a secret.
type Deduper struct {
	client *api.Client
	dryRun bool
}

// New returns a new Deduper.
func New(client *api.Client, dryRun bool) *Deduper {
	return &Deduper{client: client, dryRun: dryRun}
}

// DeduplicatePath reads the secret at path, identifies keys that share an
// identical value, retains the first occurrence, and removes the rest.
func (d *Deduper) DeduplicatePath(ctx context.Context, path string) Result {
	secret, err := d.client.KVv2("secret").Get(ctx, path)
	if err != nil {
		return Result{Path: path, Err: fmt.Errorf("read: %w", err)}
	}
	if secret == nil || secret.Data == nil {
		return Result{Path: path, Skipped: true}
	}

	seen := make(map[string]string) // value -> first key
	var dupes []string
	clean := make(map[string]interface{})

	for k, v := range secret.Data {
		str := fmt.Sprintf("%v", v)
		if first, ok := seen[str]; ok {
			_ = first
			dupes = append(dupes, k)
		} else {
			seen[str] = k
			clean[k] = v
		}
	}

	if len(dupes) == 0 {
		return Result{Path: path, Skipped: true, DryRun: d.dryRun}
	}

	if !d.dryRun {
		if _, err := d.client.KVv2("secret").Put(ctx, path, clean); err != nil {
			return Result{Path: path, Err: fmt.Errorf("write: %w", err)}
		}
	}

	return Result{
		Path:       path,
		Duplicates: dupes,
		Removed:    len(dupes),
		DryRun:     d.dryRun,
	}
}

// DeduplicatePaths runs DeduplicatePath for each path and returns all results.
func (d *Deduper) DeduplicatePaths(ctx context.Context, paths []string) []Result {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		results = append(results, d.DeduplicatePath(ctx, p))
	}
	return results
}
