package revert_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"

	"github.com/yourusername/vaultswap/internal/revert"
)

func newRevertServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "?version="):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"data": map[string]any{"key": "oldvalue"}},
			})
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"data": map[string]any{"key": "oldvalue"}},
			})
		case r.Method == http.MethodPost || r.Method == http.MethodPut:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"version": 3},
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func newRevertClient(t *testing.T, addr string) *api.Client {
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

func TestRevertPath_DryRun_DoesNotWrite(t *testing.T) {
	srv := newRevertServer(t)
	defer srv.Close()

	client := newRevertClient(t, srv.URL)
	r := revert.New(client, true)

	res := r.RevertPath(context.Background(), "secret", "myapp/config", 2)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.DryRun {
		t.Error("expected DryRun to be true")
	}
}

func TestRevertPaths_ReturnsAllResults(t *testing.T) {
	srv := newRevertServer(t)
	defer srv.Close()

	client := newRevertClient(t, srv.URL)
	r := revert.New(client, true)

	targets := map[string]int{
		"app/db": 1,
		"app/api": 2,
	}
	results := r.RevertPaths(context.Background(), "secret", targets)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestFprintResults_Labels(t *testing.T) {
	results := []revert.Result{
		{Path: "app/db", Version: 1, DryRun: true},
		{Path: "app/api", Version: 2},
		{Path: "app/cache", Version: 3, Err: context.DeadlineExceeded},
	}
	var sb strings.Builder
	revert.FprintResults(&sb, results)
	out := sb.String()
	if !strings.Contains(out, "[dry-run]") {
		t.Error("expected dry-run label")
	}
	if !strings.Contains(out, "[reverted]") {
		t.Error("expected reverted label")
	}
	if !strings.Contains(out, "[error]") {
		t.Error("expected error label")
	}
}
