package sync

import (
	"context"
	"fmt"

	"github.com/vaultswap/internal/diff"
	"github.com/vaultswap/internal/vault"
)

// Options configures a sync operation.
type Options struct {
	DryRun     bool
	MaskValues bool
	Force      bool
}

// Result holds the outcome of a sync operation.
type Result struct {
	Path    string
	Changes []diff.Change
	Skipped bool
}

// Syncer copies secrets from a source Vault path to a destination path.
type Syncer struct {
	src  *vault.Client
	dst  *vault.Client
	opts Options
}

// New creates a new Syncer.
func New(src, dst *vault.Client, opts Options) *Syncer {
	return &Syncer{src: src, dst: dst, opts: opts}
}

// Sync reads the secret at srcPath, computes a diff against dstPath,
// prints a preview, and writes the secret unless DryRun is set.
func (s *Syncer) Sync(ctx context.Context, srcPath, dstPath string) (Result, error) {
	srcData, err := s.src.ReadSecret(ctx, srcPath)
	if err != nil {
		return Result{}, fmt.Errorf("read source %q: %w", srcPath, err)
	}

	dstData, err := s.dst.ReadSecret(ctx, dstPath)
	if err != nil {
		// Treat missing destination as empty — a full add.
		dstData = map[string]string{}
	}

	changes := diff.Compare(srcData, dstData)

	result := Result{
		Path:    dstPath,
		Changes: changes,
	}

	if len(changes) == 0 && !s.opts.Force {
		result.Skipped = true
		return result, nil
	}

	if s.opts.DryRun {
		return result, nil
	}

	if err := s.dst.WriteSecret(ctx, dstPath, srcData); err != nil {
		return result, fmt.Errorf("write destination %q: %w", dstPath, err)
	}

	return result, nil
}
