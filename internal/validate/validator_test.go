package validate_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/seatgeek/vaultswap/internal/validate"
	"github.com/seatgeek/vaultswap/internal/vault"
)

func newTestServer(t *testing.T, secrets map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/secret/data/")
		data, ok := secrets[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": data},
		})
	}))
}

func newClient(t *testing.T, addr string) *vault.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	c, err := vault.NewClient(cfg.Address, "test-token", "")
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	return c
}

func TestValidatePath_AllKeysPresent(t *testing.T) {
	srv := newTestServer(t, map[string]map[string]interface{}{
		"myapp/config": {"DB_HOST": "localhost", "DB_PASS": "secret"},
	})
	defer srv.Close()

	client := newClient(t, srv.URL)
	v := validate.New(client, []string{"DB_HOST", "DB_PASS"})
	res := v.ValidatePath(context.Background(), "myapp/config")

	if !res.OK {
		t.Fatalf("expected OK, got missing: %v", res.Missing)
	}
	if len(res.Missing) != 0 {
		t.Errorf("expected no missing keys, got %v", res.Missing)
	}
}

func TestValidatePath_MissingKeys(t *testing.T) {
	srv := newTestServer(t, map[string]map[string]interface{}{
		"myapp/config": {"DB_HOST": "localhost"},
	})
	defer srv.Close()

	client := newClient(t, srv.URL)
	v := validate.New(client, []string{"DB_HOST", "DB_PASS", "API_KEY"})
	res := v.ValidatePath(context.Background(), "myapp/config")

	if res.OK {
		t.Fatal("expected not OK due to missing keys")
	}
	if len(res.Missing) != 2 {
		t.Errorf("expected 2 missing keys, got %d: %v", len(res.Missing), res.Missing)
	}
}

func TestValidatePath_ReadError(t *testing.T) {
	srv := newTestServer(t, map[string]map[string]interface{}{})
	defer srv.Close()

	client := newClient(t, srv.URL)
	v := validate.New(client, []string{"KEY"})
	res := v.ValidatePath(context.Background(), "nonexistent/path")

	if res.Err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestValidatePaths_ReturnsAllResults(t *testing.T) {
	srv := newTestServer(t, map[string]map[string]interface{}{
		"app/a": {"X": "1"},
		"app/b": {"X": "2"},
	})
	defer srv.Close()

	client := newClient(t, srv.URL)
	v := validate.New(client, []string{"X"})
	results := v.ValidatePaths(context.Background(), []string{"app/a", "app/b"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.OK {
			t.Errorf("expected OK for path %s", r.Path)
		}
	}
}
