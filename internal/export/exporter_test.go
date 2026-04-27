package export_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/user/vaultswap/internal/export"
	"github.com/user/vaultswap/internal/vault"
)

func newTestVaultServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/data/myapp/config":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"API_KEY": "abc123"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("api.NewClient: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestExportPaths_Success(t *testing.T) {
	srv := newTestVaultServer(t)
	defer srv.Close()

	c := vault.NewClient(newClient(t, srv.URL))
	ex := export.New(c)
	results := ex.ExportPaths([]string{"myapp/config"})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error != nil {
		t.Fatalf("unexpected error: %v", results[0].Error)
	}
	if results[0].Data["API_KEY"] != "abc123" {
		t.Errorf("unexpected data: %v", results[0].Data)
	}
}

func TestExportPaths_ReadError(t *testing.T) {
	srv := newTestVaultServer(t)
	defer srv.Close()

	c := vault.NewClient(newClient(t, srv.URL))
	ex := export.New(c)
	results := ex.ExportPaths([]string{"myapp/missing"})

	if results[0].Error == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestWriteFile_JSON(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.json")

	results := []export.Result{
		{Path: "myapp/config", Data: map[string]interface{}{"KEY": "val"}, Error: nil},
		{Path: "myapp/bad", Data: nil, Error: errors.New("not found")},
	}
	if err := export.WriteFile(results, dest, "json"); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	raw, _ := os.ReadFile(dest)
	var parsed map[string]interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := parsed["myapp/config"]; !ok {
		t.Error("expected myapp/config in output")
	}
	if _, ok := parsed["myapp/bad"]; ok {
		t.Error("errored path should be excluded")
	}
}

func TestWriteFile_UnsupportedFormat(t *testing.T) {
	dir := t.TempDir()
	err := export.WriteFile(nil, filepath.Join(dir, "out.xml"), "xml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}
