package rollback

import (
	"context"
	"fmt"

	"github.com/vaultswap/internal/audit"
	"github.com/vaultswap/internal/snapshot"
)

// VaultClient defines the subset of Vault operations needed for rollback.
type VaultClient interface {
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
}

// Rollbacker restores secrets from a previously saved snapshot.
type Rollbacker struct {
	client VaultClient
	logger *audit.Logger
	dryRun bool
}

// New creates a new Rollbacker.
func New(client VaultClient, logger *audit.Logger, dryRun bool) *Rollbacker {
	return &Rollbacker{
		client: client,
		logger: logger,
		dryRun: dryRun,
	}
}

// Result holds the outcome of a rollback operation.
type Result struct {
	Path    string
	DryRun  bool
	Restored bool
}

// Rollback restores secrets from the snapshot at the given file path.
func (r *Rollbacker) Rollback(ctx context.Context, snapshotPath string) ([]Result, error) {
	snap, err := snapshot.Load(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("rollback: load snapshot: %w", err)
	}

	var results []Result

	for path, data := range snap.Secrets {
		result := Result{Path: path, DryRun: r.dryRun}

		if !r.dryRun {
			if err := r.client.WriteSecret(ctx, path, data); err != nil {
				return nil, fmt.Errorf("rollback: write %s: %w", path, err)
			}
			result.Restored = true
		}

		r.logger.Log(audit.Entry{
			Operation: "rollback",
			Path:      path,
			DryRun:    r.dryRun,
		})

		results = append(results, result)
	}

	return results, nil
}
