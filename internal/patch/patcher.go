package patch

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Result holds the outcome of a single patch operation.
type Result struct {
	Path    string
	Key     string
	OldValue string
	NewValue string
	DryRun  bool
	Err     error
}

// VaultClient defines the subset of Vault operations required by Patcher.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (*api.Secret, error)
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
}

// Patcher applies targeted key-level updates to existing Vault secrets
// without overwriting unrelated keys.
type Patcher struct {
	client VaultClient
	dryRun bool
}

// New returns a new Patcher.
func New(client VaultClient, dryRun bool) *Patcher {
	return &Patcher{client: client, dryRun: dryRun}
}

// PatchPath reads the secret at path, merges the provided key/value pairs,
// and writes the result back. Existing keys not present in patches are preserved.
func (p *Patcher) PatchPath(ctx context.Context, path string, patches map[string]string) ([]Result, error) {
	secret, err := p.client.ReadSecret(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	existing := make(map[string]interface{})
	if secret != nil && secret.Data != nil {
		if data, ok := secret.Data["data"].(map[string]interface{}); ok {
			for k, v := range data {
				existing[k] = v
			}
		}
	}

	var results []Result
	for key, newVal := range patches {
		old := ""
		if v, ok := existing[key]; ok {
			old = fmt.Sprintf("%v", v)
		}
		results = append(results, Result{
			Path:     path,
			Key:      key,
			OldValue: old,
			NewValue: newVal,
			DryRun:   p.dryRun,
		})
		existing[key] = newVal
	}

	if !p.dryRun {
		if err := p.client.WriteSecret(ctx, path, existing); err != nil {
			return nil, fmt.Errorf("write %s: %w", path, err)
		}
	}

	return results, nil
}
