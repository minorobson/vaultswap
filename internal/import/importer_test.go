package importpkg_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	importpkg "github.com/your-org/vaultswap/internal/import"
	"github.com/your-org/vaultswap/internal/vault"
)

func newTestVaultServer(t *testing.T, written *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			*written = append(*written, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
	}))
}

func newClient(t *testing.T, addr string) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(vault.Config{Address: addr, Token: "test-token"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func writeTempFile(t *testing.T, payload map[string]map[string]string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "import-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(f).Encode(payload); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return f.Name()
}

func TestImportFile_WritesSecrets(t *testing.T) {
	var written []string
	srv := newTestVaultServer(t, &written)
	defer srv.Close()

	payload := map[string]map[string]string{
		"secret/app/db": {"password": "s3cr3t"},
	}
	file := writeTempFile(t, payload)

	im := importpkg.New(newClient(t, srv.URL))
	results, err := im.ImportFile(context.Background(), file, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Written {
		t.Error("expected Written=true")
	}
}

func TestImportFile_DryRun_DoesNotWrite(t *testing.T) {
	var written []string
	srv := newTestVaultServer(t, &written)
	defer srv.Close()

	payload := map[string]map[string]string{
		"secret/app/api": {"key": "abc123"},
	}
	file := writeTempFile(t, payload)

	im := importpkg.New(newClient(t, srv.URL))
	results, err := im.ImportFile(context.Background(), file, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(written) != 0 {
		t.Errorf("expected no writes, got %d", len(written))
	}
	if results[0].Written {
		t.Error("expected Written=false in dry-run mode")
	}
}

func TestImportFile_MissingFile_ReturnsError(t *testing.T) {
	im := importpkg.New(newClient(t, "http://127.0.0.1:1"))
	_, err := im.ImportFile(context.Background(), "/nonexistent/file.json", false)
	if err == nil {
		t.Error("expected error for missing file")
	}
}
