package policy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newTestVaultServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
}

func newTestClient(t *testing.T, addr string) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	client.SetToken("test-token")
	return client
}

func TestApply_DryRun_DoesNotWrite(t *testing.T) {
	srv := newTestVaultServer(t)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	applier := NewApplier(client, true)

	p := &Policy{
		Name:  "my-policy",
		Rules: []Rule{{Path: "secret/*", Capabilities: []string{"read"}}},
	}

	result, err := applier.Apply(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.DryRun {
		t.Error("expected DryRun to be true")
	}
	if !result.Skipped {
		t.Error("expected Skipped to be true")
	}
}

func TestApply_InvalidPolicy_ReturnsError(t *testing.T) {
	srv := newTestVaultServer(t)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	applier := NewApplier(client, false)

	p := &Policy{Name: "", Rules: []Rule{}}
	_, err := applier.Apply(context.Background(), p)
	if err == nil {
		t.Fatal("expected error for invalid policy")
	}
}
