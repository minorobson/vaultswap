package mirror

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Result holds the outcome of mirroring a single path.
type Result struct {
	SourcePath string
	DestPath   string
	Skipped    bool
	DryRun     bool
	Err        error
}

// Mirrorer copies secrets from a source namespace to a destination namespace,
// overwriting any existing values at the destination.
type Mirrorer struct {
	src    *api.Client
	dst    *api.Client
	dryRun bool
}

// New returns a Mirrorer that reads from src and writes to dst.
func New(src, dst *api.Client, dryRun bool) *Mirrorer {
	return &Mirrorer{src: src, dst: dst, dryRun: dryRun}
}

// MirrorPath mirrors a single KV v2 path from source to destination.
func (m *Mirrorer) MirrorPath(ctx context.Context, srcPath, dstPath string) Result {
	res := Result{SourcePath: srcPath, DestPath: dstPath, DryRun: m.dryRun}

	secret, err := m.src.KVv2("").Get(ctx, srcPath)
	if err != nil {
		res.Err = fmt.Errorf("read %s: %w", srcPath, err)
		return res
	}
	if secret == nil || secret.Data == nil {
		res.Skipped = true
		return res
	}

	if m.dryRun {
		return res
	}

	if _, err := m.dst.KVv2("").Put(ctx, dstPath, secret.Data); err != nil {
		res.Err = fmt.Errorf("write %s: %w", dstPath, err)
	}
	return res
}

// MirrorPaths mirrors multiple paths and returns all results.
func (m *Mirrorer) MirrorPaths(ctx context.Context, pairs [][2]string) []Result {
	results := make([]Result, 0, len(pairs))
	for _, p := range pairs {
		results = append(results, m.MirrorPath(ctx, p[0], p[1]))
	}
	return results
}
