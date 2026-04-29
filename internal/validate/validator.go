package validate

import (
	"context"
	"fmt"
	"strings"
)

// Client defines the interface for reading secrets from Vault.
type Client interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
}

// Result holds the outcome of validating a single secret path.
type Result struct {
	Path    string
	Missing []string
	OK      bool
	Err     error
}

// Validator checks that required keys exist at given Vault paths.
type Validator struct {
	client       Client
	requiredKeys []string
}

// New creates a new Validator with the given client and required key names.
func New(client Client, requiredKeys []string) *Validator {
	return &Validator{
		client:       client,
		requiredKeys: requiredKeys,
	}
}

// ValidatePath reads the secret at path and checks that all required keys are present.
func (v *Validator) ValidatePath(ctx context.Context, path string) Result {
	result := Result{Path: path}

	secret, err := v.client.ReadSecret(ctx, path)
	if err != nil {
		result.Err = fmt.Errorf("read %s: %w", path, err)
		return result
	}

	for _, key := range v.requiredKeys {
		if _, ok := secret[key]; !ok {
			result.Missing = append(result.Missing, key)
		}
	}

	result.OK = len(result.Missing) == 0
	return result
}

// ValidatePaths validates multiple paths and returns all results.
func (v *Validator) ValidatePaths(ctx context.Context, paths []string) []Result {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		results = append(results, v.ValidatePath(ctx, strings.TrimSpace(p)))
	}
	return results
}
