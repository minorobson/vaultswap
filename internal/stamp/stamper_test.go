package stamp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/wndhydrnt/vaultswap/internal/stamp"
	"github.com/wndhydrnt/vaultswap/internal/vault"
)

func newStampServer(t *testing.T, secret map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": secret},
			})
			return
		}
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
}

func newStampClient(t *testing.T, srv *httptest.Server) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("vault client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestStampPath_DryRun_DoesNotWrite(t *testing.T) {
	srv := newStampServer(t, map[string]interface{}{"username": "alice"})
	t.Cleanup(srv.Close)

	vc := newStampClient(t, srv)
	client := vault.NewClientFromAPI(vc)
	s := stamp.New(client, "_stamped_at", true)

	res := s.StampPath(context.Background(), "secret/data/app")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
	if !res.Stamped {
		t.Error("expected Stamped=true")
	}
}

func TestStampPath_WritesTimestamp(t *testing.T) {
	written := map[string]interface{}{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": map[string]interface{}{"key": "val"}},
			})
			return
		}
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&written)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}))
	t.Cleanup(srv.Close)

	vc := newStampClient(t, srv)
	client := vault.NewClientFromAPI(vc)
	s := stamp.New(client, "_stamped_at", false)

	res := s.StampPath(context.Background(), "secret/data/app")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.Stamped {
		t.Error("expected Stamped=true")
	}

	data, ok := written["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("written payload missing data field: %v", written)
	}
	ts, ok := data["_stamped_at"].(string)
	if !ok || !strings.Contains(ts, "T") {
		t.Errorf("expected RFC3339 timestamp, got %q", ts)
	}
}

func TestStampPaths_ReturnsAllResults(t *testing.T) {
	srv := newStampServer(t, map[string]interface{}{"x": "1"})
	t.Cleanup(srv.Close)

	vc := newStampClient(t, srv)
	client := vault.NewClientFromAPI(vc)
	s := stamp.New(client, "", true)

	paths := []string{"secret/data/a", "secret/data/b", "secret/data/c"}
	results := s.StampPaths(context.Background(), paths)
	if len(results) != len(paths) {
		t.Fatalf("expected %d results, got %d", len(paths), len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("path %s: unexpected error: %v", r.Path, r.Err)
		}
	}
}
