package pin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/vault/api"
)

func newPinServer(t *testing.T, versions map[string]interface{}, expectWrite bool) (*httptest.Server, *bool) {
	t.Helper()
	written := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"versions": versions},
			})
		case http.MethodPost, http.MethodPut:
			written = true
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &written
}

func newPinClient(t *testing.T, addr string) *api.Client {
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

func TestPinPath_Success(t *testing.T) {
	versions := map[string]interface{}{"3": map[string]interface{}{}}
	srv, written := newPinServer(t, versions, true)
	client := newPinClient(t, srv.URL)
	p := New(client, false)

	res := p.PinPath(context.Background(), "secret", "myapp/config", 3)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !*written {
		t.Error("expected metadata write, got none")
	}
	if res.Version != 3 {
		t.Errorf("expected version 3, got %d", res.Version)
	}
}

func TestPinPath_DryRun_DoesNotWrite(t *testing.T) {
	versions := map[string]interface{}{"2": map[string]interface{}{}}
	srv, written := newPinServer(t, versions, false)
	client := newPinClient(t, srv.URL)
	p := New(client, true)

	res := p.PinPath(context.Background(), "secret", "myapp/config", 2)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if *written {
		t.Error("dry-run should not write")
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
}

func TestPinPath_VersionNotFound(t *testing.T) {
	versions := map[string]interface{}{"1": map[string]interface{}{}}
	srv, _ := newPinServer(t, versions, false)
	client := newPinClient(t, srv.URL)
	p := New(client, false)

	res := p.PinPath(context.Background(), "secret", "myapp/config", 99)
	if res.Err == nil {
		t.Fatal("expected error for missing version")
	}
}

func TestPinPaths_ReturnsAllResults(t *testing.T) {
	versions := map[string]interface{}{"1": map[string]interface{}{}}
	srv, _ := newPinServer(t, versions, false)
	client := newPinClient(t, srv.URL)
	p := New(client, true)

	results := p.PinPaths(context.Background(), "secret", []string{"a", "b", "c"}, 1)
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}
