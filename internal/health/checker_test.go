package health_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/vaultswap/internal/health"
)

func newHealthServer(t *testing.T, code int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Vault-Version", "1.15.0")
		w.WriteHeader(code)
	}))
}

func TestCheck_Healthy(t *testing.T) {
	srv := newHealthServer(t, http.StatusOK)
	defer srv.Close()

	c := health.New(5 * time.Second)
	status := c.Check(context.Background(), srv.URL, "")

	if !status.Healthy {
		t.Errorf("expected healthy, got error: %s", status.Error)
	}
	if status.Version != "1.15.0" {
		t.Errorf("expected version 1.15.0, got %s", status.Version)
	}
}

func TestCheck_Sealed(t *testing.T) {
	srv := newHealthServer(t, 503)
	defer srv.Close()

	c := health.New(5 * time.Second)
	status := c.Check(context.Background(), srv.URL, "")

	if !status.Sealed {
		t.Errorf("expected sealed=true")
	}
	if status.Healthy {
		t.Errorf("expected healthy=false for sealed vault")
	}
}

func TestCheck_NamespaceHeader(t *testing.T) {
	var gotNS string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotNS = r.Header.Get("X-Vault-Namespace")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := health.New(5 * time.Second)
	c.Check(context.Background(), srv.URL, "team-a")

	if gotNS != "team-a" {
		t.Errorf("expected namespace header 'team-a', got %q", gotNS)
	}
}

func TestCheck_Unreachable(t *testing.T) {
	c := health.New(500 * time.Millisecond)
	status := c.Check(context.Background(), "http://127.0.0.1:19999", "")

	if status.Healthy {
		t.Errorf("expected unhealthy for unreachable server")
	}
	if status.Error == "" {
		t.Errorf("expected non-empty error")
	}
}

func TestCheckMany_ReturnsAll(t *testing.T) {
	srv1 := newHealthServer(t, http.StatusOK)
	defer srv1.Close()
	srv2 := newHealthServer(t, 503)
	defer srv2.Close()

	targets := []health.Target{
		{Address: srv1.URL, Namespace: "ns1"},
		{Address: srv2.URL, Namespace: "ns2"},
	}

	c := health.New(5 * time.Second)
	results := c.CheckMany(context.Background(), targets)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Healthy {
		t.Errorf("expected first result healthy")
	}
	if !results[1].Sealed {
		t.Errorf("expected second result sealed")
	}
}
