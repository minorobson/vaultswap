package stamp

import (
	"context"
	"fmt"
	"time"
)

// VaultClient is the subset of vault operations required by the stamper.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
}

// Result holds the outcome of stamping a single path.
type Result struct {
	Path    string
	Stamped bool
	DryRun  bool
	Err     error
}

// Stamper writes a metadata timestamp key into existing secrets.
type Stamper struct {
	client VaultClient
	key    string
	dryRun bool
}

// New creates a Stamper that will write the given key as a timestamp stamp.
func New(client VaultClient, key string, dryRun bool) *Stamper {
	if key == "" {
		key = "_stamped_at"
	}
	return &Stamper{client: client, key: key, dryRun: dryRun}
}

// StampPath reads the secret at path, injects the timestamp key, and writes it back.
func (s *Stamper) StampPath(ctx context.Context, path string) Result {
	existing, err := s.client.ReadSecret(ctx, path)
	if err != nil {
		return Result{Path: path, Err: fmt.Errorf("read: %w", err)}
	}

	updated := make(map[string]interface{}, len(existing)+1)
	for k, v := range existing {
		updated[k] = v
	}
	updated[s.key] = time.Now().UTC().Format(time.RFC3339)

	if s.dryRun {
		return Result{Path: path, Stamped: true, DryRun: true}
	}

	if err := s.client.WriteSecret(ctx, path, updated); err != nil {
		return Result{Path: path, Err: fmt.Errorf("write: %w", err)}
	}
	return Result{Path: path, Stamped: true}
}

// StampPaths stamps each path and returns all results.
func (s *Stamper) StampPaths(ctx context.Context, paths []string) []Result {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		results = append(results, s.StampPath(ctx, p))
	}
	return results
}
