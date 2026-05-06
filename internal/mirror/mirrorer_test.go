package mirror_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/fhemberger/vaultswap/internal/mirror"
)

func newVaultServer(t *testing.T, data map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/data/") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": data},
			})
			return
		}
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/data/") {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func newClient(t *testing.T, addr string) *api.Client {
	t.Helper()
	c, err := api.NewClient(&api.Config{Address: addr})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

func TestMirrorPath_DryRun_DoesNotWrite(t *testing.T) {
	written := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			written = true
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"key": "val"}},
		})
	}))
	defer srv.Close()

	cl := newClient(t, srv.URL)
	m := mirror.New(cl, cl, true)
	res := m.MirrorPath(context.Background(), "secret/src", "secret/dst")

	if written {
		t.Fatal("expected no write in dry-run mode")
	}
	if res.DryRun != true {
		t.Errorf("expected DryRun=true")
	}
}

func TestMirrorPaths_ReturnsAllResults(t *testing.T) {
	srv := newVaultServer(t, map[string]interface{}{"k": "v"})
	defer srv.Close()

	cl := newClient(t, srv.URL)
	m := mirror.New(cl, cl, false)
	pairs := [][2]string{{"a/b", "c/d"}, {"e/f", "g/h"}}
	results := m.MirrorPaths(context.Background(), pairs)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
