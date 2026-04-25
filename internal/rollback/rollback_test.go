package rollback_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/vaultswap/internal/audit"
	"github.com/vaultswap/internal/rollback"
	"github.com/vaultswap/internal/snapshot"
)

type fakeClient struct {
	written map[string]map[string]interface{}
}

func (f *fakeClient) WriteSecret(_ context.Context, path string, data map[string]interface{}) error {
	if f.written == nil {
		f.written = make(map[string]map[string]interface{})
	}
	f.written[path] = data
	return nil
}

func writeSnapshot(t *testing.T, secrets map[string]map[string]interface{}) string {
	t.Helper()
	snap := &snapshot.Snapshot{Secrets: secrets}
	dir := t.TempDir()
	path := filepath.Join(dir, "snap.json")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(snap); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRollback_RestoresSecrets(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"secret/app": {"key": "value"},
	}
	snapshotPath := writeSnapshot(t, secrets)

	client := &fakeClient{}
	logger, _ := audit.New("")
	rb := rollback.New(client, logger, false)

	results, err := rb.Rollback(context.Background(), snapshotPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Restored {
		t.Error("expected Restored to be true")
	}
	if _, ok := client.written["secret/app"]; !ok {
		t.Error("expected secret/app to be written")
	}
}

func TestRollback_DryRun_DoesNotWrite(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"secret/app": {"key": "value"},
	}
	snapshotPath := writeSnapshot(t, secrets)

	client := &fakeClient{}
	logger, _ := audit.New("")
	rb := rollback.New(client, logger, true)

	results, err := rb.Rollback(context.Background(), snapshotPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Restored {
		t.Error("expected Restored to be false in dry-run")
	}
	if len(client.written) != 0 {
		t.Error("expected no writes in dry-run mode")
	}
}

func TestRollback_MissingSnapshot(t *testing.T) {
	client := &fakeClient{}
	logger, _ := audit.New("")
	rb := rollback.New(client, logger, false)

	_, err := rb.Rollback(context.Background(), "/nonexistent/snap.json")
	if err == nil {
		t.Fatal("expected error for missing snapshot file")
	}
}
