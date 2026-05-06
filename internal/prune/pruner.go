package prune

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Result holds the outcome of pruning a single secret path.
type Result struct {
	Path    string
	Pruned  bool
	DryRun  bool
	Skipped bool
	Error   error
}

// Pruner deletes secret paths that have no keys (empty data maps).
type Pruner struct {
	client *api.Client
	dryRun bool
}

// New returns a new Pruner.
func New(client *api.Client, dryRun bool) *Pruner {
	return &Pruner{client: client, dryRun: dryRun}
}

// PrunePath checks a single KV v2 path and deletes it if its data is empty.
func (p *Pruner) PrunePath(ctx context.Context, mount, path string) Result {
	dataPath := fmt.Sprintf("%s/data/%s", mount, path)

	secret, err := p.client.Logical().ReadWithContext(ctx, dataPath)
	if err != nil {
		return Result{Path: path, Error: err}
	}
	if secret == nil || secret.Data == nil {
		return Result{Path: path, Skipped: true}
	}

	data, ok := secret.Data["data"]
	if !ok {
		return Result{Path: path, Skipped: true}
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok || len(dataMap) > 0 {
		return Result{Path: path, Skipped: true}
	}

	if p.dryRun {
		return Result{Path: path, Pruned: true, DryRun: true}
	}

	deletePath := fmt.Sprintf("%s/metadata/%s", mount, path)
	_, err = p.client.Logical().DeleteWithContext(ctx, deletePath)
	if err != nil {
		return Result{Path: path, Error: err}
	}

	return Result{Path: path, Pruned: true}
}

// PrunePaths runs PrunePath over a list of paths and returns all results.
func (p *Pruner) PrunePaths(ctx context.Context, mount string, paths []string) []Result {
	results := make([]Result, 0, len(paths))
	for _, path := range paths {
		results = append(results, p.PrunePath(ctx, mount, path))
	}
	return results
}
