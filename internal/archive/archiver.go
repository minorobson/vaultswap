package archive

import (
	"context"
	"fmt"
	"time"
)

// SecretVersion holds a single versioned copy of a secret's data.
type SecretVersion struct {
	Path      string
	Version   int
	Data      map[string]any
	ArchivedAt time.Time
}

// Result is the outcome of archiving a single path.
type Result struct {
	Path    string
	Version int
	Skipped bool
	DryRun  bool
	Err     error
}

// VaultClient is the subset of vault operations required by the archiver.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (map[string]any, error)
	ReadSecretVersion(ctx context.Context, path string) (int, error)
	WriteSecret(ctx context.Context, path string, data map[string]any) error
}

// Archiver copies secrets to a timestamped archive path inside Vault.
type Archiver struct {
	client VaultClient
	dryRun bool
}

// New returns a new Archiver.
func New(client VaultClient, dryRun bool) *Archiver {
	return &Archiver{client: client, dryRun: dryRun}
}

// ArchivePath reads the current secret at path and writes it to
// archiveBase/<timestamp>/<path>, returning the result.
func (a *Archiver) ArchivePath(ctx context.Context, path, archiveBase string) Result {
	data, err := a.client.ReadSecret(ctx, path)
	if err != nil {
		return Result{Path: path, Err: fmt.Errorf("read: %w", err)}
	}

	version, err := a.client.ReadSecretVersion(ctx, path)
	if err != nil {
		return Result{Path: path, Err: fmt.Errorf("read version: %w", err)}
	}

	if a.dryRun {
		return Result{Path: path, Version: version, DryRun: true}
	}

	stamp := time.Now().UTC().Format("20060102T150405Z")
	dest := fmt.Sprintf("%s/%s/%s", archiveBase, stamp, path)

	if err := a.client.WriteSecret(ctx, dest, data); err != nil {
		return Result{Path: path, Err: fmt.Errorf("write archive: %w", err)}
	}

	return Result{Path: path, Version: version}
}

// ArchivePaths archives each path and returns all results.
func (a *Archiver) ArchivePaths(ctx context.Context, paths []string, archiveBase string) []Result {
	results := make([]Result, 0, len(paths))
	for _, p := range paths {
		results = append(results, a.ArchivePath(ctx, p, archiveBase))
	}
	return results
}
