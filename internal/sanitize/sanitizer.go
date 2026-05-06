package sanitize

import (
	"fmt"
	"strings"

	"github.com/your-org/vaultswap/internal/vault"
)

// Result holds the outcome of sanitizing a single secret path.
type Result struct {
	Path    string
	Key     string
	Removed bool
	DryRun  bool
	Err     error
}

// Sanitizer removes keys whose values match a set of disallowed patterns.
type Sanitizer struct {
	client   *vault.Client
	patterns []string
	dryRun   bool
}

// New returns a Sanitizer that will strip keys whose trimmed values match any
// of the provided patterns (case-insensitive).
func New(client *vault.Client, patterns []string, dryRun bool) *Sanitizer {
	lower := make([]string, len(patterns))
	for i, p := range patterns {
		lower[i] = strings.ToLower(strings.TrimSpace(p))
	}
	return &Sanitizer{client: client, patterns: lower, dryRun: dryRun}
}

// SanitizePath reads the secret at path and removes any keys whose values
// match the disallowed patterns. Returns one Result per matched key.
func (s *Sanitizer) SanitizePath(path string) ([]Result, error) {
	data, err := s.client.ReadSecret(path)
	if err != nil {
		return []Result{{Path: path, Err: fmt.Errorf("read: %w", err)}}, nil
	}

	updated := make(map[string]interface{}, len(data))
	for k, v := range data {
		updated[k] = v
	}

	var results []Result
	for k, v := range data {
		val := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", v)))
		if s.matches(val) {
			results = append(results, Result{
				Path:    path,
				Key:     k,
				Removed: !s.dryRun,
				DryRun:  s.dryRun,
			})
			if !s.dryRun {
				delete(updated, k)
			}
		}
	}

	if !s.dryRun && len(results) > 0 {
		if err := s.client.WriteSecret(path, updated); err != nil {
			return []Result{{Path: path, Err: fmt.Errorf("write: %w", err)}}, nil
		}
	}

	return results, nil
}

// SanitizePaths runs SanitizePath across every provided path.
func (s *Sanitizer) SanitizePaths(paths []string) []Result {
	var all []Result
	for _, p := range paths {
		res, _ := s.SanitizePath(p)
		all = append(all, res...)
	}
	return all
}

func (s *Sanitizer) matches(val string) bool {
	for _, p := range s.patterns {
		if val == p {
			return true
		}
	}
	return false
}
