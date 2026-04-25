package namespace_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/example/vaultswap/internal/namespace"
)

func newTestServer(t *testing.T, namespaces []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys := make([]interface{}, len(namespaces))
		for i, ns := range namespaces {
			keys[i] = ns + "/"
		}
		resp := map[string]interface{}{
			"data": map[string]interface{}{"keys": keys},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func newClient(t *testing.T, addr string) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	client, err := vaultapi.NewClient(cfg)
	require.NoError(t, err)
	client.SetToken("test-token")
	return client
}

func TestList_ReturnsNamespaces(t *testing.T) {
	expected := []string{"team-a", "team-b", "team-c"}
	srv := newTestServer(t, expected)
	defer srv.Close()

	client := newClient(t, srv.URL)
	lister := namespace.New(client)

	result, err := lister.List(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestList_EmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
	}))
	defer srv.Close()

	client := newClient(t, srv.URL)
	lister := namespace.New(client)

	result, err := lister.List(context.Background(), "")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestFilter_BySubstring(t *testing.T) {
	namespaces := []string{"team-a", "team-b", "infra", "infra-prod"}

	result := namespace.Filter(namespaces, "infra")
	assert.Equal(t, []string{"infra", "infra-prod"}, result)
}

func TestFilter_EmptySubstring_ReturnsAll(t *testing.T) {
	namespaces := []string{"team-a", "team-b"}

	result := namespace.Filter(namespaces, "")
	assert.Equal(t, namespaces, result)
}
