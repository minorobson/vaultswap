package watch_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/vaultswap/internal/vault"
	"github.com/your-org/vaultswap/internal/watch"
)

func newWatchServer(t *testing.T, responses []map[string]interface{}) (*httptest.Server, *int32) {
	t.Helper()
	var call int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := int(atomic.AddInt32(&call, 1)) - 1
		if idx >= len(responses) {
			idx = len(responses) - 1
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"data": responses[idx]},
		})
	}))
	t.Cleanup(srv.Close)
	return srv, &call
}

func newWatchClient(t *testing.T, addr string) *vault.Client {
	t.Helper()
	c, err := vault.NewClient(vault.Config{Address: addr, Token: "test"})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestWatch_DetectsChange(t *testing.T) {
	responses := []map[string]interface{}{
		{"api_key": "old"},
		{"api_key": "new"},
	}
	srv, _ := newWatchServer(t, responses)
	client := newWatchClient(t, srv.URL)

	w := watch.New(client, []string{"secret/data/svc"}, 20*time.Millisecond, false)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	ch := w.Watch(ctx)
	var got []watch.Result
	for r := range ch {
		if r.Err == nil && r.Changed {
			got = append(got, r)
		}
	}
	if len(got) == 0 {
		t.Fatal("expected at least one change result, got none")
	}
	if got[0].Path != "secret/data/svc" {
		t.Errorf("path = %q, want %q", got[0].Path, "secret/data/svc")
	}
	if got[0].Diff == "" {
		t.Error("expected non-empty diff output")
	}
}

func TestWatch_NoChange_NotEmitted(t *testing.T) {
	responses := []map[string]interface{}{
		{"api_key": "stable"},
	}
	srv, _ := newWatchServer(t, responses)
	client := newWatchClient(t, srv.URL)

	w := watch.New(client, []string{"secret/data/svc"}, 20*time.Millisecond, false)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	ch := w.Watch(ctx)
	for r := range ch {
		if r.Changed {
			t.Errorf("unexpected change result: %+v", r)
		}
	}
}
