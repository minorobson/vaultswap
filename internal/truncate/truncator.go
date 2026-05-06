package truncate

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Result holds the outcome of truncating versions at a single path.
type Result struct {
	Path    string
	Kept    int
	Dropped int
	DryRun  bool
	Err     error
}

// Truncator deletes old secret versions beyond a retention limit.
type Truncator struct {
	client *api.Client
	mount  string
	keep   int
	dryRun bool
}

// New returns a Truncator configured with the given Vault client.
func New(client *api.Client, mount string, keep int, dryRun bool) *Truncator {
	return &Truncator{client: client, mount: mount, keep: keep, dryRun: dryRun}
}

// TruncatePath removes all versions beyond the keep limit for the given path.
func (t *Truncator) TruncatePath(ctx context.Context, path string) Result {
	metaPath := fmt.Sprintf("%s/metadata/%s", t.mount, path)

	secret, err := t.client.Logical().ReadWithContext(ctx, metaPath)
	if err != nil {
		return Result{Path: path, Err: fmt.Errorf("read metadata: %w", err)}
	}
	if secret == nil || secret.Data == nil {
		return Result{Path: path, Err: fmt.Errorf("no metadata found")}
	}

	versions, ok := secret.Data["versions"].(map[string]interface{})
	if !ok {
		return Result{Path: path, Err: fmt.Errorf("unexpected versions format")}
	}

	var versionNums []int
	for k := range versions {
		var n int
		if _, err := fmt.Sscanf(k, "%d", &n); err == nil {
			versionNums = append(versionNums, n)
		}
	}

	if len(versionNums) <= t.keep {
		return Result{Path: path, Kept: len(versionNums), Dropped: 0, DryRun: t.dryRun}
	}

	// Sort ascending to find oldest
	for i := 0; i < len(versionNums)-1; i++ {
		for j := i + 1; j < len(versionNums); j++ {
			if versionNums[j] < versionNums[i] {
				versionNums[i], versionNums[j] = versionNums[j], versionNums[i]
			}
		}
	}

	dropCount := len(versionNums) - t.keep
	toDrop := versionNums[:dropCount]

	if !t.dryRun {
		versionsPayload := make([]interface{}, len(toDrop))
		for i, v := range toDrop {
			versionsPayload[i] = v
		}
		destroyPath := fmt.Sprintf("%s/destroy/%s", t.mount, path)
		_, err = t.client.Logical().WriteWithContext(ctx, destroyPath, map[string]interface{}{
			"versions": versionsPayload,
		})
		if err != nil {
			return Result{Path: path, Err: fmt.Errorf("destroy versions: %w", err)}
		}
	}

	return Result{Path: path, Kept: t.keep, Dropped: dropCount, DryRun: t.dryRun}
}

// TruncatePaths runs TruncatePath for each entry in paths.
func (t *Truncator) TruncatePaths(ctx context.Context, paths []string) []Result {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		results = append(results, t.TruncatePath(ctx, p))
	}
	return results
}
