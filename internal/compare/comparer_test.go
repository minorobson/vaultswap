package compare_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vaultswap/internal/compare"
	"github.com/vaultswap/internal/vault"
)

func newVaultServer(t *testing.T, data map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": data},
		})
	}))
}

func newClient(t *testing.T, addr string) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(addr, "test-token", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestComparePath_NoDiff(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	srv := newVaultServer(t, data)
	defer srv.Close()

	src := newClient(t, srv.URL)
	dst := newClient(t, srv.URL)
	cmp := compare.New(src, dst, false)

	res, err := cmp.ComparePath("secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.HasDiff {
		t.Errorf("expected no diff, got %v", res.Diff)
	}
}

func TestComparePath_WithDiff(t *testing.T) {
	srcSrv := newVaultServer(t, map[string]interface{}{"key": "value1"})
	defer srcSrv.Close()
	dstSrv := newVaultServer(t, map[string]interface{}{"key": "value2"})
	defer dstSrv.Close()

	src := newClient(t, srcSrv.URL)
	dst := newClient(t, dstSrv.URL)
	cmp := compare.New(src, dst, false)

	res, err := cmp.ComparePath("secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.HasDiff {
		t.Error("expected diff between source and destination, got none")
	}
}

func TestComparePaths_ReturnsSingleResult(t *testing.T) {
	srv := newVaultServer(t, map[string]interface{}{"a": "1"})
	defer srv.Close()

	src := newClient(t, srv.URL)
	dst := newClient(t, srv.URL)
	cmp := compare.New(src, dst, false)

	results, err := cmp.ComparePaths([]string{"secret/foo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Path != "secret/foo" {
		t.Errorf("unexpected path: %s", results[0].Path)
	}
}

func TestComparePath_UnreachableSrc(t *testing.T) {
	srv := newVaultServer(t, map[string]interface{}{})
	defer srv.Close()

	src := newClient(t, "http://127.0.0.1:19999")
	dst := newClient(t, srv.URL)
	cmp := compare.New(src, dst, false)

	_, err := cmp.ComparePath("secret/foo")
	if err == nil {
		t.Fatal("expected error for unreachable source")
	}
	if !strings.Contains(err.Error(), "read source") {
		t.Errorf("unexpected error message: %v", err)
	}
}
