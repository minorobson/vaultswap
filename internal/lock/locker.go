package lock

import (
	"context"
	"fmt"
	"time"
)

// VaultClient defines the subset of vault operations needed by the locker.
type VaultClient interface {
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
	DeleteSecret(ctx context.Context, path string) error
}

// Result holds the outcome of a lock or unlock operation on a single path.
type Result struct {
	Path   string
	Action string // "locked", "unlocked", "skipped", "error"
	DryRun bool
	Err    error
}

// Locker manages advisory locks stored as Vault secrets.
type Locker struct {
	client VaultClient
	lockPath string
	dryRun   bool
}

// New creates a new Locker. lockPath is the KV path used to store lock metadata.
func New(client VaultClient, lockPath string, dryRun bool) *Locker {
	return &Locker{
		client:   client,
		lockPath: lockPath,
		dryRun:   dryRun,
	}
}

// Lock writes an advisory lock record at the configured path.
// If a lock already exists it returns a Result with action "skipped".
func (l *Locker) Lock(ctx context.Context, owner string) Result {
	existing, err := l.client.ReadSecret(ctx, l.lockPath)
	if err == nil && existing != nil {
		return Result{Path: l.lockPath, Action: "skipped", DryRun: l.dryRun}
	}

	if l.dryRun {
		return Result{Path: l.lockPath, Action: "locked", DryRun: true}
	}

	payload := map[string]interface{}{
		"owner":      owner,
		"locked_at":  time.Now().UTC().Format(time.RFC3339),
	}
	if err := l.client.WriteSecret(ctx, l.lockPath, payload); err != nil {
		return Result{Path: l.lockPath, Action: "error", DryRun: false, Err: fmt.Errorf("lock write failed: %w", err)}
	}
	return Result{Path: l.lockPath, Action: "locked", DryRun: false}
}

// Unlock removes the advisory lock at the configured path.
// If no lock exists it returns a Result with action "skipped".
func (l *Locker) Unlock(ctx context.Context) Result {
	existing, err := l.client.ReadSecret(ctx, l.lockPath)
	if err != nil || existing == nil {
		return Result{Path: l.lockPath, Action: "skipped", DryRun: l.dryRun}
	}

	if l.dryRun {
		return Result{Path: l.lockPath, Action: "unlocked", DryRun: true}
	}

	if err := l.client.DeleteSecret(ctx, l.lockPath); err != nil {
		return Result{Path: l.lockPath, Action: "error", DryRun: false, Err: fmt.Errorf("lock delete failed: %w", err)}
	}
	return Result{Path: l.lockPath, Action: "unlocked", DryRun: false}
}
