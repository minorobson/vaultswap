package pin

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Result holds the outcome of a pin operation on a single path.
type Result struct {
	Path    string
	Version int
	DryRun  bool
	Err     error
}

// Pinner pins a KV v2 secret to a specific version by writing the
// "current_version" field in the secret's metadata.
type Pinner struct {
	client *api.Client
	dryRun bool
}

// New returns a new Pinner.
func New(client *api.Client, dryRun bool) *Pinner {
	return &Pinner{client: client, dryRun: dryRun}
}

// PinPath pins the secret at path to the given version.
func (p *Pinner) PinPath(ctx context.Context, mount, path string, version int) Result {
	res := Result{Path: path, Version: version, DryRun: p.dryRun}

	metaPath := fmt.Sprintf("%s/metadata/%s", mount, path)

	// Verify the version exists before pinning.
	secret, err := p.client.Logical().ReadWithContext(ctx, metaPath)
	if err != nil {
		res.Err = fmt.Errorf("read metadata: %w", err)
		return res
	}
	if secret == nil || secret.Data == nil {
		res.Err = fmt.Errorf("no metadata found at %s", path)
		return res
	}

	versions, ok := secret.Data["versions"].(map[string]interface{})
	if !ok {
		res.Err = fmt.Errorf("unexpected metadata format at %s", path)
		return res
	}
	versionKey := fmt.Sprintf("%d", version)
	if _, exists := versions[versionKey]; !exists {
		res.Err = fmt.Errorf("version %d does not exist at %s", version, path)
		return res
	}

	if p.dryRun {
		return res
	}

	_, err = p.client.Logical().WriteWithContext(ctx, metaPath, map[string]interface{}{
		"current_version": version,
	})
	if err != nil {
		res.Err = fmt.Errorf("write metadata: %w", err)
	}
	return res
}

// PinPaths pins multiple paths and returns all results.
func (p *Pinner) PinPaths(ctx context.Context, mount string, paths []string, version int) []Result {
	results := make([]Result, 0, len(paths))
	for _, path := range paths {
		results = append(results, p.PinPath(ctx, mount, path, version))
	}
	return results
}
