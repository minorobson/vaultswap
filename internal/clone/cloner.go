package clone

import (
	"context"
	"fmt"

	"github.com/vaultswap/internal/vault"
)

// Result holds the outcome of cloning a single secret path.
type Result struct {
	Path    string
	Skipped bool
	DryRun  bool
	Err     error
}

// Cloner copies secrets from one Vault namespace/path to another.
type Cloner struct {
	src  *vault.Client
	dst  *vault.Client
	dryRun bool
}

// New creates a new Cloner.
func New(src, dst *vault.Client, dryRun bool) *Cloner {
	return &Cloner{src: src, dst: dst, dryRun: dryRun}
}

// ClonePath reads a secret from src and writes it to dst at the same path.
// If the destination already contains identical data the write is skipped.
func (c *Cloner) ClonePath(ctx context.Context, path string) Result {
	srcData, err := c.src.ReadSecret(ctx, path)
	if err != nil {
		return Result{Path: path, Err: fmt.Errorf("read src: %w", err)}
	}

	dstData, err := c.dst.ReadSecret(ctx, path)
	if err == nil && mapsEqual(srcData, dstData) {
		return Result{Path: path, Skipped: true, DryRun: c.dryRun}
	}

	if c.dryRun {
		return Result{Path: path, DryRun: true}
	}

	if err := c.dst.WriteSecret(ctx, path, srcData); err != nil {
		return Result{Path: path, Err: fmt.Errorf("write dst: %w", err)}
	}

	return Result{Path: path}
}

// ClonePaths clones multiple paths and returns all results.
func (c *Cloner) ClonePaths(ctx context.Context, paths []string) []Result {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		results = append(results, c.ClonePath(ctx, p))
	}
	return results
}

func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || fmt.Sprintf("%v", v) != fmt.Sprintf("%v", bv) {
			return false
		}
	}
	return true
}
