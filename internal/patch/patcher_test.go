package patch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/nicholasgasior/vaultswap/internal/patch"
	"github.com/nicholasgasior/vaultswap/internal/vault"
)

func newPatchServer(t *testing.T, existing map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": existing},
			})
		case http.MethodPost, http.MethodPut:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}

func newPatchClient(t *testing.T, addr string) *vaultapi.Client {
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

func TestPatchPath_UpdatesExistingKey(t *testing.T) {
	srv := newPatchServer(t, map[string]interface{}{"foo": "old", "bar": "keep"})
	defer srv.Close()

	vc := newPatchClient(t, srv.URL)
	client, err := vault.NewClient(vault.Config{Address: srv.URL, Token: "test-token"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_ = vc

	p := patch.New(client, false)
	results, err := p.PatchPath(context.Background(), "secret/data/test", map[string]string{"foo": "new"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].OldValue != "old" || results[0].NewValue != "new" {
		t.Errorf("unexpected values: old=%q new=%q", results[0].OldValue, results[0].NewValue)
	}
}

func TestPatchPath_DryRun_DoesNotWrite(t *testing.T) {
	written := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			written = true
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": map[string]interface{}{"k": "v"}},
		})
	}))
	defer srv.Close()

	client, err := vault.NewClient(vault.Config{Address: srv.URL, Token: "test-token"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	p := patch.New(client, true)
	_, err = p.PatchPath(context.Background(), "secret/data/test", map[string]string{"k": "new"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if written {
		t.Error("expected no write in dry-run mode")
	}
}

func TestFprintResults_DryRunLabel(t *testing.T) {
	results := []patch.Result{
		{Path: "secret/data/x", Key: "pw", OldValue: "a", NewValue: "b", DryRun: true},
	}
	var buf bytes.Buffer
	patch.FprintResults(&buf, results)
	if !bytes.Contains(buf.Bytes(), []byte("dry-run")) {
		t.Errorf("expected dry-run label, got: %s", buf.String())
	}
}
