package touch_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/nicholasgasior/vaultswap/internal/touch"
)

func newTouchServer(t *testing.T, data map[string]interface{}) *httptest.Server {
	t.Helper()
	written := false
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/data/") {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": data},
			})
			return
		}
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/data/") {
			written = true
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"version": 2}})
			return
		}
		_ = written // suppress unused warning
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newTouchClient(t *testing.T, srv *httptest.Server) *api.Client {
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

func TestTouchPath_DryRun_DoesNotWrite(t *testing.T) {
	var writeHit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"key": "val"}},
			})
			return
		}
		if r.Method == http.MethodPut {
			writeHit = true
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	client := newTouchClient(t, srv)
	tr := touch.New(client, true)
	res := tr.TouchPath(context.Background(), "secret/myapp/config")

	if !res.Touched {
		t.Error("expected Touched=true in dry-run")
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
	if writeHit {
		t.Error("write should not occur in dry-run mode")
	}
}

func TestTouchPath_Success(t *testing.T) {
	data := map[string]interface{}{"api_key": "abc123"}
	srv := newTouchServer(t, data)
	t.Cleanup(srv.Close)

	client := newTouchClient(t, srv)
	tr := touch.New(client, false)
	res := tr.TouchPath(context.Background(), "secret/myapp/config")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.Touched {
		t.Error("expected Touched=true")
	}
	if res.DryRun {
		t.Error("expected DryRun=false")
	}
}

func TestTouchPaths_ReturnsAllResults(t *testing.T) {
	data := map[string]interface{}{"x": "y"}
	srv := newTouchServer(t, data)
	t.Cleanup(srv.Close)

	client := newTouchClient(t, srv)
	tr := touch.New(client, false)
	paths := []string{"secret/a", "secret/b", "secret/c"}
	results := tr.TouchPaths(context.Background(), paths)

	if len(results) != len(paths) {
		t.Fatalf("expected %d results, got %d", len(paths), len(results))
	}
}
