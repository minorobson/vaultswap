package copy

import (
	"context"
	"fmt"
)

// VaultClient defines the operations required for copying secrets.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
}

// Result holds the outcome of a single copy operation.
type Result struct {
	Src    string
	Dst    string
	DryRun bool
	Err    error
}

// Copier copies secrets from one path to another.
type Copier struct {
	client VaultClient
	dryRun bool
}

// New creates a new Copier.
func New(client VaultClient, dryRun bool) *Copier {
	return &Copier{client: client, dryRun: dryRun}
}

// CopyPath reads a secret from src and writes it to dst.
func (c *Copier) CopyPath(ctx context.Context, src, dst string) Result {
	res := Result{Src: src, Dst: dst, DryRun: c.dryRun}

	data, err := c.client.ReadSecret(ctx, src)
	if err != nil {
		res.Err = fmt.Errorf("read %q: %w", src, err)
		return res
	}

	if c.dryRun {
		return res
	}

	if err := c.client.WriteSecret(ctx, dst, data); err != nil {
		res.Err = fmt.Errorf("write %q: %w", dst, err)
	}
	return res
}

// CopyPaths copies multiple src→dst pairs and returns all results.
func (c *Copier) CopyPaths(ctx context.Context, pairs [][2]string) []Result {
	results := make([]Result, 0, len(pairs))
	for _, p := range pairs {
		results = append(results, c.CopyPath(ctx, p[0], p[1]))
	}
	return results
}
