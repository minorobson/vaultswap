package search

import (
	"context"
	"strings"
)

// VaultClient defines the interface for reading secrets from Vault.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
	ListSecrets(ctx context.Context, path string) ([]string, error)
}

// Result holds a single search match.
type Result struct {
	Path  string
	Key   string
	Value string
	Err   error
}

// Searcher searches secrets in Vault for a given query.
type Searcher struct {
	client VaultClient
}

// New creates a new Searcher.
func New(client VaultClient) *Searcher {
	return &Searcher{client: client}
}

// SearchPaths searches multiple paths and returns all matches.
func (s *Searcher) SearchPaths(ctx context.Context, paths []string, query string, keysOnly bool) []Result {
	var results []Result
	for _, p := range paths {
		results = append(results, s.SearchPath(ctx, p, query, keysOnly)...)
	}
	return results
}

// SearchPath reads a secret at path and returns matching key/value pairs.
func (s *Searcher) SearchPath(ctx context.Context, path, query string, keysOnly bool) []Result {
	data, err := s.client.ReadSecret(ctx, path)
	if err != nil {
		return []Result{{Path: path, Err: err}}
	}

	var results []Result
	for k, v := range data {
		val, _ := v.(string)
		keyMatch := strings.Contains(strings.ToLower(k), strings.ToLower(query))
		valMatch := !keysOnly && strings.Contains(strings.ToLower(val), strings.ToLower(query))
		if keyMatch || valMatch {
			results = append(results, Result{
				Path:  path,
				Key:   k,
				Value: val,
			})
		}
	}
	return results
}
