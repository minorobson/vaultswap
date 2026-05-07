package prefill

import (
	"context"
	"fmt"
)

// VaultClient is the interface used to read and write KV secrets.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
}

// Result holds the outcome of a single prefill operation.
type Result struct {
	Path    string
	Key     string
	Skipped bool // true when the key already exists
	DryRun  bool
	Err     error
}

// Prefiller writes default values for missing keys in a secret path.
type Prefiller struct {
	client VaultClient
	dryRun bool
}

// New returns a new Prefiller.
func New(client VaultClient, dryRun bool) *Prefiller {
	return &Prefiller{client: client, dryRun: dryRun}
}

// PrefillPath ensures that each key in defaults exists in the secret at path.
// Keys that already have a value are left untouched.
func (p *Prefiller) PrefillPath(ctx context.Context, path string, defaults map[string]string) []Result {
	existing, err := p.client.ReadSecret(ctx, path)
	if err != nil {
		// Treat a missing secret as an empty map so we can write all defaults.
		existing = map[string]interface{}{}
	}

	merged := make(map[string]interface{}, len(existing))
	for k, v := range existing {
		merged[k] = v
	}

	var results []Result
	anyNew := false

	for key, val := range defaults {
		if _, ok := existing[key]; ok {
			results = append(results, Result{Path: path, Key: key, Skipped: true, DryRun: p.dryRun})
			continue
		}
		merged[key] = val
		anyNew = true
		results = append(results, Result{Path: path, Key: key, Skipped: false, DryRun: p.dryRun})
	}

	if anyNew && !p.dryRun {
		if writeErr := p.client.WriteSecret(ctx, path, merged); writeErr != nil {
			for i := range results {
				if !results[i].Skipped {
					results[i].Err = fmt.Errorf("write failed: %w", writeErr)
				}
			}
		}
	}

	return results
}
