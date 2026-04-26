package promote

import (
	"context"
	"fmt"

	"github.com/your-org/vaultswap/internal/vault"
)

// Result holds the outcome of a single promotion operation.
type Result struct {
	Path    string
	Src     string
	Dst     string
	Skipped bool
	DryRun  bool
	Err     error
}

// Promoter copies secrets from one namespace to another.
type Promoter struct {
	src    *vault.Client
	dst    *vault.Client
	dryRun bool
}

// New creates a new Promoter.
func New(src, dst *vault.Client, dryRun bool) *Promoter {
	return &Promoter{src: src, dst: dst, dryRun: dryRun}
}

// Promote reads each path from src and writes it to dst.
func (p *Promoter) Promote(ctx context.Context, paths []string) ([]Result, error) {
	results := make([]Result, 0, len(paths))
	for _, path := range paths {
		r := p.promotePath(ctx, path)
		results = append(results, r)
	}
	return results, nil
}

func (p *Promoter) promotePath(ctx context.Context, path string) Result {
	r := Result{
		Path:   path,
		Src:    p.src.Namespace(),
		Dst:    p.dst.Namespace(),
		DryRun: p.dryRun,
	}

	data, err := p.src.ReadSecret(ctx, path)
	if err != nil {
		r.Err = fmt.Errorf("read src %q: %w", path, err)
		return r
	}

	if p.dryRun {
		r.Skipped = true
		return r
	}

	if err := p.dst.WriteSecret(ctx, path, data); err != nil {
		r.Err = fmt.Errorf("write dst %q: %w", path, err)
		return r
	}

	return r
}
