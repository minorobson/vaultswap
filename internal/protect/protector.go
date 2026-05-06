package protect

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Result holds the outcome of a protect operation on a single path.
type Result struct {
	Path    string
	DryRun  bool
	Skipped bool
	Error   error
}

// Protector marks Vault KV v2 secrets as non-deletable by setting
// the "delete_version_after" metadata field to a sentinel value,
// or by writing a custom-metadata key that downstream tooling can
// honour as a protection flag.
type Protector struct {
	client *api.Client
	dryRun bool
}

// New creates a new Protector.
func New(client *api.Client, dryRun bool) *Protector {
	return &Protector{client: client, dryRun: dryRun}
}

// ProtectPath sets the "protected" custom-metadata key on the given
// KV v2 path so that other vaultswap commands skip destructive ops.
func (p *Protector) ProtectPath(ctx context.Context, mount, path string) Result {
	r := Result{Path: path, DryRun: p.dryRun}

	metaPath := fmt.Sprintf("%s/metadata/%s", mount, path)

	// Read existing metadata to detect already-protected paths.
	existing, err := p.client.Logical().ReadWithContext(ctx, metaPath)
	if err != nil {
		r.Error = fmt.Errorf("read metadata %s: %w", path, err)
		return r
	}
	if existing != nil {
		if cm, ok := existing.Data["custom_metadata"].(map[string]interface{}); ok {
			if cm["protected"] == "true" {
				r.Skipped = true
				return r
			}
		}
	}

	if p.dryRun {
		return r
	}

	_, err = p.client.Logical().WriteWithContext(ctx, metaPath, map[string]interface{}{
		"custom_metadata": map[string]interface{}{
			"protected": "true",
		},
	})
	if err != nil {
		r.Error = fmt.Errorf("write metadata %s: %w", path, err)
	}
	return r
}

// ProtectPaths runs ProtectPath for every entry in paths.
func (p *Protector) ProtectPaths(ctx context.Context, mount string, paths []string) []Result {
	results := make([]Result, 0, len(paths))
	for _, path := range paths {
		results = append(results, p.ProtectPath(ctx, mount, path))
	}
	return results
}
