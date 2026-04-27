package tags

import (
	"context"
	"fmt"
	"strings"
)

// VaultClient is the interface for reading and writing Vault secrets.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
}

// Result holds the outcome of a tag operation on a single path.
type Result struct {
	Path    string
	Tags    map[string]string
	DryRun  bool
	Skipped bool
	Err     error
}

// Tagger writes metadata tags into the custom_metadata field of KV v2 secrets.
type Tagger struct {
	client VaultClient
	dryRun bool
}

// New creates a new Tagger.
func New(client VaultClient, dryRun bool) *Tagger {
	return &Tagger{client: client, dryRun: dryRun}
}

// TagPath merges the provided tags into the existing secret data at path,
// storing them under a reserved "_tags" key.
func (t *Tagger) TagPath(ctx context.Context, path string, tags map[string]string) Result {
	existing, err := t.client.ReadSecret(ctx, path)
	if err != nil {
		return Result{Path: path, Tags: tags, Err: fmt.Errorf("read: %w", err)}
	}

	merged := make(map[string]interface{}, len(existing))
	for k, v := range existing {
		merged[k] = v
	}

	// Encode tags as "_tags.<key>" entries to avoid collisions.
	for k, v := range tags {
		merged["_tags."+k] = v
	}

	if t.dryRun {
		return Result{Path: path, Tags: tags, DryRun: true}
	}

	if err := t.client.WriteSecret(ctx, path, merged); err != nil {
		return Result{Path: path, Tags: tags, Err: fmt.Errorf("write: %w", err)}
	}

	return Result{Path: path, Tags: tags}
}

// TagPaths applies tags to multiple paths and returns all results.
func (t *Tagger) TagPaths(ctx context.Context, paths []string, tags map[string]string) []Result {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		results = append(results, t.TagPath(ctx, strings.TrimSpace(p), tags))
	}
	return results
}
