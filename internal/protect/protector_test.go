package protect_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/yourusername/vaultswap/internal/protect"
)

func newTestVaultServer(t *testing.T, customMeta map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/metadata/") {
			if customMeta != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{"custom_metadata": customMeta},
				})
				return
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/metadata/") {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
}

func newClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	client.SetToken("test-token")
	return client
}

func TestProtectPath_DryRun_DoesNotWrite(t *testing.T) {
	srv := newTestVaultServer(t, nil)
	defer srv.Close()

	p := protect.New(newClient(t, srv.URL), true)
	res := p.ProtectPath(context.Background(), "secret", "myapp/config")

	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if res.Skipped {
		t.Fatal("expected not skipped")
	}
	if !res.DryRun {
		t.Fatal("expected dry-run flag")
	}
}

func TestProtectPath_AlreadyProtected_Skipped(t *testing.T) {
	srv := newTestVaultServer(t, map[string]interface{}{"protected": "true"})
	defer srv.Close()

	p := protect.New(newClient(t, srv.URL), false)
	res := p.ProtectPath(context.Background(), "secret", "myapp/config")

	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if !res.Skipped {
		t.Fatal("expected skipped result")
	}
}

func TestProtectPath_WritesMetadata(t *testing.T) {
	srv := newTestVaultServer(t, nil)
	defer srv.Close()

	p := protect.New(newClient(t, srv.URL), false)
	res := p.ProtectPath(context.Background(), "secret", "myapp/config")

	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if res.Skipped || res.DryRun {
		t.Fatalf("expected live protect, got skipped=%v dryRun=%v", res.Skipped, res.DryRun)
	}
}

func TestProtectPaths_ReturnsAllResults(t *testing.T) {
	srv := newTestVaultServer(t, nil)
	defer srv.Close()

	p := protect.New(newClient(t, srv.URL), true)
	results := p.ProtectPaths(context.Background(), "secret", []string{"a", "b", "c"})

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}
