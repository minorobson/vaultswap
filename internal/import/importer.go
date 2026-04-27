package importpkg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/your-org/vaultswap/internal/vault"
)

// Result holds the outcome of importing a single secret path.
type Result struct {
	Path    string
	Written bool
	DryRun  bool
	Err     error
}

// Importer reads secrets from a JSON file and writes them into Vault.
type Importer struct {
	client *vault.Client
}

// New returns a new Importer backed by the given Vault client.
func New(client *vault.Client) *Importer {
	return &Importer{client: client}
}

// ImportFile reads a map of path→data from a JSON file and writes each
// entry to Vault. When dryRun is true no writes are performed.
func (im *Importer) ImportFile(ctx context.Context, filePath string, dryRun bool) ([]Result, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open import file: %w", err)
	}
	defer f.Close()

	var payload map[string]map[string]string
	if err := json.NewDecoder(f).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode import file: %w", err)
	}

	var results []Result
	for path, data := range payload {
		r := Result{Path: path, DryRun: dryRun}
		if !dryRun {
			if err := im.client.WriteSecret(ctx, path, toAnyMap(data)); err != nil {
				r.Err = err
			} else {
				r.Written = true
			}
		}
		results = append(results, r)
	}
	return results, nil
}

func toAnyMap(m map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
