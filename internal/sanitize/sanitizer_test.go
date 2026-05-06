package sanitize

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/vaultswap/internal/vault"
)

func newTestVaultServer(t *testing.T, secrets map[string]interface{}, written *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": secrets},
			})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if written != nil {
				if d, ok := body["data"].(map[string]interface{}); ok {
					*written = d
				}
			}
			w.WriteHeader(http.StatusOK)
		}
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

func TestSanitizePath_RemovesMatchingKeys(t *testing.T) {
	secrets := map[string]interface{}{"api_key": "changeme", "password": "s3cr3t", "host": "localhost"}
	var written map[string]interface{}
	srv := newTestVaultServer(t, secrets, &written)
	defer srv.Close()

	s := New(newClient(t, srv.URL), []string{"changeme", "s3cr3t"}, false)
	results, err := s.SanitizePath("secret/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("result error: %v", r.Err)
		}
		if !r.Removed {
			t.Errorf("expected key %q to be marked removed", r.Key)
		}
	}
	if _, ok := written["host"]; !ok {
		t.Errorf("expected 'host' to be retained in written payload")
	}
}

func TestSanitizePath_DryRun_DoesNotWrite(t *testing.T) {
	secrets := map[string]interface{}{"token": "changeme"}
	var written map[string]interface{}
	srv := newTestVaultServer(t, secrets, &written)
	defer srv.Close()

	s := New(newClient(t, srv.URL), []string{"changeme"}, true)
	results, err := s.SanitizePath("secret/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Removed {
		t.Error("dry-run should not mark key as removed")
	}
	if written != nil {
		t.Error("dry-run should not write to vault")
	}
}

func TestSanitizePaths_ReturnsAllResults(t *testing.T) {
	secrets := map[string]interface{}{"pw": "todo", "keep": "real-value"}
	srv := newTestVaultServer(t, secrets, nil)
	defer srv.Close()

	s := New(newClient(t, srv.URL), []string{"todo"}, false)
	results := s.SanitizePaths([]string{"secret/a", "secret/b"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results across 2 paths, got %d", len(results))
	}
}
