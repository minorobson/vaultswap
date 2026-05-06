package truncate_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/user/vaultswap/internal/truncate"
)

func newTruncateServer(t *testing.T, versions map[string]interface{}, expectDestroy bool) *httptest.Server {
	t.Helper()
	destroyCalled := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			payload := map[string]interface{}{
				"data": map[string]interface{}{"versions": versions},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(payload)
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			destroyCalled = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	t.Cleanup(func() {
		if expectDestroy && !destroyCalled {
			t.Error("expected destroy to be called but it was not")
		}
		if !expectDestroy && destroyCalled {
			t.Error("destroy was called but should not have been")
		}
		srv.Close()
	})
	return srv
}

func newTruncateClient(t *testing.T, addr string) *api.Client {
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

func TestTruncatePath_DryRun_DoesNotWrite(t *testing.T) {
	versions := map[string]interface{}{
		"1": map[string]interface{}{}, "2": map[string]interface{}{},
		"3": map[string]interface{}{}, "4": map[string]interface{}{},
	}
	srv := newTruncateServer(t, versions, false)
	client := newTruncateClient(t, srv.URL)
	tr := truncate.New(client, "secret", 2, true)

	res := tr.TruncatePath(context.Background(), "myapp/config")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
	if res.Dropped != 2 {
		t.Errorf("expected Dropped=2, got %d", res.Dropped)
	}
}

func TestTruncatePath_WithinLimit_NoDrop(t *testing.T) {
	versions := map[string]interface{}{
		"1": map[string]interface{}{}, "2": map[string]interface{}{},
	}
	srv := newTruncateServer(t, versions, false)
	client := newTruncateClient(t, srv.URL)
	tr := truncate.New(client, "secret", 5, false)

	res := tr.TruncatePath(context.Background(), "myapp/config")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Dropped != 0 {
		t.Errorf("expected Dropped=0, got %d", res.Dropped)
	}
	if res.Kept != 2 {
		t.Errorf("expected Kept=2, got %d", res.Kept)
	}
}

func TestTruncatePaths_ReturnsAllResults(t *testing.T) {
	versions := map[string]interface{}{
		"1": map[string]interface{}{}, "2": map[string]interface{}{},
	}
	srv := newTruncateServer(t, versions, false)
	client := newTruncateClient(t, srv.URL)
	tr := truncate.New(client, "secret", 5, false)

	paths := []string{"a/b", "c/d"}
	results := tr.TruncatePaths(context.Background(), paths)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
