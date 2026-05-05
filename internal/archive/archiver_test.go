package archive_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/nicholasgasior/vaultswap/internal/archive"
	"github.com/nicholasgasior/vaultswap/internal/vault"
)

func newVaultServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// metadata endpoint returns version 3
	mux.HandleFunc("/v1/secret/metadata/app/db", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"current_version":3}}`))
	})

	// data read endpoint
	mux.HandleFunc("/v1/secret/data/app/db", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"data":{"password":"s3cr3t"}}}`))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// catch-all write
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	return httptest.NewServer(mux)
}

func newClient(t *testing.T, addr string) *vaultapi.Client {
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

type stubClient struct {
	readData    map[string]any
	readVersion int
	readErr     error
	writeErr    error
	written     map[string]map[string]any
}

func (s *stubClient) ReadSecret(_ context.Context, _ string) (map[string]any, error) {
	return s.readData, s.readErr
}
func (s *stubClient) ReadSecretVersion(_ context.Context, _ string) (int, error) {
	return s.readVersion, s.readErr
}
func (s *stubClient) WriteSecret(_ context.Context, path string, data map[string]any) error {
	if s.written == nil {
		s.written = make(map[string]map[string]any)
	}
	s.written[path] = data
	return s.writeErr
}

func TestArchivePath_Success(t *testing.T) {
	stub := &stubClient{readData: map[string]any{"key": "val"}, readVersion: 5}
	a := archive.New(stub, false)
	r := a.ArchivePath(context.Background(), "app/db", "archive")
	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
	if r.Version != 5 {
		t.Errorf("expected version 5, got %d", r.Version)
	}
	if len(stub.written) == 0 {
		t.Error("expected write to occur")
	}
}

func TestArchivePath_DryRun_DoesNotWrite(t *testing.T) {
	stub := &stubClient{readData: map[string]any{"key": "val"}, readVersion: 2}
	a := archive.New(stub, true)
	r := a.ArchivePath(context.Background(), "app/db", "archive")
	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
	if !r.DryRun {
		t.Error("expected DryRun flag")
	}
	if len(stub.written) != 0 {
		t.Error("expected no writes in dry-run mode")
	}
}

func TestArchivePath_ReadError(t *testing.T) {
	stub := &stubClient{readErr: errors.New("permission denied")}
	a := archive.New(stub, false)
	r := a.ArchivePath(context.Background(), "app/db", "archive")
	if r.Err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArchivePaths_ReturnsAllResults(t *testing.T) {
	stub := &stubClient{readData: map[string]any{"x": "y"}, readVersion: 1}
	a := archive.New(stub, false)
	results := a.ArchivePaths(context.Background(), []string{"a", "b", "c"}, "arch")
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
	_ = vault.NewClient // suppress unused import
}
