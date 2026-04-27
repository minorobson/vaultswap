package export

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// VaultReader reads a secret from a Vault path.
type VaultReader interface {
	ReadSecret(path string) (map[string]interface{}, error)
}

// Result holds the outcome of a single export operation.
type Result struct {
	Path  string
	Data  map[string]interface{}
	Error error
}

// Exporter writes Vault secrets to a local file.
type Exporter struct {
	client VaultReader
}

// New creates a new Exporter.
func New(client VaultReader) *Exporter {
	return &Exporter{client: client}
}

// ExportPaths reads secrets from each path and returns results.
func (e *Exporter) ExportPaths(paths []string) []Result {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		data, err := e.client.ReadSecret(p)
		results = append(results, Result{Path: p, Data: data, Error: err})
	}
	return results
}

// WriteFile writes results to a file in the given format ("json" or "yaml").
func WriteFile(results []Result, dest, format string) error {
	payload := make(map[string]interface{})
	for _, r := range results {
		if r.Error != nil {
			continue
		}
		payload[r.Path] = r.Data
	}

	var raw []byte
	var err error

	switch format {
	case "yaml":
		raw, err = yaml.Marshal(payload)
	case "json":
		raw, err = json.MarshalIndent(payload, "", "  ")
	default:
		return fmt.Errorf("unsupported format %q: must be json or yaml", format)
	}
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.WriteFile(dest, raw, 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
