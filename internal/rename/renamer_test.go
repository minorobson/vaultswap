package rename_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/yourusername/vaultswap/internal/rename"
	"github.com/yourusername/vaultswap/internal/vault"
)

func newRenameServer(t *testing.T) (*httptest.Server, map[string]map[string]interface{}) {
	t.Helper()
	store := map[string]map[string]interface{}{
		"secret/data/src": {"foo": "bar"},
	}
	deleted := map[string]bool{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")
		switch r.Method {
		case http.MethodGet:
			if data, ok := store[path]; ok {
				json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": data}})
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"]; ok {
				store[path] = d.(map[string]interface{})
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			deleted[path] = true
			delete(store, path)
			w.WriteHeader(http.StatusNoContent)
		}
		_ = deleted
	}))
	return srv, store
}

func newRenameClient(t *testing.T, addr string) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(vault.Config{Address: addr, Token: "test"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestRename_Success(t *testing.T) {
	srv, store := newRenameServer(t)
	defer srv.Close()

	client := newRenameClient(t, srv.URL)
	r := rename.New(client, false)
	res := r.RenamePath(context.Background(), "secret/src", "secret/dst")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if _, ok := store["secret/data/dst"]; !ok {
		t.Error("expected dst to be written")
	}
	if _, ok := store["secret/data/src"]; ok {
		t.Error("expected src to be deleted")
	}
}

func TestRename_DryRun_DoesNotWrite(t *testing.T) {
	srv, store := newRenameServer(t)
	defer srv.Close()

	client := newRenameClient(t, srv.URL)
	r := rename.New(client, true)
	res := r.RenamePath(context.Background(), "secret/src", "secret/dst")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
	if _, ok := store["secret/data/dst"]; ok {
		t.Error("dry-run should not write dst")
	}
	if _, ok := store["secret/data/src"]; !ok {
		t.Error("dry-run should not delete src")
	}
}

func TestFprintResults_Labels(t *testing.T) {
	results := []rename.Result{
		{Src: "a", Dst: "b", DryRun: true},
		{Src: "c", Dst: "d"},
		{Src: "e", Dst: "f", Err: fmt.Errorf("boom")},
	}
	var buf bytes.Buffer
	rename.FprintResults(&buf, results)
	out := buf.String()

	if !strings.Contains(out, "[dry-run]") {
		t.Error("expected dry-run label")
	}
	if !strings.Contains(out, "[renamed]") {
		t.Error("expected renamed label")
	}
	if !strings.Contains(out, "[error]") {
		t.Error("expected error label")
	}
}

var _ = vaultapi.DefaultConfig // ensure vault import used
