package flatten_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/seatgeek/vaultswap/internal/flatten"
)

func newVaultServer(t *testing.T, secrets map[string]map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/secret/data/")
		if r.Method == http.MethodGet {
			data, ok := secrets[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": data},
			})
			return
		}
		if r.Method == http.MethodPut || r.Method == http.MethodPost {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				secrets[path] = d
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

func newClient(t *testing.T, srv *httptest.Server) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("vault client: %v", err)
	}
	c.SetToken("test-token")
	return c
}

type stubClient struct {
	data   map[string]interface{}
	write  map[string]interface{}
	readerr error
	writeerr error
}

func (s *stubClient) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	return s.data, s.readErr
}
func (s *stubClient) WriteSecret(_ context.Context, _ string, data map[string]interface{}) error {
	s.write = data
	return s.writeErr
}

func TestFlattenPath_AlreadyFlat_Skipped(t *testing.T) {
	stub := &stubClient{data: map[string]interface{}{"key": "value"}}
	f := flatten.New(stub, false)
	res := f.FlattenPath(context.Background(), "secret/flat")
	if !res.Skipped {
		t.Fatal("expected skipped result for already-flat secret")
	}
}

func TestFlattenPath_Nested_WritesFlat(t *testing.T) {
	stub := &stubClient{
		data: map[string]interface{}{
			"db": map[string]interface{}{
				"host": "localhost",
				"port": "5432",
			},
		},
	}
	f := flatten.New(stub, false)
	res := f.FlattenPath(context.Background(), "secret/nested")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Skipped {
		t.Fatal("expected non-skipped result")
	}
	if _, ok := stub.write["db.host"]; !ok {
		t.Error("expected key 'db.host' in written data")
	}
	if _, ok := stub.write["db.port"]; !ok {
		t.Error("expected key 'db.port' in written data")
	}
}

func TestFlattenPath_DryRun_DoesNotWrite(t *testing.T) {
	stub := &stubClient{
		data: map[string]interface{}{
			"cfg": map[string]interface{}{"a": "1"},
		},
	}
	f := flatten.New(stub, true)
	res := f.FlattenPath(context.Background(), "secret/dryrun")
	if !res.DryRun {
		t.Fatal("expected dry-run flag")
	}
	if stub.write != nil {
		t.Fatal("expected no write in dry-run mode")
	}
}

func TestFlattenPath_ReadError(t *testing.T) {
	stub := &stubClient{readErr: fmt.Errorf("permission denied")}
	f := flatten.New(stub, false)
	res := f.FlattenPath(context.Background(), "secret/err")
	if res.Err == nil {
		t.Fatal("expected error from read failure")
	}
}

func TestFlattenPaths_ReturnsAllResults(t *testing.T) {
	stub := &stubClient{data: map[string]interface{}{"k": "v"}}
	f := flatten.New(stub, false)
	results := f.FlattenPaths(context.Background(), []string{"a", "b", "c"})
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}
