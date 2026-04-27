package tags_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/seatgeek/vaultswap/internal/tags"
	"github.com/seatgeek/vaultswap/internal/vault"
)

func newTestVaultServer(t *testing.T, secrets map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/secret/data/")
		if r.Method == http.MethodGet {
			data, ok := secrets[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				secrets[path] = d
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	}))
}

func newClient(t *testing.T, addr string) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("vault client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestTagPath_DryRun_DoesNotWrite(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"myapp/config": {"API_KEY": "abc123"},
	}
	srv := newTestVaultServer(t, secrets)
	defer srv.Close()

	c := vault.NewClient(newClient(t, srv.URL), "")
	tagger := tags.New(c, true)

	res := tagger.TagPath(context.Background(), "myapp/config", map[string]string{"env": "prod"})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
	// Verify no write occurred.
	if _, ok := secrets["myapp/config"]["_tags.env"]; ok {
		t.Error("tag should not have been written in dry-run mode")
	}
}

func TestTagPath_WritesTags(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"myapp/config": {"API_KEY": "abc123"},
	}
	srv := newTestVaultServer(t, secrets)
	defer srv.Close()

	c := vault.NewClient(newClient(t, srv.URL), "")
	tagger := tags.New(c, false)

	res := tagger.TagPath(context.Background(), "myapp/config", map[string]string{"team": "platform"})
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.DryRun {
		t.Error("expected DryRun=false")
	}
	if v, ok := secrets["myapp/config"]["_tags.team"]; !ok || v != "platform" {
		t.Errorf("expected tag _tags.team=platform, got %v", secrets["myapp/config"])
	}
}

func TestTagPaths_ReturnsAllResults(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"svc/a": {"X": "1"},
		"svc/b": {"Y": "2"},
	}
	srv := newTestVaultServer(t, secrets)
	defer srv.Close()

	c := vault.NewClient(newClient(t, srv.URL), "")
	tagger := tags.New(c, false)

	results := tagger.TagPaths(context.Background(), []string{"svc/a", "svc/b"}, map[string]string{"owner": "infra"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("path %s: unexpected error: %v", r.Path, r.Err)
		}
	}
}
