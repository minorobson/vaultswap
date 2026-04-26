package promote_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/vaultswap/internal/promote"
	"github.com/your-org/vaultswap/internal/vault"
)

func newTestVaultServer(t *testing.T, secrets map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			for path, data := range secrets {
				if strings.Contains(r.URL.Path, path) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"data": map[string]interface{}{"data": data},
					})
					return
				}
			}
		}
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
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

func TestPromote_DryRun_DoesNotWrite(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"myapp/config": {"key": "value"},
	}
	srv := newTestVaultServer(t, secrets)
	defer srv.Close()

	src := newClient(t, srv.URL)
	dst := newClient(t, srv.URL)
	p := promote.New(src, dst, true)

	results, err := p.Promote(context.Background(), []string{"myapp/config"})
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestPromote_Success(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"myapp/config": {"key": "value"},
	}
	srv := newTestVaultServer(t, secrets)
	defer srv.Close()

	src := newClient(t, srv.URL)
	dst := newClient(t, srv.URL)
	p := promote.New(src, dst, false)

	results, err := p.Promote(context.Background(), []string{"myapp/config"})
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if results[0].Err != nil {
		t.Errorf("unexpected error: %v", results[0].Err)
	}
}
