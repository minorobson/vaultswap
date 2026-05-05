package merge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

// newVaultServer spins up a minimal fake Vault KV-v2 server.
func newVaultServer(t *testing.T, store map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/secret/data/")
		switch r.Method {
		case http.MethodGet:
			data, ok := store[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"]; ok {
				if dm, ok := d.(map[string]interface{}); ok {
					store[path] = dm
				}
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
}

func newClient(t *testing.T, addr string) VaultClient {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("vault client: %v", err)
	}
	c.SetToken("test-token")
	return &vaultClientAdapter{c}
}

// vaultClientAdapter wraps the real vault client to satisfy VaultClient.
type vaultClientAdapter struct{ c *vaultapi.Client }

func (a *vaultClientAdapter) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	s, err := a.c.KVv2("secret").Get(ctx, path)
	if err != nil {
		return nil, err
	}
	return s.Data, nil
}

func (a *vaultClientAdapter) WriteSecret(ctx context.Context, path string, data map[string]interface{}) error {
	_, err := a.c.KVv2("secret").Put(ctx, path, data)
	return err
}

func TestMergePath_NoOverwrite_SkipsExistingKeys(t *testing.T) {
	store := map[string]map[string]interface{}{
		"src":  {"a": "1", "b": "2"},
		"dest": {"a": "original"},
	}
	srv := newVaultServer(t, store)
	defer srv.Close()

	m := New(newClient(t, srv.URL), Options{Overwrite: false})
	r := m.MergePath(context.Background(), "dest", "src")

	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
	if r.Merged != 1 {
		t.Errorf("merged = %d, want 1", r.Merged)
	}
	if r.Skipped != 1 {
		t.Errorf("skipped = %d, want 1", r.Skipped)
	}
}

func TestMergePath_Overwrite_ReplacesKeys(t *testing.T) {
	store := map[string]map[string]interface{}{
		"src":  {"a": "new"},
		"dest": {"a": "old"},
	}
	srv := newVaultServer(t, store)
	defer srv.Close()

	m := New(newClient(t, srv.URL), Options{Overwrite: true})
	r := m.MergePath(context.Background(), "dest", "src")

	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
	if r.Merged != 1 {
		t.Errorf("merged = %d, want 1", r.Merged)
	}
	if r.Skipped != 0 {
		t.Errorf("skipped = %d, want 0", r.Skipped)
	}
}

func TestMergePath_DryRun_DoesNotWrite(t *testing.T) {
	store := map[string]map[string]interface{}{
		"src":  {"x": "val"},
		"dest": {},
	}
	srv := newVaultServer(t, store)
	defer srv.Close()

	m := New(newClient(t, srv.URL), Options{DryRun: true})
	r := m.MergePath(context.Background(), "dest", "src")

	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
	if !r.DryRun {
		t.Error("expected DryRun=true")
	}
	if _, written := store["dest"]["x"]; written {
		t.Error("dry run should not have written to dest")
	}
}
