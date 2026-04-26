package expire

import (
	"context"
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// Result holds expiry information for a single secret path.
type Result struct {
	Path      string
	Namespace string
	ExpiresAt time.Time
	TTL       time.Duration
	Expired   bool
	Error     error
}

// Checker checks secret lease / metadata expiry across Vault paths.
type Checker struct {
	client    *vaultapi.Client
	namespace string
}

// New returns a Checker for the given Vault client and namespace.
func New(client *vaultapi.Client, namespace string) *Checker {
	return &Checker{client: client, namespace: namespace}
}

// CheckPath reads the KV v2 metadata for path and returns an expiry Result.
func (c *Checker) CheckPath(ctx context.Context, path string) Result {
	result := Result{Path: path, Namespace: c.namespace}

	metaPath := fmt.Sprintf("secret/metadata/%s", path)
	secret, err := c.client.Logical().ReadWithContext(ctx, metaPath)
	if err != nil {
		result.Error = fmt.Errorf("read metadata %s: %w", path, err)
		return result
	}
	if secret == nil || secret.Data == nil {
		result.Error = fmt.Errorf("no metadata found at %s", path)
		return result
	}

	// KV v2 metadata stores delete_version_after as a duration string.
	if raw, ok := secret.Data["delete_version_after"]; ok {
		if s, ok := raw.(string); ok && s != "0s" {
			d, err := time.ParseDuration(s)
			if err == nil {
				// Approximate expiry from created_time of current version.
				result.TTL = d
				if ct := currentVersionCreatedTime(secret.Data); !ct.IsZero() {
					result.ExpiresAt = ct.Add(d)
					result.Expired = time.Now().After(result.ExpiresAt)
				}
			}
		}
	}

	return result
}

// CheckPaths runs CheckPath for every path and returns all results.
func (c *Checker) CheckPaths(ctx context.Context, paths []string) []Result {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		results = append(results, c.CheckPath(ctx, p))
	}
	return results
}

// currentVersionCreatedTime extracts the created_time of the current version
// from KV v2 metadata.
func currentVersionCreatedTime(data map[string]interface{}) time.Time {
	versions, ok := data["versions"].(map[string]interface{})
	if !ok {
		return time.Time{}
	}
	currentRaw, ok := data["current_version"]
	if !ok {
		return time.Time{}
	}
	currentKey := fmt.Sprintf("%v", currentRaw)
	v, ok := versions[currentKey].(map[string]interface{})
	if !ok {
		return time.Time{}
	}
	ts, ok := v["created_time"].(string)
	if !ok {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return time.Time{}
	}
	return t
}
