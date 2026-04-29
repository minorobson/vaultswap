package search_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/seatgeek/vaultswap/internal/search"
)

func newSearchServer(t *testing.T, secrets map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{"data": map[string]interface{}{"data": secrets}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(data)
	}))
}

func newSearchClient(t *testing.T, addr string) search.VaultClient {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("vault client: %v", err)
	}
	c.SetToken("test-token")
	return &fakeClient{secrets: map[string]interface{}{"key1": "hello-world", "key2": "other"}}
}

type fakeClient struct {
	secrets map[string]interface{}
}

func (f *fakeClient) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	return f.secrets, nil
}

func (f *fakeClient) ListSecrets(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func TestSearchPath_ValueMatch(t *testing.T) {
	client := &fakeClient{secrets: map[string]interface{}{"apikey": "abc123", "token": "xyz"}}
	s := search.New(client)
	results := s.SearchPath(context.Background(), "secret/app", "abc", false)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Key != "apikey" {
		t.Errorf("expected key 'apikey', got %q", results[0].Key)
	}
}

func TestSearchPath_KeyMatch(t *testing.T) {
	client := &fakeClient{secrets: map[string]interface{}{"db_password": "secret", "api_token": "tok"}}
	s := search.New(client)
	results := s.SearchPath(context.Background(), "secret/app", "password", true)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Key != "db_password" {
		t.Errorf("expected 'db_password', got %q", results[0].Key)
	}
}

func TestSearchPath_NoMatch(t *testing.T) {
	client := &fakeClient{secrets: map[string]interface{}{"foo": "bar"}}
	s := search.New(client)
	results := s.SearchPath(context.Background(), "secret/app", "zzz", false)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchPaths_AggregatesResults(t *testing.T) {
	client := &fakeClient{secrets: map[string]interface{}{"match_key": "value"}}
	s := search.New(client)
	results := s.SearchPaths(context.Background(), []string{"a", "b", "c"}, "match", true)
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}
