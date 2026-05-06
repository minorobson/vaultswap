package redact

import (
	"fmt"
	"regexp"

	"github.com/your-org/vaultswap/internal/vault"
)

// Result holds the outcome of redacting a single secret path.
type Result struct {
	Path    string
	Keys    []string
	DryRun  bool
	Err     error
}

// Redactor replaces secret values matching a pattern with a placeholder.
type Redactor struct {
	client    *vault.Client
	pattern   *regexp.Regexp
	placeholder string
	dryRun    bool
}

// New creates a Redactor that replaces values matching pattern with placeholder.
func New(client *vault.Client, pattern, placeholder string, dryRun bool) (*Redactor, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}
	return &Redactor{
		client:      client,
		pattern:     re,
		placeholder: placeholder,
		dryRun:      dryRun,
	}, nil
}

// RedactPath reads the secret at path, replaces matching values, and writes back.
func (r *Redactor) RedactPath(path string) Result {
	secret, err := r.client.ReadSecret(path)
	if err != nil {
		return Result{Path: path, Err: err}
	}

	updated := make(map[string]interface{}, len(secret))
	var matched []string
	for k, v := range secret {
		str, ok := v.(string)
		if ok && r.pattern.MatchString(str) {
			updated[k] = r.placeholder
			matched = append(matched, k)
		} else {
			updated[k] = v
		}
	}

	if len(matched) == 0 {
		return Result{Path: path, Keys: nil, DryRun: r.dryRun}
	}

	if !r.dryRun {
		if err := r.client.WriteSecret(path, updated); err != nil {
			return Result{Path: path, Err: err}
		}
	}

	return Result{Path: path, Keys: matched, DryRun: r.dryRun}
}

// RedactPaths runs RedactPath over each path and returns all results.
func (r *Redactor) RedactPaths(paths []string) []Result {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		results = append(results, r.RedactPath(p))
	}
	return results
}
