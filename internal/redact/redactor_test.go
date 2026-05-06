package redact_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/vaultswap/internal/redact"
	"github.com/your-org/vaultswap/internal/vault"
)

func newTestVaultServer(t *testing.T, data map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": data},
			})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}

func newClient(t *testing.T, addr string) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(vault.Config{Address: addr, Token: "test"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestRedactPath_DryRun_DoesNotWrite(t *testing.T) {
	written := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			written = true
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"password": "secret123"}},
		})
	}))
	defer srv.Close()

	r, err := redact.New(newClient(t, srv.URL), `secret`, "REDACTED", true)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := r.RedactPath("myapp/config")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if written {
		t.Fatal("expected no write in dry-run mode")
	}
	if len(res.Keys) == 0 {
		t.Fatal("expected matched keys")
	}
}

func TestRedactPath_NoMatch_Skipped(t *testing.T) {
	srv := newTestVaultServer(t, map[string]interface{}{"api_key": "abc123"})
	defer srv.Close()

	r, err := redact.New(newClient(t, srv.URL), `^Bearer `, "REDACTED", false)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := r.RedactPath("myapp/config")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if len(res.Keys) != 0 {
		t.Fatalf("expected no matched keys, got %v", res.Keys)
	}
}

func TestRedactPaths_ReturnsAllResults(t *testing.T) {
	srv := newTestVaultServer(t, map[string]interface{}{"token": "Bearer xyz"})
	defer srv.Close()

	r, err := redact.New(newClient(t, srv.URL), `Bearer`, "REDACTED", true)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	results := r.RedactPaths([]string{"a", "b", "c"})
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	srv := newTestVaultServer(t, nil)
	defer srv.Close()

	_, err := redact.New(newClient(t, srv.URL), `[invalid`, "X", false)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
	if !strings.Contains(err.Error(), "invalid pattern") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
