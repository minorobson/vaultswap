package snapshot_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/vaultswap/internal/snapshot"
)

// fakeReader implements snapshot.SecretReader.
type fakeReader struct {
	data map[string]string
	err  error
}

func (f *fakeReader) ReadSecret(_, _ string) (map[string]string, error) {
	return f.data, f.err
}

func TestTake_Success(t *testing.T) {
	reader := &fakeReader{
		data: map[string]string{"api_key": "abc123", "db_pass": "secret"},
	}
	taker := snapshot.New(reader)

	snap, err := taker.Take("ns1", "kv/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Path != "kv/myapp" {
		t.Errorf("expected path kv/myapp, got %s", snap.Path)
	}
	if snap.Namespace != "ns1" {
		t.Errorf("expected namespace ns1, got %s", snap.Namespace)
	}
	if snap.Data["api_key"] != "abc123" {
		t.Errorf("expected api_key=abc123, got %s", snap.Data["api_key"])
	}
	if snap.CapturedAt.IsZero() {
		t.Error("expected CapturedAt to be set")
	}
}

func TestTake_ReadError(t *testing.T) {
	reader := &fakeReader{err: errors.New("vault unavailable")}
	taker := snapshot.New(reader)

	_, err := taker.Take("ns1", "kv/myapp")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	reader := &fakeReader{
		data: map[string]string{"token": "xyz"},
	}
	taker := snapshot.New(reader)
	snap, err := taker.Take("prod", "kv/service")
	if err != nil {
		t.Fatalf("take: %v", err)
	}

	tmpFile := filepath.Join(t.TempDir(), "snap.json")
	if err := snap.Save(tmpFile); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := snapshot.Load(tmpFile)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Path != snap.Path {
		t.Errorf("path mismatch: got %s", loaded.Path)
	}
	if loaded.Data["token"] != "xyz" {
		t.Errorf("data mismatch: got %s", loaded.Data["token"])
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := snapshot.Load("/nonexistent/path/snap.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSave_InvalidPath(t *testing.T) {
	snap := &snapshot.Snapshot{Path: "kv/test", Data: map[string]string{}}
	err := snap.Save("/nonexistent_dir/snap.json")
	if !errors.Is(err, os.ErrNotExist) && err == nil {
		t.Fatal("expected error for invalid save path")
	}
}
