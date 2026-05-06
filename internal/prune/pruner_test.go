package prune

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newPruneServer(t *testing.T, mount, path string, data map[string]interface{}, allowDelete bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dataPath := "/v1/" + mount + "/data/" + path
		metaPath := "/v1/" + mount + "/metadata/" + path
		switch {
		case r.Method == http.MethodGet && r.URL.Path == dataPath:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
		case r.Method == http.MethodDelete && r.URL.Path == metaPath:
			if allowDelete {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newPruneClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	cfg := api.DefaultConfig()
	cfg.Address = addr
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.SetToken("test-token")
	return client
}

func TestPrunePath_DryRun_DoesNotWrite(t *testing.T) {
	srv := newPruneServer(t, "secret", "empty/key", map[string]interface{}{}, true)
	defer srv.Close()

	p := New(newPruneClient(t, srv.URL), true)
	res := p.PrunePath(context.Background(), "secret", "empty/key")

	if !res.Pruned || !res.DryRun {
		t.Errorf("expected dry-run prune, got pruned=%v dryRun=%v", res.Pruned, res.DryRun)
	}
}

func TestPrunePath_NonEmpty_Skipped(t *testing.T) {
	srv := newPruneServer(t, "secret", "full/key", map[string]interface{}{"foo": "bar"}, false)
	defer srv.Close()

	p := New(newPruneClient(t, srv.URL), false)
	res := p.PrunePath(context.Background(), "secret", "full/key")

	if !res.Skipped {
		t.Errorf("expected skipped result for non-empty secret, got %+v", res)
	}
}

func TestPrunePath_DeleteForbidden_ReturnsError(t *testing.T) {
	srv := newPruneServer(t, "secret", "empty/key", map[string]interface{}{}, false)
	defer srv.Close()

	p := New(newPruneClient(t, srv.URL), false)
	res := p.PrunePath(context.Background(), "secret", "empty/key")

	if res.Err == nil {
		t.Errorf("expected an error when delete is forbidden, got nil")
	}
}

func TestPrunePaths_ReturnsAllResults(t *testing.T) {
	srv := newPruneServer(t, "secret", "empty/key", map[string]interface{}{}, true)
	defer srv.Close()

	p := New(newPruneClient(t, srv.URL), false)
	results := p.PrunePaths(context.Background(), "secret", []string{"empty/key", "missing/key"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestFprintResults_Labels(t *testing.T) {
	results := []Result{
		{Path: "a", Pruned: true},
		{Path: "b", Pruned: true, DryRun: true},
		{Path: "c", Skipped: true},
	}
	var sb strings.Builder
	FprintResults(&sb, results)
	out := sb.String()

	for _, want := range []string{"[pruned]", "[dry-run]", "[skipped]"}  {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}
