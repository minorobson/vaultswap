package rename_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/vaultswap/internal/rename"
	"github.com/yourusername/vaultswap/internal/vault"
)

func newRenameServer(t *testing.T) (*httptest.Server, map[string]map[string]interface{}) {
	t.Helper()
	store := map[string]map[string]interface{}{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")

		switch r.Method {
		case http.MethodGet:
			data, ok := store[path]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"data": data},
			})
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if d, ok := body["data"]; ok {
				if dm, ok := d.(map[string]interface{}); ok {
					store[path] = dm
				}
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			delete(store, path)
			w.WriteHeader(http.StatusNoContent)
		}
	}))

	t.Cleanup(server.Close)
	return server, store
}

func newRenameClient(t *testing.T, addr string) *vaultapi.Client {
	t.Helper()
	c, err := vault.NewClient(vault.Config{
		Address: addr,
		Token:   "test-token",
	})
	require.NoError(t, err)
	return c
}

func TestRename_Success(t *testing.T) {
	server, store := newRenameServer(t)
	client := newRenameClient(t, server.URL)

	// Pre-populate source path
	store["secret/data/old-key"] = map[string]interface{}{"username": "admin", "password": "secret"}

	r := rename.New(client)
	result, err := r.Rename("secret", "old-key", "new-key", false)
	require.NoError(t, err)

	assert.Equal(t, "old-key", result.OldPath)
	assert.Equal(t, "new-key", result.NewPath)
	assert.False(t, result.DryRun)
	assert.NoError(t, result.Err)

	// Destination should now exist
	_, destExists := store["secret/data/new-key"]
	assert.True(t, destExists, "destination path should exist after rename")

	// Source should be removed
	_, srcExists := store["secret/data/old-key"]
	assert.False(t, srcExists, "source path should be deleted after rename")
}

func TestRename_DryRun_DoesNotWrite(t *testing.T) {
	server, store := newRenameServer(t)
	client := newRenameClient(t, server.URL)

	store["secret/data/old-key"] = map[string]interface{}{"api_key": "abc123"}

	r := rename.New(client)
	result, err := r.Rename("secret", "old-key", "new-key", true)
	require.NoError(t, err)

	assert.True(t, result.DryRun)
	assert.Equal(t, "old-key", result.OldPath)
	assert.Equal(t, "new-key", result.NewPath)

	// Nothing should have changed
	_, srcExists := store["secret/data/old-key"]
	assert.True(t, srcExists, "source should still exist in dry-run mode")

	_, destExists := store["secret/data/new-key"]
	assert.False(t, destExists, "destination should not be created in dry-run mode")
}

func TestFprintResults_Labels(t *testing.T) {
	var sb strings.Builder

	results := []rename.Result{
		{OldPath: "secrets/alpha", NewPath: "secrets/beta", DryRun: false, Err: nil},
		{OldPath: "secrets/gamma", NewPath: "secrets/delta", DryRun: true, Err: nil},
		{OldPath: "secrets/bad", NewPath: "secrets/worse", DryRun: false, Err: assert.AnError},
	}

	rename.FprintResults(&sb, results)
	out := sb.String()

	assert.Contains(t, out, "renamed")
	assert.Contains(t, out, "dry-run")
	assert.Contains(t, out, "error")
	assert.Contains(t, out, "secrets/alpha")
	assert.Contains(t, out, "secrets/beta")
}
