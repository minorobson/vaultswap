package sync_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaultswap/internal/sync"
	"github.com/vaultswap/internal/vault"
)

func newTestVaultServer(t *testing.T, data map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			payload := map[string]interface{}{
				"data": map[string]interface{}{"data": data},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(payload)
		case http.MethodPost, http.MethodPut:
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
}

func newClient(t *testing.T, srv *httptest.Server) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(vault.Config{
		Address: srv.URL,
		Token:   "test-token",
		Mount:   "secret",
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestSync_DryRun_DoesNotWrite(t *testing.T) {
	srcSrv := newTestVaultServer(t, map[string]string{"key": "newval"})
	dstSrv := newTestVaultServer(t, map[string]string{"key": "oldval"})
	defer srcSrv.Close()
	defer dstSrv.Close()

	s := sync.New(newClient(t, srcSrv), newClient(t, dstSrv), sync.Options{DryRun: true})
	res, err := s.Sync(context.Background(), "app/config", "app/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Skipped {
		t.Error("expected changes, got skipped")
	}
	if len(res.Changes) == 0 {
		t.Error("expected at least one change")
	}
}

func TestSync_NoChanges_Skipped(t *testing.T) {
	data := map[string]string{"key": "value"}
	srcSrv := newTestVaultServer(t, data)
	dstSrv := newTestVaultServer(t, data)
	defer srcSrv.Close()
	defer dstSrv.Close()

	s := sync.New(newClient(t, srcSrv), newClient(t, dstSrv), sync.Options{})
	res, err := s.Sync(context.Background(), "app/config", "app/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Skipped {
		t.Error("expected result to be skipped when no changes")
	}
}
