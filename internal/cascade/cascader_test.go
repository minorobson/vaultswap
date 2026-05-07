package cascade_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/fvbock/vaultswap/internal/cascade"
)

func newVaultServer(t *testing.T, data map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/data/") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newClient(t *testing.T, srv *httptest.Server) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	client.SetToken("test-token")
	return client
}

func TestCascadePath_DryRun_DoesNotWrite(t *testing.T) {
	srv := newVaultServer(t, map[string]interface{}{"key": "value"})
	defer srv.Close()
	client := newClient(t, srv)

	c := cascade.New(client, true)
	results := c.CascadePath(context.Background(), "source/secret", []string{"dest/a", "dest/b"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error: %v", r.Err)
		}
		if !r.DryRun {
			t.Errorf("expected DryRun=true for %s", r.Dest)
		}
	}
}

func TestCascadePath_Success(t *testing.T) {
	srv := newVaultServer(t, map[string]interface{}{"api_key": "abc123"})
	defer srv.Close()
	client := newClient(t, srv)

	c := cascade.New(client, false)
	results := c.CascadePath(context.Background(), "source/secret", []string{"dest/a", "dest/b", "dest/c"})

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.Dest, r.Err)
		}
		if r.Source != "source/secret" {
			t.Errorf("unexpected source: %s", r.Source)
		}
	}
}

func TestCascadePath_ReturnsAllResults(t *testing.T) {
	srv := newVaultServer(t, map[string]interface{}{"token": "secret"})
	defer srv.Close()
	client := newClient(t, srv)

	c := cascade.New(client, false)
	dests := []string{"env/prod", "env/staging", "env/dev"}
	results := c.CascadePath(context.Background(), "shared/config", dests)

	if len(results) != len(dests) {
		t.Fatalf("expected %d results, got %d", len(dests), len(results))
	}
}
