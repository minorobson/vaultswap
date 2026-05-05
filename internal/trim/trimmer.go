package trim

import (
	"context"
	"fmt"

	"github.com/your-org/vaultswap/internal/vault"
)

// Result holds the outcome of a trim operation on a single path.
type Result struct {
	Path    string
	Key     string
	Removed bool
	DryRun  bool
	Err     error
}

// Trimmer removes specified keys from secrets at given paths.
type Trimmer struct {
	client *vault.Client
	dryRun bool
}

// New returns a new Trimmer.
func New(client *vault.Client, dryRun bool) *Trimmer {
	return &Trimmer{client: client, dryRun: dryRun}
}

// TrimKey removes a single key from the secret at path.
func (t *Trimmer) TrimKey(ctx context.Context, path, key string) Result {
	result := Result{Path: path, Key: key, DryRun: t.dryRun}

	data, err := t.client.ReadSecret(ctx, path)
	if err != nil {
		result.Err = fmt.Errorf("read %s: %w", path, err)
		return result
	}

	if _, exists := data[key]; !exists {
		return result
	}

	if t.dryRun {
		result.Removed = true
		return result
	}

	delete(data, key)

	if err := t.client.WriteSecret(ctx, path, data); err != nil {
		result.Err = fmt.Errorf("write %s: %w", path, err)
		return result
	}

	result.Removed = true
	return result
}

// TrimKeys removes multiple keys from the secret at path.
func (t *Trimmer) TrimKeys(ctx context.Context, path string, keys []string) []Result {
	results := make([]Result, 0, len(keys))
	for _, key := range keys {
		results = append(results, t.TrimKey(ctx, path, key))
	}
	return results
}
