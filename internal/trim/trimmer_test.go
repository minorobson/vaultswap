package trim_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/vaultswap/internal/trim"
	"github.com/your-org/vaultswap/internal/vault"
)

func newTrimServer(t *testing.T, secret map[string]interface{}) *httptest.Server {
	t.Helper()
	data := map[string]interface{}{"data": map[string]interface{}{"data": secret}}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(data)
		case http.MethodPut, http.MethodPost:
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
}

func newTrimClient(t *testing.T, addr string) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(vault.Config{Address: addr, Token: "test-token"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestTrimKey_RemovesKey(t *testing.T) {
	srv := newTrimServer(t, map[string]interface{}{"foo": "bar", "baz": "qux"})
	defer srv.Close()

	client := newTrimClient(t, srv.URL)
	trimmer := trim.New(client, false)

	result := trimmer.TrimKey(context.Background(), "secret/data/mypath", "foo")
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if !result.Removed {
		t.Error("expected Removed=true")
	}
}

func TestTrimKey_DryRun_DoesNotWrite(t *testing.T) {
	written := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			written = true
		}
		data := map[string]interface{}{"data": map[string]interface{}{"data": map[string]interface{}{"key": "val"}}}
		_ = json.NewEncoder(w).Encode(data)
	}))
	defer srv.Close()

	client := newTrimClient(t, srv.URL)
	trimmer := trim.New(client, true)

	result := trimmer.TrimKey(context.Background(), "secret/data/mypath", "key")
	if written {
		t.Error("expected no write in dry-run mode")
	}
	if !result.DryRun {
		t.Error("expected DryRun=true")
	}
	if !result.Removed {
		t.Error("expected Removed=true in dry-run")
	}
}

func TestFprintResults_Labels(t *testing.T) {
	results := []trim.Result{
		{Path: "p", Key: "k1", Removed: true, DryRun: false},
		{Path: "p", Key: "k2", Removed: false},
		{Path: "p", Key: "k3", Removed: true, DryRun: true},
	}
	var buf bytes.Buffer
	trim.FprintResults(&buf, results)
	out := buf.String()

	if !strings.Contains(out, "removed") {
		t.Error("expected 'removed' label")
	}
	if !strings.Contains(out, "skipped") {
		t.Error("expected 'skipped' label")
	}
	if !strings.Contains(out, "dry-run") {
		t.Error("expected 'dry-run' label")
	}
}
