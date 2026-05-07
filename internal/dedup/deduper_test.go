package dedup_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/yourusername/vaultswap/internal/dedup"
)

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
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				store[path] = d
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
}

func newClient(t *testing.T, srv *httptest.Server) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = srv.URL
	c, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestDeduplicatePath_NoDuplicates_Skipped(t *testing.T) {
	store := map[string]map[string]interface{}{
		"myapp/config": {"alpha": "one", "beta": "two"},
	}
	srv := newVaultServer(t, store)
	defer srv.Close()

	d := dedup.New(newClient(t, srv), false)
	res := d.DeduplicatePath(context.Background(), "myapp/config")

	if !res.Skipped {
		t.Errorf("expected Skipped=true, got false")
	}
	if res.Removed != 0 {
		t.Errorf("expected Removed=0, got %d", res.Removed)
	}
}

func TestDeduplicatePath_RemovesDuplicates(t *testing.T) {
	store := map[string]map[string]interface{}{
		"myapp/config": {"alpha": "same", "beta": "same", "gamma": "unique"},
	}
	srv := newVaultServer(t, store)
	defer srv.Close()

	d := dedup.New(newClient(t, srv), false)
	res := d.DeduplicatePath(context.Background(), "myapp/config")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Removed != 1 {
		t.Errorf("expected Removed=1, got %d", res.Removed)
	}
	if len(res.Duplicates) != 1 {
		t.Errorf("expected 1 duplicate key, got %d", len(res.Duplicates))
	}
}

func TestDeduplicatePath_DryRun_DoesNotWrite(t *testing.T) {
	original := map[string]interface{}{"x": "dup", "y": "dup"}
	store := map[string]map[string]interface{}{
		"myapp/secrets": original,
	}
	srv := newVaultServer(t, store)
	defer srv.Close()

	d := dedup.New(newClient(t, srv), true)
	res := d.DeduplicatePath(context.Background(), "myapp/secrets")

	if !res.DryRun {
		t.Errorf("expected DryRun=true")
	}
	if res.Removed != 1 {
		t.Errorf("expected Removed=1, got %d", res.Removed)
	}
	// store should be unchanged
	if len(store["myapp/secrets"]) != 2 {
		t.Errorf("store was mutated during dry-run")
	}
}

func TestDeduplicatePaths_ReturnsAllResults(t *testing.T) {
	store := map[string]map[string]interface{}{
		"a": {"k1": "v", "k2": "v"},
		"b": {"k1": "unique"},
	}
	srv := newVaultServer(t, store)
	defer srv.Close()

	d := dedup.New(newClient(t, srv), false)
	results := d.DeduplicatePaths(context.Background(), []string{"a", "b"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
