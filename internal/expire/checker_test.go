package expire_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/yourusername/vaultswap/internal/expire"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func newClient(t *testing.T, addr string) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("new vault client: %v", err)
	}
	client.SetToken("test-token")
	return client
}

func TestCheckPath_NoMetadata(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	checker := expire.New(newClient(t, srv.URL), "ns1")
	res := checker.CheckPath(context.Background(), "app/db")
	if res.Error == nil {
		t.Fatal("expected error for missing metadata, got nil")
	}
}

func TestCheckPath_WithTTL(t *testing.T) {
	createdAt := time.Now().Add(-10 * time.Hour).UTC().Format(time.RFC3339Nano)
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"current_version":      1,
			"delete_version_after": "24h",
			"versions": map[string]interface{}{
				"1": map[string]interface{}{
					"created_time": createdAt,
				},
			},
		},
	}
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	})
	checker := expire.New(newClient(t, srv.URL), "ns1")
	res := checker.CheckPath(context.Background(), "app/db")
	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if res.TTL != 24*time.Hour {
		t.Errorf("expected TTL 24h, got %v", res.TTL)
	}
	if res.Expired {
		t.Error("expected secret to not be expired yet")
	}
}

func TestCheckPaths_ReturnsAllResults(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	checker := expire.New(newClient(t, srv.URL), "ns1")
	paths := []string{"a/b", "c/d", "e/f"}
	results := checker.CheckPaths(context.Background(), paths)
	if len(results) != len(paths) {
		t.Errorf("expected %d results, got %d", len(paths), len(results))
	}
	for i, res := range results {
		if res.Path != paths[i] {
			t.Errorf("result[%d]: expected path %s, got %s", i, paths[i], res.Path)
		}
	}
}
