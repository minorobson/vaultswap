package copy_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vaultclient "vaultswap/internal/vault"

	"vaultswap/internal/copy"
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
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				store[path] = d
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
}

func newClient(t *testing.T, addr string) *vaultclient.Client {
	t.Helper()
	c, err := vaultclient.NewClient(vaultclient.Config{Address: addr, Token: "test"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestCopyPath_Success(t *testing.T) {
	store := map[string]map[string]interface{}{
		"src/key": {"password": "s3cr3t"},
	}
	srv := newVaultServer(t, store)
	defer srv.Close()

	c := newClient(t, srv.URL)
	cp := copy.New(c, false)
	res := cp.CopyPath(context.Background(), "src/key", "dst/key")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if _, ok := store["dst/key"]; !ok {
		t.Fatal("expected dst/key to be written")
	}
}

func TestCopyPath_DryRun_DoesNotWrite(t *testing.T) {
	store := map[string]map[string]interface{}{
		"src/key": {"token": "abc"},
	}
	srv := newVaultServer(t, store)
	defer srv.Close()

	c := newClient(t, srv.URL)
	cp := copy.New(c, true)
	res := cp.CopyPath(context.Background(), "src/key", "dst/key")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.DryRun {
		t.Fatal("expected DryRun=true")
	}
	if _, ok := store["dst/key"]; ok {
		t.Fatal("expected dst/key NOT to be written in dry-run mode")
	}
}

func TestCopyPaths_ReturnsAllResults(t *testing.T) {
	store := map[string]map[string]interface{}{
		"a": {"x": "1"},
		"b": {"y": "2"},
	}
	srv := newVaultServer(t, store)
	defer srv.Close()

	c := newClient(t, srv.URL)
	cp := copy.New(c, false)
	results := cp.CopyPaths(context.Background(), [][2]string{{"a", "a2"}, {"b", "b2"}})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.Src, r.Err)
		}
	}
}
