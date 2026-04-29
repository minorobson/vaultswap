package rename

import (
	"context"
	"fmt"
)

// VaultClient defines the methods required by the renamer.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
	DeleteSecret(ctx context.Context, path string) error
}

// Result holds the outcome of a single rename operation.
type Result struct {
	Src    string
	Dst    string
	DryRun bool
	Err    error
}

// Renamer moves secrets from one path to another.
type Renamer struct {
	client VaultClient
	dryRun bool
}

// New creates a new Renamer.
func New(client VaultClient, dryRun bool) *Renamer {
	return &Renamer{client: client, dryRun: dryRun}
}

// RenamePath reads src, writes to dst, then deletes src.
func (r *Renamer) RenamePath(ctx context.Context, src, dst string) Result {
	res := Result{Src: src, Dst: dst, DryRun: r.dryRun}

	data, err := r.client.ReadSecret(ctx, src)
	if err != nil {
		res.Err = fmt.Errorf("read %s: %w", src, err)
		return res
	}

	if r.dryRun {
		return res
	}

	if err := r.client.WriteSecret(ctx, dst, data); err != nil {
		res.Err = fmt.Errorf("write %s: %w", dst, err)
		return res
	}

	if err := r.client.DeleteSecret(ctx, src); err != nil {
		res.Err = fmt.Errorf("delete %s: %w", src, err)
		return res
	}

	return res
}

// RenamePaths renames multiple src→dst pairs.
func (r *Renamer) RenamePaths(ctx context.Context, pairs [][2]string) []Result {
	results := make([]Result, 0, len(pairs))
	for _, p := range pairs {
		results = append(results, r.RenamePath(ctx, p[0], p[1]))
	}
	return results
}
