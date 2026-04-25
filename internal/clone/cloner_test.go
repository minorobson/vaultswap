package clone_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaultswap/internal/clone"
	"github.com/vaultswap/internal/vault"
)

func newVaultServer(t *testing.T, data map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": data},
			})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}

func newClient(t *testing.T, srv *httptest.Server) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(srv.URL, "test-token", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestClonePath_Success(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	srcSrv := newVaultServer(t, data)
	dstSrv := newVaultServer(t, nil)
	defer srcSrv.Close()
	defer dstSrv.Close()

	c := clone.New(newClient(t, srcSrv), newClient(t, dstSrv), false)
	res := c.ClonePath(context.Background(), "secret/myapp")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Skipped {
		t.Error("expected not skipped")
	}
}

func TestClonePath_DryRun_DoesNotWrite(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	srcSrv := newVaultServer(t, data)
	dstSrv := newVaultServer(t, nil)
	defer srcSrv.Close()
	defer dstSrv.Close()

	c := clone.New(newClient(t, srcSrv), newClient(t, dstSrv), true)
	res := c.ClonePath(context.Background(), "secret/myapp")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestClonePaths_ReturnsAllResults(t *testing.T) {
	data := map[string]interface{}{"x": "1"}
	srcSrv := newVaultServer(t, data)
	dstSrv := newVaultServer(t, nil)
	defer srcSrv.Close()
	defer dstSrv.Close()

	c := clone.New(newClient(t, srcSrv), newClient(t, dstSrv), false)
	paths := []string{"secret/a", "secret/b", "secret/c"}
	results := c.ClonePaths(context.Background(), paths)

	if len(results) != len(paths) {
		t.Fatalf("expected %d results, got %d", len(paths), len(results))
	}
}
