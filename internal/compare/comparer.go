package compare

import (
	"fmt"

	"github.com/vaultswap/internal/diff"
	"github.com/vaultswap/internal/vault"
)

// Result holds the comparison outcome for a single secret path.
type Result struct {
	Path    string
	Diff    []diff.Change
	HasDiff bool
}

// Comparer reads secrets from two Vault clients and diffs them.
type Comparer struct {
	src  *vault.Client
	dst  *vault.Client
	mask bool
}

// New creates a Comparer between src and dst Vault clients.
func New(src, dst *vault.Client, maskValues bool) *Comparer {
	return &Comparer{src: src, dst: dst, mask: maskValues}
}

// ComparePath reads the secret at path from both clients and returns a Result.
func (c *Comparer) ComparePath(path string) (Result, error) {
	srcData, err := c.src.ReadSecret(path)
	if err != nil {
		return Result{}, fmt.Errorf("read source %q: %w", path, err)
	}

	dstData, err := c.dst.ReadSecret(path)
	if err != nil {
		return Result{}, fmt.Errorf("read destination %q: %w", path, err)
	}

	changes := diff.Compare(srcData, dstData)
	return Result{
		Path:    path,
		Diff:    changes,
		HasDiff: len(changes) > 0,
	}, nil
}

// ComparePaths runs ComparePath over multiple paths and collects results.
func (c *Comparer) ComparePaths(paths []string) ([]Result, error) {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		r, err := c.ComparePath(p)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}
