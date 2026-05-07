package cascade

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Result holds the outcome of cascading a secret to a single destination path.
type Result struct {
	Source string
	Dest   string
	DryRun bool
	Skipped bool
	Err    error
}

// Cascader propagates a source secret to multiple destination paths.
type Cascader struct {
	client *api.Client
	dryRun bool
}

// New creates a new Cascader.
func New(client *api.Client, dryRun bool) *Cascader {
	return &Cascader{client: client, dryRun: dryRun}
}

// CascadePath reads the secret at src and writes it to each path in dests.
func (c *Cascader) CascadePath(ctx context.Context, src string, dests []string) []Result {
	srcPath := kvV2DataPath(src)
	secret, err := c.client.Logical().ReadWithContext(ctx, srcPath)
	if err != nil {
		results := make([]Result, len(dests))
		for i, d := range dests {
			results[i] = Result{Source: src, Dest: d, DryRun: c.dryRun, Err: fmt.Errorf("read source: %w", err)}
		}
		return results
	}
	if secret == nil || secret.Data == nil {
		results := make([]Result, len(dests))
		for i, d := range dests {
			results[i] = Result{Source: src, Dest: d, DryRun: c.dryRun, Err: fmt.Errorf("source path not found: %s", src)}
		}
		return results
	}

	data, _ := secret.Data["data"].(map[string]interface{})

	results := make([]Result, 0, len(dests))
	for _, dest := range dests {
		r := Result{Source: src, Dest: dest, DryRun: c.dryRun}
		if c.dryRun {
			results = append(results, r)
			continue
		}
		_, err := c.client.Logical().WriteWithContext(ctx, kvV2DataPath(dest), map[string]interface{}{"data": data})
		if err != nil {
			r.Err = fmt.Errorf("write dest %s: %w", dest, err)
		}
		results = append(results, r)
	}
	return results
}

func kvV2DataPath(path string) string {
	return "secret/data/" + path
}
