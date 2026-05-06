package touch

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Result holds the outcome of a touch operation on a single path.
type Result struct {
	Path    string
	Touched bool
	DryRun  bool
	Err     error
}

// Toucher re-writes a secret at a path with its existing data, bumping
// the KV v2 version without changing any values. This is useful for
// resetting TTL clocks or triggering audit events.
type Toucher struct {
	client *api.Client
	dryRun bool
}

// New returns a Toucher backed by the provided Vault client.
func New(client *api.Client, dryRun bool) *Toucher {
	return &Toucher{client: client, dryRun: dryRun}
}

// TouchPath reads the current secret at path and immediately re-writes it,
// creating a new version with identical data.
func (t *Toucher) TouchPath(ctx context.Context, path string) Result {
	secret, err := t.client.KVv2("").Get(ctx, path)
	if err != nil {
		return Result{Path: path, Err: fmt.Errorf("read %q: %w", path, err)}
	}
	if secret == nil || secret.Data == nil {
		return Result{Path: path, Err: fmt.Errorf("path %q returned no data", path)}
	}

	if t.dryRun {
		return Result{Path: path, Touched: true, DryRun: true}
	}

	_, err = t.client.KVv2("").Put(ctx, path, secret.Data)
	if err != nil {
		return Result{Path: path, Err: fmt.Errorf("write %q: %w", path, err)}
	}
	return Result{Path: path, Touched: true}
}

// TouchPaths runs TouchPath for every entry in paths and returns all results.
func (t *Toucher) TouchPaths(ctx context.Context, paths []string) []Result {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		results = append(results, t.TouchPath(ctx, p))
	}
	return results
}
