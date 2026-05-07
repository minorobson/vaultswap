package prefill_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/yourusername/vaultswap/internal/prefill"
)

func newPrefillServer(t *testing.T, initial map[string]interface{}) (*httptest.Server, map[string]interface{}) {
	t.Helper()
	store := initial
	if store == nil {
		store = map[string]interface{}{}
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if len(store) == 0 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"data": store}})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"].(map[string]interface{}); ok {
				for k, v := range d {
					store[k] = v
				}
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
	return srv, store
}

func newPrefillClient(t *testing.T, addr string) *vaultapi.Client {
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

// stubClient satisfies prefill.VaultClient for unit tests.
type stubClient struct {
	data   map[string]interface{}
	readErr error
	writeErr error
	written map[string]interface{}
}

func (s *stubClient) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	return s.data, s.readErr
}
func (s *stubClient) WriteSecret(_ context.Context, _ string, data map[string]interface{}) error {
	s.written = data
	return s.writeErr
}

func TestPrefillPath_SkipsExistingKeys(t *testing.T) {
	client := &stubClient{data: map[string]interface{}{"foo": "bar"}}
	p := prefill.New(client, false)
	results := p.PrefillPath(context.Background(), "secret/test", map[string]string{"foo": "NEW"})
	if len(results) != 1 || !results[0].Skipped {
		t.Fatalf("expected key to be skipped, got %+v", results)
	}
	if client.written != nil {
		t.Fatal("expected no write for skipped key")
	}
}

func TestPrefillPath_WritesMissingKeys(t *testing.T) {
	client := &stubClient{data: map[string]interface{}{}}
	p := prefill.New(client, false)
	results := p.PrefillPath(context.Background(), "secret/test", map[string]string{"newkey": "defaultval"})
	if len(results) != 1 || results[0].Skipped || results[0].Err != nil {
		t.Fatalf("unexpected result: %+v", results)
	}
	if client.written["newkey"] != "defaultval" {
		t.Fatalf("expected written value, got %v", client.written)
	}
}

func TestPrefillPath_DryRun_DoesNotWrite(t *testing.T) {
	client := &stubClient{data: map[string]interface{}{}}
	p := prefill.New(client, true)
	results := p.PrefillPath(context.Background(), "secret/test", map[string]string{"k": "v"})
	if len(results) != 1 || results[0].Skipped || !results[0].DryRun {
		t.Fatalf("unexpected result: %+v", results)
	}
	if client.written != nil {
		t.Fatal("dry-run must not write")
	}
}

func TestFprintResults_Labels(t *testing.T) {
	results := []prefill.Result{
		{Path: "p", Key: "a", Skipped: true},
		{Path: "p", Key: "b", DryRun: true},
		{Path: "p", Key: "c"},
	}
	var buf bytes.Buffer
	prefill.FprintResults(&buf, results)
	out := buf.String()
	for _, want := range []string{"SKIPPED", "DRY-RUN", "PREFILLED"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}
