package revert

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Result holds the outcome of reverting a single secret path to a prior version.
type Result struct {
	Path    string
	Version int
	DryRun  bool
	Err     error
	Skipped bool
}

// Reverter rolls back individual KV v2 secrets to a specified version.
type Reverter struct {
	client *api.Client
	dryRun bool
}

// New creates a new Reverter.
func New(client *api.Client, dryRun bool) *Reverter {
	return &Reverter{client: client, dryRun: dryRun}
}

// RevertPath rolls the secret at path back to the given version.
func (r *Reverter) RevertPath(ctx context.Context, mount, path string, version int) Result {
	res := Result{Path: path, Version: version, DryRun: r.dryRun}

	// Read the target version first to confirm it exists.
	versioned, err := r.client.KVv2(mount).GetVersion(ctx, path, version)
	if err != nil {
		res.Err = fmt.Errorf("read version %d: %w", version, err)
		return res
	}
	if versioned == nil || versioned.Data == nil {
		res.Err = fmt.Errorf("version %d not found or destroyed", version)
		return res
	}

	if r.dryRun {
		return res
	}

	_, err = r.client.KVv2(mount).Put(ctx, path, versioned.Data)
	if err != nil {
		res.Err = fmt.Errorf("write reverted data: %w", err)
	}
	return res
}

// RevertPaths reverts multiple paths, each to its specified version.
func (r *Reverter) RevertPaths(ctx context.Context, mount string, targets map[string]int) []Result {
	results := make([]Result, 0, len(targets))
	for path, version := range targets {
		results = append(results, r.RevertPath(ctx, mount, path, version))
	}
	return results
}
